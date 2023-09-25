package engine

import (
	"context"
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
	"github.com/herohde/morlock/pkg/search/searchctl"
	"github.com/seekerror/build"
	"github.com/seekerror/logw"
	"github.com/seekerror/stdlib/pkg/lang"
	"sync"
)

var version = build.NewVersion(0, 89, 3)

// Options are search creation options.
type Options struct {
	// Depth is the search depth limit. If zero, there is no limit. Overridden by search
	// options if provided.
	Depth uint
	// Hash is the transposition table size in MB. If zero, the engine will not use
	// a transposition table.
	Hash uint
	// Noise adds some millipawn randomness to the leaf evaluations.
	Noise uint
}

func (o Options) String() string {
	return fmt.Sprintf("{depth=%v, hash=%v, noise=%v}", o.Depth, o.Hash, o.Noise)
}

// Engine encapsulates game-playing logic, search and evaluation.
type Engine struct {
	name, author string

	launcher searchctl.Launcher
	factory  search.TranspositionTableFactory
	zt       *board.ZobristTable
	seed     int64
	opts     Options

	b      *board.Board
	tt     search.TranspositionTable
	noise  eval.Random
	active searchctl.Handle
	mu     sync.Mutex
}

// Option is an engine creation option.
type Option func(*Engine)

// WithTable configures the engine to use the given transposition table factory.
func WithTable(factory search.TranspositionTableFactory) Option {
	return func(e *Engine) {
		e.factory = factory
	}
}

// WithOptions sets default runtime options.
func WithOptions(opts Options) Option {
	return func(e *Engine) {
		e.opts = opts
	}
}

// WithZobrist configures the engine to use the given random seed instead of the
// default seed of zero.
func WithZobrist(seed int64) Option {
	return func(e *Engine) {
		e.seed = seed
	}
}

func New(ctx context.Context, name, author string, root search.Search, opts ...Option) *Engine {
	e := &Engine{
		name:     name,
		author:   author,
		launcher: &searchctl.Iterative{Root: root},
		factory:  search.NewTranspositionTable,
	}
	for _, fn := range opts {
		fn(e)
	}
	e.zt = board.NewZobristTable(e.seed)

	_ = e.Reset(ctx, fen.Initial)

	logw.Infof(ctx, "Initialized engine: %v, options=%v", e.Name(), e.opts)
	return e
}

// Name returns the engine name and version.
func (e *Engine) Name() string {
	return fmt.Sprintf("%v %v", e.name, version)
}

// Author returns the author.
func (e *Engine) Author() string {
	return e.author
}

func (e *Engine) Options() Options {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.opts
}

func (e *Engine) SetDepth(depth uint) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.opts.Depth = depth
}

func (e *Engine) SetHash(size uint) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.opts.Hash = size
}

func (e *Engine) SetNoise(millipawns uint) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.opts.Noise = millipawns
}

// Board returns a forked board.
func (e *Engine) Board() *board.Board {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.b.Fork()
}

// Position returns the current position in FEN format. Convenience function.
func (e *Engine) Position() string {
	e.mu.Lock()
	defer e.mu.Unlock()

	return fen.Encode(e.b.Position(), e.b.Turn(), e.b.NoProgress(), e.b.FullMoves())
}

// Reset resets the engine to a new starting position in FEN format.
func (e *Engine) Reset(ctx context.Context, position string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	logw.Infof(ctx, "Reset %v, depth=%v, TT=%vMB, noise=%vcp", position, e.opts.Depth, e.opts.Hash, e.opts.Noise/10)

	_, _ = e.haltSearchIfActive(ctx)

	pos, turn, noprogress, fullmoves, err := fen.Decode(position)
	if err != nil {
		return err
	}
	e.b = board.NewBoard(e.zt, pos, turn, noprogress, fullmoves)

	e.tt = search.NoTranspositionTable{}
	if e.opts.Hash > 0 {
		e.tt = e.factory(ctx, uint64(e.opts.Hash)<<20)
	}
	e.noise = eval.Random{}
	if e.opts.Noise > 0 {
		e.noise = eval.NewRandom(int(e.opts.Noise), e.seed)
	}

	logw.Infof(ctx, "New board: %v", e.b)
	return nil
}

// Move selects the given move, usually an opponent move.
func (e *Engine) Move(ctx context.Context, move string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	logw.Infof(ctx, "Move %v", move)

	candidate, err := board.ParseMove(move)
	if err != nil {
		return fmt.Errorf("invalid move: %v", err)
	}

	_, _ = e.haltSearchIfActive(ctx)

	moves := e.b.Position().PseudoLegalMoves(e.b.Turn())
	for _, m := range moves {
		if !candidate.Equals(m) {
			continue
		}

		// Candidate is at least pseudo-legal.

		if !e.b.PushMove(m) {
			return fmt.Errorf("illegal move: %v", m)
		}

		logw.Infof(ctx, "Move %v: %v", m, e.b)
		return nil
	}
	return fmt.Errorf("invalid move: %v", candidate)
}

// TakeBack undoes the latest move.
func (e *Engine) TakeBack(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, _ = e.haltSearchIfActive(ctx)

	m, ok := e.b.PopMove()
	if !ok {
		return fmt.Errorf("no move to take back")
	}

	logw.Infof(ctx, "Takeback %v", m)
	return nil
}

// Analyze analyzes the current position.
func (e *Engine) Analyze(ctx context.Context, opt searchctl.Options) (<-chan search.PV, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := opt.DepthLimit.V(); !ok {
		opt.DepthLimit = lang.Some(e.opts.Depth)
	}

	logw.Infof(ctx, "Analyze %v, opt=%v", e.b, opt)

	if e.active != nil {
		return nil, fmt.Errorf("search already active")
	}

	handle, out := e.launcher.Launch(ctx, e.b.Fork(), e.tt, e.noise, opt)
	e.active = handle
	return out, nil
}

// Halt halts the active search and returns the principal variation, if any.
func (e *Engine) Halt(ctx context.Context) (search.PV, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	logw.Infof(ctx, "Halt")

	pv, ok := e.haltSearchIfActive(ctx)
	if !ok {
		return search.PV{}, fmt.Errorf("no active search")
	}
	return pv, nil
}

func (e *Engine) haltSearchIfActive(ctx context.Context) (search.PV, bool) {
	if e.active != nil {
		pv := e.active.Halt()
		logw.Infof(ctx, "Search %v halted: %v", e.b, pv)

		e.active = nil
		return pv, true
	}
	return search.PV{}, false
}

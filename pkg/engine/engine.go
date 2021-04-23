package engine

import (
	"context"
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/search"
	"github.com/seekerror/build"
	"github.com/seekerror/logw"
	"sync"
)

var version = build.NewVersion(0, 89, 0)

// Engine encapsulates game-playing logic, search and evaluation.
type Engine struct {
	name, author string

	launcher search.Launcher
	zt       *board.ZobristTable
	opts     options

	b      *board.Board
	tt     search.TranspositionTable
	active search.Handle
	mu     sync.Mutex
}

// Option is an engine option.
type Option func(*options)

type options struct {
	depth   *int
	factory search.TranspositionTableFactory
	seed    int64
}

func WithDepthLimit(depth int) Option {
	return func(o *options) {
		o.depth = &depth
	}
}

// WithTable configures the engine to use the given transposition table factory.
func WithTable(factory search.TranspositionTableFactory) Option {
	return func(o *options) {
		o.factory = factory
	}
}

// WithZobrist configures the engine to use the given random seed instead of the
// default seed of zero.
func WithZobrist(seed int64) Option {
	return func(o *options) {
		o.seed = seed
	}
}

func New(ctx context.Context, name, author string, root search.Search, opts ...Option) *Engine {
	e := &Engine{
		name:     name,
		author:   author,
		launcher: &search.Iterative{Root: root},
		opts: options{
			factory: search.NewTranspositionTable,
			seed:    0,
		},
	}
	for _, fn := range opts {
		fn(&e.opts)
	}
	e.zt = board.NewZobristTable(e.opts.seed)

	_ = e.Reset(ctx, fen.Initial, 0)

	logw.Infof(ctx, "Initialized engine: %v", e.Name())
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

// Position returns the current position in FEN format.
func (e *Engine) Position() string {
	e.mu.Lock()
	defer e.mu.Unlock()

	return fen.Encode(e.b.Position(), e.b.Turn(), e.b.NoProgress(), e.b.FullMoves())
}

// Reset resets the engine to a new starting position in FEN format.
func (e *Engine) Reset(ctx context.Context, position string, hash uint64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	logw.Infof(ctx, "Reset %v, TT=%vMB", position, hash>>20)

	_, _ = e.haltSearchIfActive(ctx)

	pos, turn, noprogress, fullmoves, err := fen.Decode(position)
	if err != nil {
		return err
	}
	e.b = board.NewBoard(e.zt, pos, turn, noprogress, fullmoves)

	e.tt = search.NoTranspositionTable{}
	if hash > 0 {
		e.tt = e.opts.factory(ctx, hash)
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

// Analyze analyzes the current position.
func (e *Engine) Analyze(ctx context.Context, opt search.Options) (<-chan search.PV, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if opt.DepthLimit == nil {
		opt.DepthLimit = e.opts.depth
	}

	logw.Infof(ctx, "Analyze %v, opt=%v", e.b, opt)

	if e.active != nil {
		return nil, fmt.Errorf("search already active")
	}

	handle, out := e.launcher.Launch(ctx, e.b.Fork(), e.tt, opt)
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

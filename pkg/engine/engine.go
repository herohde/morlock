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

var version = build.NewVersion(0, 86, 0)

// Engine encapsulates game-playing logic, search and evaluation.
type Engine struct {
	name, author string
	zt           *board.ZobristTable
	launcher     search.Launcher

	b      *board.Board
	active search.Handle
	mu     sync.Mutex
}

func New(ctx context.Context, name, author string, launcher search.Launcher) *Engine {
	e := &Engine{
		name:     name,
		author:   author,
		zt:       board.NewZobristTable(0),
		launcher: launcher,
	}
	_ = e.Reset(ctx, fen.Initial)

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

// Reset resets the engine to a new starting position.
func (e *Engine) Reset(ctx context.Context, position string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	logw.Infof(ctx, "Reset %v", position)

	_, _ = e.haltSearchIfActive(ctx)

	pos, turn, noprogress, fullmoves, err := fen.Decode(position)
	if err != nil {
		return err
	}
	e.b = board.NewBoard(e.zt, pos, turn, noprogress, fullmoves)

	logw.Infof(ctx, "New board: %v", e.b)
	return nil
}

// Move selects the given move, usually an opponent move.
func (e *Engine) Move(ctx context.Context, candidate board.Move) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	logw.Infof(ctx, "Move %v", candidate)

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

	logw.Infof(ctx, "Analyze %v, opt=%v", e.b, opt)

	if e.active != nil {
		return nil, fmt.Errorf("search already active")
	}

	handle, out := e.launcher.Launch(ctx, e.b.Fork(), opt)
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

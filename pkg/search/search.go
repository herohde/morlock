// Package search contains search functionality and utilities.
package search

import (
	"context"
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"time"
)

// PV represents the principal variation for some search depth.
type PV struct {
	Moves []board.Move
	Score board.Score
	Nodes uint64
	Time  time.Duration
}

func (p PV) String() string {
	pv := board.FormatMoves(p.Moves, func(m board.Move) string {
		return m.String()
	})
	return fmt.Sprintf("depth=%v score=%v nodes=%v time=%v pv=%v", len(p.Moves), p.Score, p.Nodes, p.Time, pv)
}

// Options hold dynamic search options. The user may change these on a particular search.
type Options struct {
	DepthLimit int // 0 == no limit
}

// Launcher is a Search generator.
type Launcher interface {
	// Launch a new search from the given position. It expects an exclusive (forked) board and
	// returns a PV channel for iteratively deeper searches. If the search is exhausted, the
	// channel is closed. The search can be stopped at any time.
	Launch(ctx context.Context, b *board.Board, opt Options) (Handle, <-chan PV)
}

// Handle is an interface for the engine to manage searches. The engine is expected to spin off
// searches with forked boards and close/abandon them when no longer needed. This design keeps
// stopping conditions and re-synchronization trivial.
type Handle interface {
	// Halt halts the search, if running. Idempotent.
	Halt() PV
}

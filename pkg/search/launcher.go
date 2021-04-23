// Package search contains search functionality and utilities.
package search

import (
	"context"
	"errors"
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"strings"
)

// ErrHalted is an error indicating that the search was halted.
var ErrHalted = errors.New("search halted")

// Options hold dynamic search options. The user may change these on a particular search.
type Options struct {
	// DepthLimit, if set, limits the search to the given ply depth.
	DepthLimit *int
	// TimeControl, if set, limits the search to the given time parameters.
	TimeControl *TimeControl
}

func (o Options) String() string {
	var ret []string
	if o.DepthLimit != nil {
		ret = append(ret, fmt.Sprintf("depth=%v", *o.DepthLimit))
	}
	if o.TimeControl != nil {
		ret = append(ret, fmt.Sprintf("time=%v", *o.TimeControl))
	}
	return fmt.Sprintf("[%v]", strings.Join(ret, ", "))
}

// Launcher is an interface for managing searches.
type Launcher interface {
	// Launch a new search from the given position. It expects an exclusive (forked) board and
	// returns a PV channel for iteratively deeper searches. If the search is exhausted, the
	// channel is closed. The search can be stopped at any time.
	Launch(ctx context.Context, b *board.Board, tt TranspositionTable, opt Options) (Handle, <-chan PV)
}

// Handle is an interface for the engine to manage searches. The engine is expected to spin off
// searches with forked boards and close/abandon them when no longer needed. This design keeps
// stopping conditions and re-synchronization trivial.
type Handle interface {
	// Halt halts the search, if running. Idempotent.
	Halt() PV
}

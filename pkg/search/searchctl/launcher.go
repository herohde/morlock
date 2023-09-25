// Package searchctl contains search functionality and utilities.
package searchctl

import (
	"context"
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
	"github.com/seekerror/stdlib/pkg/lang"
	"strings"
)

// Options hold dynamic search options. The user may change these on a particular search.
type Options struct {
	// DepthLimit, if set, limits the search to the given ply depth. Zero means no limit.
	DepthLimit lang.Optional[uint]
	// TimeControl, if set, limits the search to the given time parameters.
	TimeControl lang.Optional[TimeControl]
}

func (o Options) String() string {
	var ret []string
	if v, ok := o.DepthLimit.V(); ok {
		ret = append(ret, fmt.Sprintf("depth=%v", v))
	}

	if v, ok := o.TimeControl.V(); ok {
		ret = append(ret, fmt.Sprintf("time=%v", v))
	}
	return fmt.Sprintf("[%v]", strings.Join(ret, ", "))
}

// Launcher is an interface for managing searches.
type Launcher interface {
	// Launch a new search from the given position. It expects an exclusive (forked) board and
	// returns a PV channel for iteratively deeper searches. If the search is exhausted, the
	// channel is closed. The search can be stopped at any time.
	Launch(ctx context.Context, b *board.Board, tt search.TranspositionTable, noise eval.Random, opt Options) (Handle, <-chan search.PV)
}

// Handle is an interface for the engine to manage searches. The engine is expected to spin off
// searches with forked boards and close/abandon them when no longer needed. This design keeps
// stopping conditions and re-synchronization trivial.
type Handle interface {
	// Halt halts the search, if running. Idempotent.
	Halt() search.PV
}

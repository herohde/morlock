package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/seekerror/logw"
	"go.uber.org/atomic"
	"sync"
	"time"
)

// Iterative is a search harness for iterative deepening search.
type Iterative struct {
	Root Search
}

func (i *Iterative) Launch(ctx context.Context, b *board.Board, tt TranspositionTable, opt Options) (Handle, <-chan PV) {
	out := make(chan PV, 1)
	h := &handle{
		init: make(chan struct{}),
		quit: make(chan struct{}),
	}
	go h.process(ctx, i.Root, b, tt, opt, out)

	return h, out
}

type handle struct {
	init, quit        chan struct{}
	initialized, done atomic.Bool

	pv PV
	mu sync.Mutex
}

func (h *handle) process(ctx context.Context, search Search, b *board.Board, tt TranspositionTable, opt Options, out chan PV) {
	defer h.markInitialized()
	defer close(out)

	sctx := &Context{Alpha: eval.NegInfScore, Beta: eval.InfScore, TT: tt}
	soft, useSoft := EnforceTimeControl(ctx, h, opt.TimeControl, b.Turn())

	depth := 1
	for !h.done.Load() {
		start := time.Now()

		nodes, score, moves, err := search.Search(ctx, sctx, b, depth, h.quit)
		if err != nil {
			if err == ErrHalted {
				return // Halt was called.
			}
			logw.Errorf(ctx, "Search failed on %v at depth=%v: %v", b, depth, err)
			return
		}

		pv := PV{
			Depth: depth,
			Nodes: nodes,
			Score: score,
			Moves: moves,
			Time:  time.Since(start),
		}
		if tt != nil {
			pv.Hash = tt.Used()
		}

		logw.Debugf(ctx, "Searched %v: %v", b.Position(), pv)

		h.mu.Lock()
		h.pv = pv
		h.mu.Unlock()

		select {
		case <-out:
		default:
		}
		out <- pv

		h.markInitialized()
		if opt.DepthLimit != nil && depth == *opt.DepthLimit {
			return // halt: reached max depth
		}
		if md, ok := score.MateDistance(); ok && int(md) < depth {
			return // halt: forced mate found within full width search. Exact result.
		}
		if useSoft && soft < time.Since(start) {
			return // halt: exceeded soft time limit. Do not start new search.
		}
		depth++
	}
}

func (h *handle) Halt() PV {
	<-h.init
	if h.done.CAS(false, true) {
		close(h.quit)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	return h.pv
}

func (h *handle) markInitialized() {
	if h.initialized.CAS(false, true) {
		close(h.init)
	}
}

// IsClosed return true iff the quit channel is closed.
func IsClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

package searchctl

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
	"github.com/seekerror/logw"
	"github.com/seekerror/stdlib/pkg/util/contextx"
	"github.com/seekerror/stdlib/pkg/util/iox"
	"sync"
	"time"
)

// Iterative is a search harness for iterative deepening search.
type Iterative struct {
	Root search.Search
}

func (i *Iterative) Launch(ctx context.Context, b *board.Board, tt search.TranspositionTable, opt Options) (Handle, <-chan search.PV) {
	out := make(chan search.PV, 1)
	h := &handle{
		init: iox.NewAsyncCloser(),
		quit: iox.NewAsyncCloser(),
	}
	go h.process(ctx, i.Root, b, tt, opt, out)

	return h, out
}

type handle struct {
	init, quit iox.AsyncCloser

	pv search.PV
	mu sync.Mutex
}

func (h *handle) process(ctx context.Context, root search.Search, b *board.Board, tt search.TranspositionTable, opt Options, out chan search.PV) {
	defer h.init.Close()
	defer close(out)

	sctx := &search.Context{Alpha: eval.NegInfScore, Beta: eval.InfScore, TT: tt}
	soft, useSoft := EnforceTimeControl(ctx, h, opt.TimeControl, b.Turn())

	wctx, cancel := contextx.WithQuitCancel(ctx, h.quit.Closed())
	defer cancel()

	depth := 1
	for !h.quit.IsClosed() {
		start := time.Now()

		nodes, score, moves, err := root.Search(wctx, sctx, b, depth)
		if err != nil {
			if err == search.ErrHalted {
				return // Halt was called.
			}
			logw.Errorf(ctx, "Search failed on %v at depth=%v: %v", b, depth, err)
			return
		}

		pv := search.PV{
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

		h.init.Close()
		if limit, ok := opt.DepthLimit.V(); ok && uint(depth) == limit {
			return // halt: reached max depth
		}
		if md, ok := score.MateDistance(); ok && int(md) <= depth {
			return // halt: forced mate found within full width search. Exact result.
		}
		if useSoft && soft < time.Since(start) {
			return // halt: exceeded soft time limit. Do not start new search.
		}
		depth++
	}
}

func (h *handle) Halt() search.PV {
	<-h.init.Closed()
	h.quit.Close()

	h.mu.Lock()
	defer h.mu.Unlock()

	return h.pv
}

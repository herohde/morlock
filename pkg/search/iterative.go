package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/seekerror/logw"
	"go.uber.org/atomic"
	"sync"
	"time"
)

// Iterative is a search harness for iterative deepening search.
type Iterative struct {
	search     Search
	depthLimit int // 0 if not max
}

func NewIterative(search Search, depthLimit int) Launcher {
	return &Iterative{
		search:     search,
		depthLimit: depthLimit,
	}
}

func (i *Iterative) Launch(ctx context.Context, b *board.Board, opt Options) (Handle, <-chan PV) {
	if i.depthLimit > 0 && opt.DepthLimit == nil {
		opt.DepthLimit = &i.depthLimit
	}

	out := make(chan PV, 1)
	h := &handle{
		init: make(chan struct{}),
		quit: make(chan struct{}),
	}
	go h.process(ctx, i.search, b, opt, out)

	return h, out
}

type handle struct {
	init, quit        chan struct{}
	initialized, done atomic.Bool

	pv PV
	mu sync.Mutex
}

func (h *handle) process(ctx context.Context, search Search, b *board.Board, opt Options, out chan PV) {
	defer h.markInitialized()
	defer close(out)

	var soft, hard time.Duration
	if opt.TimeControl != nil {
		soft, hard = opt.TimeControl.Limits(b.Turn())

		time.AfterFunc(hard, func() {
			h.Halt()
		})

		logw.Debugf(ctx, "Search time limits: [%v; %v]", soft, hard)
	}

	depth := 1
	for !h.done.Load() {
		start := time.Now()

		nodes, score, moves, err := search.Search(ctx, b, depth, h.quit)
		if err != nil {
			if err == ErrHalted {
				return // Halt was called.
			}
			logw.Errorf(ctx, "Search failed on %v at depth=%v: %v", b, depth, err)
			return
		}

		pv := PV{
			Nodes: nodes,
			Score: score,
			Moves: moves,
			Time:  time.Since(start),
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
		if soft > 0 && soft < time.Since(start) {
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

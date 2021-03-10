package search

import (
	"context"
	"errors"
	"github.com/herohde/morlock/pkg/board"
	"github.com/seekerror/logw"
	"go.uber.org/atomic"
	"sync"
	"time"
)

// ErrHalted is an error indicating that the search was halted.
var ErrHalted = errors.New("search halted")

// Searcher implements search of the game tree to a given depth. Thread-safe.
type Searcher interface {
	Search(ctx context.Context, b *board.Board, depth int, quit <-chan struct{}) (uint64, board.Score, []board.Move, error)
}

// Iterative is a search harness for iterative deepening search.
type Iterative struct {
	search Searcher
}

func NewIterative(search Searcher) Launcher {
	return &Iterative{
		search: search,
	}
}

func (i *Iterative) Launch(ctx context.Context, b *board.Board, opt Options) (Handle, <-chan PV) {
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

func (h *handle) process(ctx context.Context, search Searcher, b *board.Board, opt Options, out chan PV) {
	defer h.markInitialized()
	defer close(out)

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
		if depth == opt.DepthLimit {
			return
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

func isClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

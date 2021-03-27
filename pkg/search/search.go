package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// Search implements search of the game tree to a given depth. Thread-safe.
type Search interface {
	Search(ctx context.Context, b *board.Board, depth int, quit <-chan struct{}) (uint64, eval.Score, []board.Move, error)
}

// QuietSearch is a limited quiescence search in a given [alpha;beta] window,
// where standing pat is an option for evaluation purposes. Thread-safe.
type QuietSearch interface {
	QuietSearch(ctx context.Context, b *board.Board, alpha, beta eval.Score, quit <-chan struct{}) (uint64, eval.Score)
}

// ZeroPly is an evaluator wrapped as a QuietSearch.
type ZeroPly struct {
	Eval eval.Evaluator
}

func (z ZeroPly) QuietSearch(ctx context.Context, b *board.Board, alpha, beta eval.Score, quit <-chan struct{}) (uint64, eval.Score) {
	return 1, z.Eval.Evaluate(ctx, b)
}

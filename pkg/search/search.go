package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// Search implements search of the game tree to a given depth. Thread-safe.
type Search interface {
	Search(ctx context.Context, b *board.Board, depth int, quit <-chan struct{}) (uint64, board.Score, []board.Move, error)
}

// Quiescence is a limited quiescence search in a given [alpha;beta] window. Thread-safe.
type Quiescence interface {
	QuietSearch(ctx context.Context, b *board.Board, alpha, beta board.Score, quit <-chan struct{}) (uint64, board.Score)
}

// ZeroPly is an evaluator wrapped as a Quiescence search.
type ZeroPly struct {
	Eval eval.Evaluator
}

func (z ZeroPly) QuietSearch(ctx context.Context, b *board.Board, alpha, beta board.Score, quit <-chan struct{}) (uint64, board.Score) {
	return 0, z.Eval.Evaluate(ctx, b.Position(), b.Turn())
}

package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// Context holds optional context for search implementations.
type Context struct {
	Alpha, Beta eval.Score // Limit search to a [Alpha;Beta] Window

	TT TranspositionTable // HashTable
}

var EmptyContext = &Context{TT: NoTranspositionTable{}}

// Search implements search of the game tree to a given depth. Thread-safe.
type Search interface {
	Search(ctx context.Context, sctx *Context, b *board.Board, depth int, quit <-chan struct{}) (uint64, eval.Score, []board.Move, error)
}

// QuietSearch is a limited quiescence search, where standing pat is an option
// for evaluation purposes. Thread-safe.
type QuietSearch interface {
	QuietSearch(ctx context.Context, sctx *Context, b *board.Board, quit <-chan struct{}) (uint64, eval.Score)
}

// ZeroPly is an evaluator wrapped as a QuietSearch.
type ZeroPly struct {
	Eval eval.Evaluator
}

func (z ZeroPly) QuietSearch(ctx context.Context, sctx *Context, b *board.Board, quit <-chan struct{}) (uint64, eval.Score) {
	return 1, eval.HeuristicScore(z.Eval.Evaluate(ctx, b))
}

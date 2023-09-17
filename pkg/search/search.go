package search

import (
	"context"
	"errors"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// ErrHalted is an error indicating that the search was halted.
var ErrHalted = errors.New("search halted")

// Context holds optional context for search implementations.
type Context struct {
	Alpha, Beta eval.Score   // Limit search to a [Alpha;Beta] Window
	Ponder      []board.Move // Limit search to variation, if present.

	TT    TranspositionTable // HashTable (user configurable)
	Noise eval.Random        // Evaluation noise (user configurable)
}

var EmptyContext = &Context{TT: NoTranspositionTable{}}

// Search implements search of the game tree to a given depth. Context is cancelled if halted. Thread-safe.
type Search interface {
	Search(ctx context.Context, sctx *Context, b *board.Board, depth int) (uint64, eval.Score, []board.Move, error)
}

// QuietSearch is a limited quiescence search, where standing pat is an option for evaluation purposes.
// Context is cancelled if halted. Thread-safe.
type QuietSearch interface {
	QuietSearch(ctx context.Context, sctx *Context, b *board.Board) (uint64, eval.Score)
}

// Evaluator is a static evaluator in a search context.
type Evaluator interface {
	// Evaluate returns the position score in Pawns.
	Evaluate(ctx context.Context, sctx *Context, b *board.Board) eval.Pawns
}

// Leaf is a leaf evaluator in a search context. It implicitly adds user-configurable evaluation noise.
type Leaf struct {
	Eval eval.Evaluator
}

func (s Leaf) Evaluate(ctx context.Context, sctx *Context, b *board.Board) eval.Pawns {
	return s.Eval.Evaluate(ctx, b) + sctx.Noise.Evaluate(ctx, b)
}

func (s Leaf) QuietSearch(ctx context.Context, sctx *Context, b *board.Board) (uint64, eval.Score) {
	return 1, eval.HeuristicScore(s.Evaluate(ctx, sctx, b))
}

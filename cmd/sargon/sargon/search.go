package sargon

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
)

// OnePlyIfChecked implements the SARGON search extension if searching 1 ply deeper if in check.
type OnePlyIfChecked struct {
	Eval eval.Evaluator
}

func (q OnePlyIfChecked) QuietSearch(ctx context.Context, b *board.Board, alpha, beta eval.Score, quit <-chan struct{}) (uint64, eval.Score) {
	if !b.Position().IsChecked(b.Turn()) {
		return 1, eval.HeuristicScore(q.Eval.Evaluate(ctx, b))
	}

	nodes, score, _, _ := search.Minimax{Eval: q.Eval}.Search(ctx, b, 1, quit)
	return nodes, score
}

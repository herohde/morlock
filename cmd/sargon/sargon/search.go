package sargon

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
)

// Hook is a Search wrapper that resets Points.
type Hook struct {
	Eval search.Search
	Hook *Points
}

func (h Hook) Search(ctx context.Context, sctx *search.Context, b *board.Board, depth int, quit <-chan struct{}) (uint64, eval.Score, []board.Move, error) {
	h.Hook.Reset(ctx, b)
	return h.Eval.Search(ctx, sctx, b, depth, quit)
}

// OnePlyIfChecked implements the SARGON search extension if searching 1 ply deeper if in check.
type OnePlyIfChecked struct {
	Eval eval.Evaluator
}

func (q OnePlyIfChecked) QuietSearch(ctx context.Context, sctx *search.Context, b *board.Board, quit <-chan struct{}) (uint64, eval.Score) {
	if !b.Position().IsChecked(b.Turn()) {
		return 1, eval.HeuristicScore(q.Eval.Evaluate(ctx, b))
	}

	s := search.AlphaBeta{
		Eval: search.ZeroPly{Eval: q.Eval},
	}

	nodes, score, _, _ := s.Search(ctx, sctx, b, 1, quit)
	return nodes, score
}

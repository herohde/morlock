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

func (h Hook) Search(ctx context.Context, sctx *search.Context, b *board.Board, depth int) (uint64, eval.Score, []board.Move, error) {
	h.Hook.Reset(ctx, b)
	return h.Eval.Search(ctx, sctx, b, depth)
}

// OnePlyIfChecked implements the SARGON search extension if searching 1 ply deeper if in check.
type OnePlyIfChecked struct {
	Leaf search.Leaf
}

func (q OnePlyIfChecked) QuietSearch(ctx context.Context, sctx *search.Context, b *board.Board) (uint64, eval.Score) {
	if !b.Position().IsChecked(b.Turn()) {
		return 1, eval.HeuristicScore(q.Leaf.Evaluate(ctx, sctx, b))
	}

	s := search.AlphaBeta{
		Eval: q.Leaf,
	}

	nodes, score, _, _ := s.Search(ctx, sctx, b, 1)
	return nodes, score
}

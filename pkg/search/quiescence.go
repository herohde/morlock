package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"sort"
)

// Quiescence implements a configurable alpha-beta QuietSearch.
type Quiescence struct {
	Pick Selection
	Eval eval.Evaluator
}

func (q Quiescence) QuietSearch(ctx context.Context, b *board.Board, alpha, beta eval.Score, quit <-chan struct{}) (uint64, eval.Score) {
	run := &runQuiescence{pick: q.Pick, eval: q.Eval, b: b, quit: quit}
	score := run.search(ctx, alpha, beta)
	return run.nodes, score
}

type runQuiescence struct {
	pick  Selection
	eval  eval.Evaluator
	b     *board.Board
	nodes uint64

	quit <-chan struct{}
}

// search returns the positive score for the color.
func (r *runQuiescence) search(ctx context.Context, alpha, beta eval.Score) eval.Score {
	if IsClosed(r.quit) {
		return eval.ZeroScore
	}
	if r.b.Result().Outcome == board.Draw {
		return eval.ZeroScore
	}

	r.nodes++

	hasLegalMoves := false
	turn := r.b.Turn()
	score := eval.HeuristicScore(r.eval.Evaluate(ctx, r.b))
	alpha = eval.Max(alpha, score)

	// NOTE: Don't cutoff based on evaluation here. See if any legal moves first.
	// Also do not report mate-in-X endings.

	moves := r.b.Position().PseudoLegalMoves(turn)
	sort.Sort(board.ByMVVLVA(moves))

	for _, m := range moves {
		if !r.b.PushMove(m) {
			continue
		}

		if r.pick(ctx, m, r.b) {
			score := r.search(ctx, beta.Negate(), alpha.Negate())
			score = eval.IncrementMateDistance(score).Negate()
			alpha = eval.Max(alpha, score)
		}

		r.b.PopMove()
		hasLegalMoves = true

		if alpha == beta || beta.Less(alpha) {
			break // cutoff
		}
	}

	if !hasLegalMoves {
		if result := r.b.AdjudicateNoLegalMoves(); result.Reason == board.Checkmate {
			return eval.NegInfScore
		}
		return eval.ZeroScore
	}
	return alpha
}

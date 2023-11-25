package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/seekerror/stdlib/pkg/util/contextx"
)

// Quiescence implements a configurable alpha-beta QuietSearch.
type Quiescence struct {
	Explore Exploration
	Eval    Evaluator
}

func (q Quiescence) QuietSearch(ctx context.Context, sctx *Context, b *board.Board) (uint64, eval.Score) {
	run := &runQuiescence{explore: q.Explore, eval: q.Eval, b: b}

	low, high := eval.NegInfScore, eval.InfScore
	if !sctx.Alpha.IsInvalid() {
		low = sctx.Alpha
	}
	if !sctx.Beta.IsInvalid() {
		high = sctx.Beta
	}

	score := run.search(ctx, sctx, low, high)
	return run.nodes, score
}

type runQuiescence struct {
	explore Exploration
	eval    Evaluator
	b       *board.Board
	nodes   uint64
}

// search returns the positive score for the color.
func (r *runQuiescence) search(ctx context.Context, sctx *Context, alpha, beta eval.Score) eval.Score {
	if contextx.IsCancelled(ctx) {
		return eval.ZeroScore
	}
	if r.b.Result().Outcome == board.Draw {
		return eval.ZeroScore
	}

	r.nodes++

	hasLegalMoves := false
	turn := r.b.Turn()
	score := eval.HeuristicScore(r.eval.Evaluate(ctx, sctx, r.b))
	alpha = eval.Max(alpha, score)

	// NOTE: Don't cutoff based on evaluation here. See if any legal moves first.
	// Also do not report mate-in-X endings.

	priority, explore := r.explore(ctx, r.b)

	moves := board.NewMoveList(r.b.Position().PseudoLegalMoves(turn), priority)
	for {
		m, ok := moves.Next()
		if !ok {
			break
		}
		if !r.b.PushMove(m) {
			continue // skip: not legal
		}

		if explore(m) {
			score := r.search(ctx, sctx, beta.Negate(), alpha.Negate())
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

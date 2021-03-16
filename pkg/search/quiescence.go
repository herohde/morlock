package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"sort"
)

// Quiescence implements a configurable alpha-beta QuietSearch.
type Quiescence struct {
	Eval eval.Evaluator
}

func (q Quiescence) QuietSearch(ctx context.Context, b *board.Board, alpha, beta eval.Score, quit <-chan struct{}) (uint64, eval.Score) {
	run := &runQuiescence{eval: q.Eval, b: b, quit: quit}
	score := run.search(ctx, alpha, beta)
	return run.nodes, eval.Unit(b.Turn()) * score
}

type runQuiescence struct {
	eval  eval.Evaluator
	b     *board.Board
	nodes uint64

	quit <-chan struct{}
}

// search returns the positive score for the color.
func (r *runQuiescence) search(ctx context.Context, alpha, beta eval.Score) eval.Score {
	if IsClosed(r.quit) {
		return 0
	}
	if r.b.Result().Outcome == board.Draw {
		return 0
	}

	r.nodes++

	hasLegalMoves := false
	turn := r.b.Turn()
	score := eval.Unit(turn) * r.eval.Evaluate(ctx, r.b.Position(), turn)
	alpha = eval.Max(alpha, score)

	// NOTE: Don't cutoff based on evaluation here. See if any legal moves first.

	moves := r.b.Position().PseudoLegalMoves(turn)
	sort.Sort(board.ByMVVLVA(moves))

	for _, m := range moves {
		if !r.b.PushMove(m) {
			continue
		}

		explore := m.IsPromotion()
		if m.IsCapture() {
			if eval.NominalValue(m.Piece) < eval.NominalValue(m.Capture) {
				explore = true
			}
			if !r.b.Position().IsAttacked(turn, m.To) {
				explore = true
			}
		}

		if explore {
			score := r.search(ctx, -beta, -alpha)
			alpha = eval.Max(alpha, -score)
		}

		r.b.PopMove()
		hasLegalMoves = true

		if alpha >= beta {
			break // cutoff
		}
	}

	if !hasLegalMoves {
		if result := r.b.AdjudicateNoLegalMoves(); result.Reason == board.Checkmate {
			return eval.MinScore
		}
		return 0
	}
	return alpha
}

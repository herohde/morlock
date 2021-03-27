package turochamp

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
	"sort"
)

// Quiescence implements the selective "considerable moves" search:
//   (1) Re-captures are considerable.
//   (2) Capture of en prise pieces are considerable.
//   (3) Capture of higher value pieces are considerable.
//   (4) Checkmate are considerable.
// Additionally, it adds the "has already castled" bonus to the evaluator.
type Quiescence struct {
	Eval eval.Evaluator
}

func (q Quiescence) QuietSearch(ctx context.Context, b *board.Board, alpha, beta eval.Score, quit <-chan struct{}) (uint64, eval.Score) {
	run := &runQuiescence{eval: q.Eval, b: b, quit: quit}
	score := run.search(ctx, alpha, beta)
	return run.nodes, score
}

type runQuiescence struct {
	eval  eval.Evaluator
	b     *board.Board
	nodes uint64

	quit <-chan struct{}
}

// search returns the positive score for the color.
func (r *runQuiescence) search(ctx context.Context, alpha, beta eval.Score) eval.Score {
	if search.IsClosed(r.quit) {
		return eval.ZeroScore
	}
	if r.b.Result().Outcome == board.Draw {
		return eval.ZeroScore
	}

	r.nodes++

	hasLegalMoves := false
	turn := r.b.Turn()
	score := evaluate(ctx, r.b, r.eval)
	alpha = eval.Max(alpha, score)

	mayRecapture := false
	var target board.Square
	if m, ok := r.b.LastMove(); ok && m.IsCapture() {
		mayRecapture = true
		target = m.To
	}

	// lastmove, _ := r.b.LastMove();
	// logw.Debugf(ctx, "SCORE: %v (last: %v, recapture=%v)= %v", r.b.Position(), lastmove, mayRecapture, score)

	moves := r.b.Position().PseudoLegalMoves(turn)
	sort.Sort(board.ByMVVLVA(moves))

	for _, m := range moves {
		if !r.b.PushMove(m) {
			continue
		}

		considerable := false
		if r.b.Position().IsCheckMate(turn.Opponent()) {
			considerable = true
		}
		if m.IsCapture() {
			if mayRecapture && m.To == target {
				considerable = true
			}
			if pieceValue(m.Piece) < pieceValue(m.Capture) {
				considerable = true
			}
			if !r.b.Position().IsAttacked(turn, m.To) {
				considerable = true
			}
		}

		if considerable {
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

func evaluate(ctx context.Context, b *board.Board, evaluator eval.Evaluator) eval.Score {
	score := evaluator.Evaluate(ctx, b.Position(), b.Turn())
	if b.HasCastled(b.Turn()) && score.Type == eval.Heuristic {
		score = eval.HeuristicScore(score.Pawns + 0.1)
	}
	return score
}

package turochamp

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
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

func (q Quiescence) QuietSearch(ctx context.Context, b *board.Board, alpha, beta board.Score, quit <-chan struct{}) (uint64, board.Score) {
	run := &runQuiescence{eval: q.Eval, b: b, quit: quit}
	score := run.search(ctx)
	return run.nodes, b.Turn().Unit() * score
}

type runQuiescence struct {
	eval  eval.Evaluator
	b     *board.Board
	nodes uint64

	quit <-chan struct{}
}

// search returns the positive score for the color.
func (r *runQuiescence) search(ctx context.Context) board.Score {
	if search.IsClosed(r.quit) {
		return 0
	}

	r.nodes++

	hasLegalMoves := false
	turn := r.b.Turn()
	score := turn.Unit() * evaluate(ctx, r.b, r.eval)

	mayRecapture := false
	var target board.Square
	if m, ok := r.b.LastMove(); ok && m.IsCapture() {
		mayRecapture = true
		target = m.To
	}

	// lastmove, _ := r.b.LastMove();
	// logw.Debugf(ctx, "SCORE: %v (last: %v, recapture=%v)= %v", r.b.Position(), lastmove, mayRecapture, score)

	moves := r.b.Position().PseudoLegalMoves(turn)
	for _, m := range moves {
		if r.b.PushMove(m) {
			hasLegalMoves = true

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
				if s := r.search(ctx); score < -s {
					score = -s
				}
			}
			r.b.PopMove()
		}
	}

	if !hasLegalMoves {
		if result := r.b.AdjudicateNoLegalMoves(); result.Reason == board.Checkmate {
			return board.MinScore
		}
		return 0
	}
	return score
}

func evaluate(ctx context.Context, b *board.Board, evaluator eval.Evaluator) board.Score {
	score := evaluator.Evaluate(ctx, b.Position(), b.Turn())
	if b.HasCastled(b.Turn()) {
		score += b.Turn().Unit() * 10
	}
	return score
}

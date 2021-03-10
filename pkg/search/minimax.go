package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// Minimax implements naive minimax search. Useful for comparison and validation.
type Minimax struct {
	Eval eval.Evaluator
}

func (m Minimax) Search(ctx context.Context, b *board.Board, depth int, quit <-chan struct{}) (uint64, board.Score, []board.Move, error) {
	run := &runMinimax{eval: m.Eval, b: b, quit: quit}
	score, moves := run.search(ctx, depth)
	if isClosed(quit) {
		return 0, 0, nil, ErrHalted
	}
	return run.nodes, b.Turn().Unit() * score, moves, nil
}

type runMinimax struct {
	eval  eval.Evaluator
	b     *board.Board
	nodes uint64

	quit <-chan struct{}
}

// search returns the positive score for the color.
func (m *runMinimax) search(ctx context.Context, depth int) (board.Score, []board.Move) {
	m.nodes++

	if isClosed(m.quit) {
		return 0, nil
	}
	if m.b.Result().Outcome == board.Draw {
		return 0, nil
	}
	if depth == 0 {
		return m.b.Turn().Unit() * m.eval.Evaluate(ctx, m.b.Position(), m.b.Turn()), nil
	}

	hasLegalMove := false
	score := board.MinScore - 1
	var pv []board.Move

	moves := m.b.Position().PseudoLegalMoves(m.b.Turn())
	for _, move := range moves {
		if m.b.PushMove(move) {
			s, rem := m.search(ctx, depth-1)
			m.b.PopMove()

			hasLegalMove = true
			if score < -s {
				score = -s
				pv = append([]board.Move{move}, rem...)
			}
		}
	}

	if !hasLegalMove {
		if m.b.Position().IsChecked(m.b.Turn()) {
			m.b.Adjudicate(board.Result{Outcome: board.Loss(m.b.Turn()), Reason: board.Checkmate})
			score = board.MinScore
		} else {
			m.b.Adjudicate(board.Result{Outcome: board.Draw, Reason: board.Stalemate})
			score = 0
		}
	}

	return score, pv
}

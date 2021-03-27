package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// Minimax implements naive minimax search. Useful for comparison and validation.
// Pseudo-code:
//
// function minimax(node, depth, maximizingPlayer) is
//    if depth = 0 or node is a terminal node then
//        return the heuristic value of node
//    if maximizingPlayer then
//        value := −∞
//        for each child of node do
//            value := max(value, minimax(child, depth − 1, FALSE))
//        return value
//    else (* minimizing player *)
//        value := +∞
//        for each child of node do
//            value := min(value, minimax(child, depth − 1, TRUE))
//        return value
//
// See: https://en.wikipedia.org/wiki/Minimax.
type Minimax struct {
	Eval eval.Evaluator
}

func (m Minimax) Search(ctx context.Context, b *board.Board, depth int, quit <-chan struct{}) (uint64, eval.Score, []board.Move, error) {
	run := &runMinimax{eval: m.Eval, b: b, quit: quit}
	score, moves := run.search(ctx, depth)
	if IsClosed(quit) {
		return 0, eval.Score{}, nil, ErrHalted
	}
	return run.nodes, score, moves, nil
}

type runMinimax struct {
	eval  eval.Evaluator
	b     *board.Board
	nodes uint64

	quit <-chan struct{}
}

// search returns the positive score for the color.
func (m *runMinimax) search(ctx context.Context, depth int) (eval.Score, []board.Move) {
	m.nodes++

	if IsClosed(m.quit) {
		return eval.ZeroScore, nil
	}
	if m.b.Result().Outcome == board.Draw {
		return eval.ZeroScore, nil
	}
	if depth == 0 {
		return m.eval.Evaluate(ctx, m.b), nil
	}

	hasLegalMove := false
	score := eval.NegInfScore
	var pv []board.Move

	moves := m.b.Position().PseudoLegalMoves(m.b.Turn())
	for _, move := range moves {
		if m.b.PushMove(move) {
			s, rem := m.search(ctx, depth-1)
			m.b.PopMove()

			hasLegalMove = true
			s = eval.IncrementMateDistance(s).Negate()
			if score.Less(s) {
				score = s
				pv = append([]board.Move{move}, rem...)
			}
		}
	}

	if !hasLegalMove {
		if result := m.b.AdjudicateNoLegalMoves(); result.Reason == board.Checkmate {
			return eval.NegInfScore, nil
		}
		return eval.ZeroScore, nil
	}

	return score, pv
}

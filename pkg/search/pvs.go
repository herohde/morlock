package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// PVS implements principal variation search. Pseudo-code:
//
// function pvs(node, depth, α, β, color) is
//    if depth = 0 or node is a terminal node then
//        return color × the heuristic value of node
//    for each child of node do
//        if child is first child then
//            score := −pvs(child, depth − 1, −β, −α, −color)
//        else
//            score := −pvs(child, depth − 1, −α − 1, −α, −color) (* search with a null window *)
//            if α < score < β then
//                score := −pvs(child, depth − 1, −β, −score, −color) (* if it failed high, do a full re-search *)
//        α := max(α, score)
//        if α ≥ β then
//            break (* beta cut-off *)
//    return α
//
// See: https://en.wikipedia.org/wiki/Principal_variation_search.
type PVS struct {
	Eval eval.Evaluator
}

func (p PVS) Search(ctx context.Context, b *board.Board, depth int, quit <-chan struct{}) (uint64, board.Score, []board.Move, error) {
	run := &runPVS{eval: p.Eval, b: b, quit: quit}
	score, moves := run.search(ctx, depth, board.MinScore-1, board.MaxScore+1)
	if isClosed(quit) {
		return 0, 0, nil, ErrHalted
	}
	return run.nodes, b.Turn().Unit() * score, moves, nil
}

type runPVS struct {
	eval  eval.Evaluator
	b     *board.Board
	nodes uint64

	quit <-chan struct{}
}

// search returns the positive score for the color.
func (m *runPVS) search(ctx context.Context, depth int, alpha, beta board.Score) (board.Score, []board.Move) {
	m.nodes++

	if isClosed(m.quit) {
		return 0, nil
	}
	if m.b.Result().Outcome == board.Draw {
		return 0, nil
	}
	if depth == 0 {
		return m.b.Turn().Unit() * board.CropScore(m.eval.Evaluate(ctx, m.b.Position(), m.b.Turn())), nil
	}

	hasLegalMove := false
	var pv []board.Move

	moves := m.b.Position().PseudoLegalMoves(m.b.Turn())
	for _, move := range moves {
		if m.b.PushMove(move) {
			var score board.Score
			var rem []board.Move

			if !hasLegalMove {
				score, rem = m.search(ctx, depth-1, -beta, -alpha)
				pv = append([]board.Move{move}, rem...)
			} else {
				// Search with a null window.

				score, rem = m.search(ctx, depth-1, -alpha-1, -alpha)
				if alpha < -score { // && -score < beta {
					// If it fails high, re-search with a full window.
					score, rem = m.search(ctx, depth-1, -beta, score)
				}
			}
			m.b.PopMove()

			hasLegalMove = true
			if alpha < -score {
				alpha = -score
				pv = append([]board.Move{move}, rem...)
			}

			if alpha >= beta {
				break // cutoff
			}
		}
	}

	if !hasLegalMove {
		if m.b.Position().IsChecked(m.b.Turn()) {
			m.b.Adjudicate(board.Result{Outcome: board.Loss(m.b.Turn()), Reason: board.Checkmate})
			return board.MinScore, nil
		} else {
			m.b.Adjudicate(board.Result{Outcome: board.Draw, Reason: board.Stalemate})
			return 0, nil
		}
	}

	return alpha, pv
}

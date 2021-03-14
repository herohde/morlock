package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"sort"
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
	Eval QuietSearch
}

func (p PVS) Search(ctx context.Context, b *board.Board, depth int, quit <-chan struct{}) (uint64, board.Score, []board.Move, error) {
	run := &runPVS{eval: p.Eval, b: b, quit: quit}
	score, moves := run.search(ctx, depth, board.MinScore-1, board.MaxScore+1)
	if IsClosed(quit) {
		return 0, 0, nil, ErrHalted
	}
	return run.nodes, b.Turn().Unit() * score, moves, nil
}

type runPVS struct {
	eval  QuietSearch
	b     *board.Board
	nodes uint64

	quit <-chan struct{}
}

// search returns the positive score for the color.
func (m *runPVS) search(ctx context.Context, depth int, alpha, beta board.Score) (board.Score, []board.Move) {
	if IsClosed(m.quit) {
		return 0, nil
	}
	if m.b.Result().Outcome == board.Draw {
		return 0, nil
	}
	if depth == 0 {
		if m.b.Turn() == board.Black {
			alpha, beta = -beta, -alpha
		}
		nodes, score := m.eval.QuietSearch(ctx, m.b, alpha, beta, m.quit)
		m.nodes += nodes
		return m.b.Turn().Unit() * score, nil
	}

	m.nodes++

	hasLegalMove := false
	var pv []board.Move

	moves := m.b.Position().PseudoLegalMoves(m.b.Turn())
	sort.Sort(board.ByScore(moves))

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
		if result := m.b.AdjudicateNoLegalMoves(); result.Reason == board.Checkmate {
			return board.MinScore, nil
		}
		return 0, nil
	}

	return alpha, pv
}

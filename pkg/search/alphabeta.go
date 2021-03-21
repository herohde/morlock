package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"sort"
)

// AlphaBeta implements alpha-beta pruning. Pseudo-code:
//
// function alphabeta(node, depth, α, β, maximizingPlayer) is
//    if depth = 0 or node is a terminal node then
//        return the heuristic value of node
//    if maximizingPlayer then
//        value := −∞
//        for each child of node do
//            value := max(value, alphabeta(child, depth − 1, α, β, FALSE))
//            α := max(α, value)
//            if α ≥ β then
//                break (* β cutoff *)
//        return value
//    else
//        value := +∞
//        for each child of node do
//            value := min(value, alphabeta(child, depth − 1, α, β, TRUE))
//            β := min(β, value)
//            if β ≤ α then
//                break (* α cutoff *)
//        return value
//
// See: https://en.wikipedia.org/wiki/Alpha–beta_pruning.
type AlphaBeta struct {
	Eval QuietSearch
}

func (p AlphaBeta) Search(ctx context.Context, b *board.Board, depth int, quit <-chan struct{}) (uint64, eval.Score, []board.Move, error) {
	run := &runAlphaBeta{eval: p.Eval, b: b, quit: quit}
	score, moves := run.search(ctx, depth, eval.NegInfScore, eval.InfScore)
	if IsClosed(quit) {
		return 0, eval.ZeroScore, nil, ErrHalted
	}
	return run.nodes, score, moves, nil
}

type runAlphaBeta struct {
	eval  QuietSearch
	b     *board.Board
	nodes uint64

	quit <-chan struct{}
}

// search returns the positive score for the color.
func (m *runAlphaBeta) search(ctx context.Context, depth int, alpha, beta eval.Score) (eval.Score, []board.Move) {
	if IsClosed(m.quit) {
		return eval.ZeroScore, nil
	}
	if m.b.Result().Outcome == board.Draw {
		return eval.ZeroScore, nil
	}
	if depth == 0 {
		nodes, score := m.eval.QuietSearch(ctx, m.b, alpha, beta, m.quit)
		m.nodes += nodes
		return score, nil
	}

	m.nodes++

	hasLegalMove := false
	var pv []board.Move

	moves := m.b.Position().PseudoLegalMoves(m.b.Turn())
	sort.Sort(board.ByMVVLVA(moves))

	for _, move := range moves {
		if m.b.PushMove(move) {
			score, rem := m.search(ctx, depth-1, beta.Negate(), alpha.Negate())
			m.b.PopMove()

			hasLegalMove = true
			score = eval.IncrementMateInX(score).Negate()
			if alpha.Less(score) {
				alpha = score
				pv = append([]board.Move{move}, rem...)
			}
			if alpha == beta || beta.Less(alpha) {
				break // cutoff
			}
		}
	}

	if !hasLegalMove {
		if result := m.b.AdjudicateNoLegalMoves(); result.Reason == board.Checkmate {
			return eval.NegInfScore, nil
		}
		return eval.ZeroScore, nil
	}

	return alpha, pv
}

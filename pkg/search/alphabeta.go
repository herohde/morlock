package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
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
	Pick Selection
	Eval QuietSearch
}

func (p AlphaBeta) Search(ctx context.Context, sctx *Context, b *board.Board, depth int, quit <-chan struct{}) (uint64, eval.Score, []board.Move, error) {
	run := &runAlphaBeta{pick: pick(p.Pick), ponder: sctx.Ponder, eval: p.Eval, tt: sctx.TT, b: b, quit: quit}
	low, high := eval.NegInfScore, eval.InfScore
	if !sctx.Alpha.IsInvalid() {
		low = sctx.Alpha
	}
	if !sctx.Beta.IsInvalid() {
		high = sctx.Beta
	}

	score, moves := run.search(ctx, depth, low, high)
	if IsClosed(quit) {
		return 0, eval.InvalidScore, nil, ErrHalted
	}
	return run.nodes, score, moves, nil
}

type runAlphaBeta struct {
	pick  Selection
	eval  QuietSearch
	tt    TranspositionTable
	b     *board.Board
	nodes uint64

	ponder []board.Move

	quit <-chan struct{}
}

// search returns the positive score for the color.
func (m *runAlphaBeta) search(ctx context.Context, depth int, alpha, beta eval.Score) (eval.Score, []board.Move) {
	if IsClosed(m.quit) {
		return eval.InvalidScore, nil
	}
	if m.b.Result().Outcome == board.Draw {
		return eval.ZeroScore, nil
	}

	var best board.Move
	if bound, d, score, m, ok := m.tt.Read(m.b.Hash()); ok {
		best = m
		if depth <= d {
			isOpp := (d-depth)%2 != 0
			if isOpp {
				// if opposing side move, then score is negative and an upper bound.
				score = score.Negate()
			}
			if (bound == ExactBound || !isOpp) && (beta == score || beta.Less(score)) {
				// logw.Debugf(ctx, "TT: %v@%v = %v, %v", bound, d, score, move)
				return score, nil // cutoff
			}
		}
	}

	if depth == 0 {
		sctx := &Context{Alpha: alpha, Beta: beta, TT: m.tt}
		nodes, score := m.eval.QuietSearch(ctx, sctx, m.b, m.quit)
		m.nodes += nodes

		m.tt.Write(m.b.Hash(), ExactBound, m.b.Ply(), 0, score, board.Move{})
		return score, nil
	}

	m.nodes++

	var ponder board.Move
	ponderOnly := false
	if len(m.ponder) > 0 {
		ponder = m.ponder[0]
		m.ponder = m.ponder[1:]
		ponderOnly = true
	}

	hasLegalMove := false
	bound := ExactBound
	var pv []board.Move

	moves := NewMoveList(m.b.Position().PseudoLegalMoves(m.b.Turn()), First(best).MVVLVA)
	for {
		move, ok := moves.Next()
		if !ok {
			break
		}

		if !m.b.PushMove(move) {
			continue
		}

		if m.pick(ctx, move, m.b) && (!ponderOnly || ponder.Equals(move)) {
			score, rem := m.search(ctx, depth-1, beta.Negate(), alpha.Negate())
			score = eval.IncrementMateDistance(score).Negate()
			if alpha.Less(score) {
				alpha = score
				pv = append([]board.Move{move}, rem...)
			}
		}

		m.b.PopMove()
		hasLegalMove = true

		if alpha == beta || beta.Less(alpha) {
			bound = LowerBound
			break // cutoff
		}
	}

	if !hasLegalMove {
		if result := m.b.AdjudicateNoLegalMoves(); result.Reason == board.Checkmate {
			return eval.NegInfScore, nil
		}
		return eval.ZeroScore, nil
	}

	m.tt.Write(m.b.Hash(), bound, m.b.Ply(), depth, alpha, first(pv))
	return alpha, pv
}

func first(pv []board.Move) board.Move {
	if len(pv) == 0 {
		return board.Move{}
	}
	return pv[0]
}

func pick(p Selection) Selection {
	if p == nil {
		return IsAnyMove
	}
	return p
}

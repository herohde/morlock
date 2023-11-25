package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/seekerror/stdlib/pkg/util/contextx"
)

// AlphaBeta implements alpha-beta pruning. Pseudo-code:
//
// function alphabeta(node, depth, α, β, maximizingPlayer) is
//
//	if depth = 0 or node is a terminal node then
//	    return the heuristic value of node
//	if maximizingPlayer then
//	    value := −∞
//	    for each child of node do
//	        value := max(value, alphabeta(child, depth − 1, α, β, FALSE))
//	        α := max(α, value)
//	        if α ≥ β then
//	            break (* β cutoff *)
//	    return value
//	else
//	    value := +∞
//	    for each child of node do
//	        value := min(value, alphabeta(child, depth − 1, α, β, TRUE))
//	        β := min(β, value)
//	        if β ≤ α then
//	            break (* α cutoff *)
//	    return value
//
// See: https://en.wikipedia.org/wiki/Alpha–beta_pruning.
type AlphaBeta struct {
	Explore Exploration
	Eval    QuietSearch
}

func (p AlphaBeta) Search(ctx context.Context, sctx *Context, b *board.Board, depth int) (uint64, eval.Score, []board.Move, error) {
	run := &runAlphaBeta{
		explore: fullIfNotSet(p.Explore),
		eval:    p.Eval,
		tt:      sctx.TT,
		noise:   sctx.Noise,
		ponder:  sctx.Ponder,
		b:       b,
	}
	low, high := eval.NegInfScore, eval.InfScore
	if !sctx.Alpha.IsInvalid() {
		low = sctx.Alpha
	}
	if !sctx.Beta.IsInvalid() {
		high = sctx.Beta
	}

	score, moves := run.search(ctx, depth, low, high)
	if contextx.IsCancelled(ctx) {
		return 0, eval.InvalidScore, nil, ErrHalted
	}
	return run.nodes, score, moves, nil
}

type runAlphaBeta struct {
	explore Exploration
	eval    QuietSearch
	tt      TranspositionTable
	noise   eval.Random
	b       *board.Board
	nodes   uint64

	ponder []board.Move
}

// search returns the positive score for the color.
func (m *runAlphaBeta) search(ctx context.Context, depth int, alpha, beta eval.Score) (eval.Score, []board.Move) {
	if contextx.IsCancelled(ctx) {
		return eval.InvalidScore, nil
	}
	if m.b.Result().Outcome == board.Draw {
		return eval.ZeroScore, nil
	}

	var best board.Move
	if bound, d, score, m, ok := m.tt.Read(m.b.Hash()); ok {
		best = m
		if depth == d && bound == ExactBound {
			// logw.Debugf(ctx, "TT: %v@%v = %v, %v", bound, d, score, move)
			return score, nil // cutoff
		} // else: not deep enough or precise enough
	}

	if depth == 0 {
		sctx := &Context{Alpha: alpha, Beta: beta, TT: m.tt, Noise: m.noise}
		nodes, score := m.eval.QuietSearch(ctx, sctx, m.b)
		m.nodes += nodes

		m.tt.Write(m.b.Hash(), ExactBound, m.b.Ply(), 0, score, board.Move{})
		return score, nil
	}

	m.nodes++

	hasLegalMove := false
	bound := ExactBound
	var pv []board.Move

	priority, explore := m.explore(ctx, m.b)

	if len(m.ponder) > 0 {
		explore = m.ponder[0].Equals // overwrite: use ponder move even if not intended to be explored
		m.ponder = m.ponder[1:]
	}

	moves := board.NewMoveList(m.b.Position().PseudoLegalMoves(m.b.Turn()), board.First(best, priority))
	for {
		move, ok := moves.Next()
		if !ok {
			break
		}
		if !m.b.PushMove(move) {
			continue // skip: not legal
		}

		if explore(move) {
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

	if bound == ExactBound {
		m.tt.Write(m.b.Hash(), bound, m.b.Ply(), depth, alpha, firstOrNone(pv))
	}
	return alpha, pv
}

func firstOrNone(pv []board.Move) board.Move {
	if len(pv) == 0 {
		return board.Move{}
	}
	return pv[0]
}

func fullIfNotSet(p Exploration) Exploration {
	if p == nil {
		return FullExploration
	}
	return p
}

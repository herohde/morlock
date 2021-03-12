// Package turochamp implements the evaluation and search heuristics used by TUROCHAMP.
package turochamp

import (
	"context"
	"math"

	"github.com/herohde/morlock/pkg/board"
)

// Eval implements the TUROCHAMP evaluation function.
type Eval struct{}

func (Eval) Evaluate(ctx context.Context, pos *board.Position, turn board.Color) board.Score {
	mat := Material{}.Evaluate(ctx, pos, turn)
	if mat != 0 {
		// If nonzero, add 10k to the material score to ensure it dominates. Respect sign.
		mat = board.Score(math.Copysign(math.Abs(float64(mat))+10000, float64(mat)))
	}
	return mat + PositionPlay{}.Evaluate(ctx, pos, turn)
}

// Material returns the material advantage balance as a ratio, W/B. Turing (and Champernowne)
// used the following piece values: pawn=1, knight=3, bishop=3½, rook=5, queen=10. We convert
// the 100x ratio into a score that dominates the PositionPlay value to obtain Turing's semantics.
type Material struct{}

func (Material) Evaluate(ctx context.Context, pos *board.Position, turn board.Color) board.Score {
	own := material(pos, turn)
	opp := material(pos, turn.Opponent())

	switch {
	case own == opp:
		return 0
	case own > opp:
		return turn.Unit() * ratio(own, opp)
	default: // opp > own
		return turn.Opponent().Unit() * ratio(opp, own)
	}
}

func material(pos *board.Position, turn board.Color) board.Score {
	var score board.Score
	for _, piece := range board.QueenRookKnightBishopPawn {
		score += pieceValue(piece) * board.Score(pos.Piece(turn, piece).PopCount())
	}
	return score
}

func pieceValue(piece board.Piece) board.Score {
	switch piece {
	case board.King:
		return 10000
	case board.Queen:
		return 1000
	case board.Rook:
		return 500
	case board.Bishop:
		return 350
	case board.Knight:
		return 300
	case board.Pawn:
		return 100
	default:
		panic("invalid piece")
	}
}

func ratio(a, b board.Score) board.Score {
	if b == 0 {
		return a
	}
	return board.Score(100 * float64(a) / float64(b))
}

// PositionPlay captures the following positional evaluation functions:
//
//  * Mobility. For the Q,R,B,N, add the square root of the number of legal moves the
//    piece can make; count each capture as two moves.
//
//  * Piece safety. For the R,B,N, add 1.0 point if it is defended, and 1.5 points
//    if it is defended at least twice.
//
//  * King mobility. For the K, the same as (1) except for castling moves.
//
//  * King safety. For the K, deduct points for its vulnerability as follows: assume
//    that a Queen of the same colour is on the King's square; calculate its mobility,
//    and then subtract this value from the score.
//
//  * Castling. Add 1.0 point for the possibility of still being able to castle on a
//    later move if a King or Rook move is being considered; add another point if
//    castling can take place on the next move; finally add one more point for
//    actually castling.
//
//  * Pawn credit. Add 0.2 point for each rank advanced, and 0.3 point for being
//    defended by a non-Pawn.
//
//  * Mates and checks. Add 1.0 point for the threat of mate and 0.5 point for a check.
//
// We score with a 10x multiplier to get to 1 decimal point precision as described.
type PositionPlay struct{}

func (PositionPlay) Evaluate(ctx context.Context, pos *board.Position, turn board.Color) board.Score {
	var score board.Score

	if pos.Castling()&board.CastlingRights(turn) != 0 {
		score += 10
	}

	// (1) Analyze mobility, castling and checks/checkmates.

	mobility := map[board.Square]int{}
	var mayCheckMate, mayCheck, mayCastle bool

	for _, m := range pos.PseudoLegalMoves(turn) {
		next, ok := pos.Move(m)
		if !ok {
			continue // not legal
		}

		if !mayCheckMate && next.IsCheckMate(turn.Opponent()) {
			mayCheckMate = true
			score += 10
		} else if !mayCheck && next.IsChecked(turn.Opponent()) {
			mayCheck = true
			score += 10
		}
		if !mayCastle && m.IsCastle() {
			mayCastle = true
			score += 10
		}

		if m.Piece != board.Pawn && !m.IsCastle() {
			mobility[m.From]++
			if m.Type == board.Capture {
				mobility[m.From]++
			}
		}
	}
	for _, n := range mobility {
		score += board.Score(10 * math.Sqrt(float64(n)))
	}

	// (2) Analyze Rook, Knight, Bishop defence.

	middle := pos.Piece(turn, board.Rook) | pos.Piece(turn, board.Knight) | pos.Piece(turn, board.Bishop)
	for middle != 0 {
		from := middle.LastPopSquare()
		middle ^= board.BitMask(from)

		defenders := 0
		for _, p := range board.KingQueenRookKnightBishop {
			if bb := board.Attackboard(pos.Rotated(), from, p) & pos.Piece(turn, p); bb != 0 {
				defenders += bb.PopCount()
			}
		}
		if bb := board.PawnCaptureboard(turn, pos.Piece(turn, board.Pawn)); bb != 0 {
			defenders += bb.PopCount()
		}
		if defenders > 0 {
			score += 10
		}
		if defenders > 1 {
			score += 5
		}
	}

	// (3) Analyze King safety.

	if king := pos.Piece(turn, board.King); king != 0 {
		attackboard := board.QueenAttackboard(pos.Rotated(), king.LastPopSquare())
		safety := attackboard.PopCount()
		safety += (attackboard & pos.Color(turn.Opponent())).PopCount()

		score -= board.Score(10 * math.Sqrt(float64(safety)))
	}

	// (4) Analyze Pawn progress and defence.

	pawns := pos.Piece(turn, board.Pawn)
	for pawns != 0 {
		from := pawns.LastPopSquare()
		pawns ^= board.BitMask(from)

		ranks := 0
		if turn == board.White {
			ranks += int(from.Rank() - 2)
		} else {
			ranks += int(7 - from.Rank())
		}

		score += 2 * board.Score(ranks)

		for _, p := range board.KingQueenRookKnightBishop {
			if board.Attackboard(pos.Rotated(), from, p)&pos.Piece(turn, p) != 0 {
				score += 3
				break
			}
		}
	}

	return turn.Unit() * score
}

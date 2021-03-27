// Package turochamp implements the evaluation and search heuristics used by TUROCHAMP.
package turochamp

import (
	"context"
	"github.com/herohde/morlock/pkg/eval"
	"math"

	"github.com/herohde/morlock/pkg/board"
)

// Eval implements the TUROCHAMP evaluation function.
type Eval struct{}

func (Eval) Evaluate(ctx context.Context, b *board.Board) eval.Score {
	mat := Material{}.Evaluate(ctx, b)
	pp := PositionPlay{}.Evaluate(ctx, b)

	// Combine scores to ensure material strictly dominates: MMMMMP.PP.

	m := eval.Pawns(math.Round(float64(mat.Pawns)*100) * 10)
	p := eval.Pawns(math.Round(float64(pp.Pawns)*100) / 1000)

	// println(fmt.Sprintf("POS %v MAT: %v -> %v, PP: %v -> %v => %v", pos, mat, m, pp, p, m+p))

	return eval.HeuristicScore(m + p)
}

// Material returns the material advantage balance as a ratio, W/B. Turing and Champernowne
// used the following piece values: pawn=1, knight=3, bishop=3½, rook=5, queen=10. The ratio
// in the range of [-226;226]. We use a negative ratio for when behind to let position-play
// dominate in that case.
type Material struct{}

func (Material) Evaluate(ctx context.Context, b *board.Board) eval.Score {
	pos := b.Position()
	turn := b.Turn()

	own := material(pos, turn)
	opp := material(pos, turn.Opponent())

	switch {
	case own == opp:
		return eval.ZeroScore
	case own > opp:
		return eval.HeuristicScore(own / opp)
	default: // opp > own
		return eval.HeuristicScore(-opp / own)
	}
}

func material(pos *board.Position, turn board.Color) eval.Pawns {
	var score eval.Pawns
	for _, piece := range board.QueenRookKnightBishopPawn {
		score += pieceValue(piece) * eval.Pawns(pos.Piece(turn, piece).PopCount())
	}
	if score == 0 {
		return 0.5 // half-a-pawn if only piece left is the king
	}
	return score
}

func pieceValue(piece board.Piece) eval.Pawns {
	switch piece {
	case board.King:
		return 100
	case board.Queen:
		return 10
	case board.Rook:
		return 5
	case board.Bishop:
		return 3.5
	case board.Knight:
		return 3
	case board.Pawn:
		return 1
	default:
		panic("invalid piece")
	}
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
// We score with 1 decimal point precision as described. The range is [-55;55].
type PositionPlay struct{}

func (PositionPlay) Evaluate(ctx context.Context, b *board.Board) eval.Score {
	pos := b.Position()
	turn := b.Turn()

	var score eval.Pawns

	if pos.Castling()&board.CastlingRights(turn) != 0 {
		score += 1
	}
	if b.HasCastled(turn) {
		score += 1
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
			score += 1
		} else if !mayCheck && next.IsChecked(turn.Opponent()) {
			mayCheck = true
			score += 1
		}
		if !mayCastle && m.IsCastle() {
			mayCastle = true
			score += 1
		}

		if m.Piece != board.Pawn && !m.IsCastle() {
			mobility[m.From]++
			if m.Type == board.Capture {
				mobility[m.From]++
			}
		}
	}
	for _, n := range mobility {
		score += eval.Pawns(math.Round(10*math.Sqrt(float64(n)))) / 10
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
		if bb := board.PawnCaptureboard(turn, pos.Piece(turn, board.Pawn)) & board.BitMask(from); bb != 0 {
			defenders += bb.PopCount()
		}
		if defenders > 0 {
			score += 1
		}
		if defenders > 1 {
			score += 0.5
		}
	}

	// (3) Analyze King safety.

	if king := pos.Piece(turn, board.King); king != 0 {
		attackboard := board.QueenAttackboard(pos.Rotated(), king.LastPopSquare())
		safety := attackboard.PopCount()
		safety += (attackboard & pos.Color(turn.Opponent())).PopCount()

		score -= eval.Pawns(math.Round(10*math.Sqrt(float64(safety)))) / 10
	}

	// (4) Analyze Pawn progress and defence.

	pawns := pos.Piece(turn, board.Pawn)
	for pawns != 0 {
		from := pawns.LastPopSquare()
		pawns ^= board.BitMask(from)

		ranks := 0
		if turn == board.White {
			ranks += int(from.Rank() - board.Rank2)
		} else {
			ranks += int(board.Rank7 - from.Rank())
		}

		score += 0.2 * eval.Pawns(ranks)

		for _, p := range board.KingQueenRookKnightBishop {
			if bb := board.Attackboard(pos.Rotated(), from, p) & pos.Piece(turn, p); bb != 0 {
				score += 0.3
				break
			}
		}
	}

	return eval.HeuristicScore(score)
}

package bernstein

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/seekerror/stdlib/pkg/util/mathx"
)

// Eval implements the evaluation heuristic: the value, or score, is measured
// by four considerations:
//
//	(1) mobility of the pieces
//	(2) control of important squares.
//	(3) defense of the king (= number controlled squares around the King);
//	(4) gain of material (a pawn counting as one unit, a knight or bishop three,
//	    a rook five and the queen nine);
//
// "The program counts the number of available moves for each side: the number of
// squares completely controlled by each side; the number of controlled squares
// around each king and the material count of each side.
//
// These criteria are parametrized for easy variation. so that different values
// may be given to center squares, or to the squares around the king. At present
// criteria 1, 2, and 3 are added together while 4 is multiplied by a large factor
// before adding it to the total. This prevents the machine from sacrificing
// material, and encourages it to exchange when it is ahead, as it considers the
// ratio of its own and opponent's scores"
type Eval struct {
	Factor int
}

func (e Eval) Evaluate(ctx context.Context, b *board.Board) eval.Pawns {
	self := Evaluate(b.Position(), e.Factor, b.Turn())
	opp := Evaluate(b.Position(), e.Factor, b.Turn().Opponent())

	// Convert ratio to signed value. Of course, the absolute number does not make
	// sense as a "pawns" evaluation.

	switch {
	case self == opp:
		return 0
	case self > opp:
		return eval.Pawns(self) / eval.Pawns(opp)
	default:
		return -eval.Pawns(opp) / eval.Pawns(self)
	}
}

func Evaluate(pos *board.Position, factor int, side board.Color) int {
	mobility := Mobility(pos, side)
	control := Control(pos, side)
	defense := KingDefense(pos, side)
	material := Material(pos, side)

	score := mobility + control + defense + factor*material

	// NOTE(herohde) 11/19/2023: as a technicality, we don't return zero for a side with
	// no moves, no pieces and no control (= a potential stalemate position). There is no
	// precise description of what "control" means, so it is possible it never happened
	// in the original program.

	return mathx.Max(1, score)
}

// Material returns the nominal material values for the side, ignoring the king.
func Material(pos *board.Position, side board.Color) int {
	ret := MaterialValue(board.Queen) * pos.Piece(side, board.Queen).PopCount()
	ret += MaterialValue(board.Rook) * pos.Piece(side, board.Rook).PopCount()
	ret += MaterialValue(board.Knight) * pos.Piece(side, board.Knight).PopCount()
	ret += MaterialValue(board.Bishop) * pos.Piece(side, board.Bishop).PopCount()
	ret += MaterialValue(board.Pawn) * pos.Piece(side, board.Pawn).PopCount()
	return ret
}

// MaterialValue returns the nominal material values for a piece.
func MaterialValue(piece board.Piece) int {
	switch piece {
	case board.King:
		return 100
	case board.Queen:
		return 9
	case board.Rook:
		return 5
	case board.Knight, board.Bishop:
		return 3
	case board.Pawn:
		return 1
	default:
		return 0
	}
}

// Mobility returns the number of legal moves.
func Mobility(pos *board.Position, side board.Color) int {
	return len(pos.LegalMoves(side))
}

// Control returns the number of squares defended by the given side, but with no opponent
// attackers. Populated squares included.
func Control(pos *board.Position, side board.Color) int {
	ret := 0
	for sq := board.ZeroSquare; sq < board.NumSquares; sq++ {
		if pos.IsDefended(side, sq) && !pos.IsAttacked(side, sq) {
			ret++
		}
	}
	return ret
}

// KingDefense returns the number of squares around the king defended by the given side, but
// with no opponent attackers. Populated squares included.
func KingDefense(pos *board.Position, side board.Color) int {
	ret := 0
	for _, sq := range board.KingAttackboard(pos.KingSquare(side)).ToSquares() {
		if pos.IsDefended(side, sq) && !pos.IsAttacked(side, sq) {
			ret++
		}
	}
	return ret
}

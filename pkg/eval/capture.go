package eval

import (
	"github.com/herohde/morlock/pkg/board"
	"sort"
)

// FindCapture returns the pieces of the given color that directly target the square.
func FindCapture(pos *board.Position, side board.Color, sq board.Square) []board.Placement {
	var ret []board.Placement

	for _, piece := range board.KingQueenRookKnightBishop {
		bb := board.Attackboard(pos.Rotated(), sq, piece) & pos.Piece(side, piece)
		for _, from := range bb.ToSquares() {
			ret = append(ret, board.Placement{Piece: piece, Color: side, Square: from})
		}
	}
	bb := board.PawnCaptureboard(side.Opponent() /* reverse direction */, board.BitMask(sq)) & pos.Piece(side, board.Pawn)
	for _, from := range bb.ToSquares() {
		ret = append(ret, board.Placement{Piece: board.Pawn, Color: side, Square: from})
	}

	return ret
}

// SortByNominalValue orders the placement list by nominal material value, low to high.
func SortByNominalValue(pieces []board.Placement) []board.Placement {
	sort.SliceStable(pieces, func(i, j int) bool {
		return NominalValue(pieces[i].Piece) < NominalValue(pieces[j].Piece)
	})
	return pieces
}

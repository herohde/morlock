package bernstein

import (
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// TODO(herohde) 11/24/2023: unclear to what extent static exchange evaluation is performed.
// For now, keep it simple and explore how it predicts the published games.

// IsMoveSafe evaluates whether a move is safe, i.e., that the piece is adequately defended
// at its destination. Assumes legal. Takes into account sliding piece reach post-move.
func IsMoveSafe(pos *board.Position, side board.Color, move board.Move) bool {
	next, ok := pos.Move(move)
	if !ok {
		return false
	}
	return IsSafe(next, side, move.Piece, move.To)
}

// IsSafe evaluates whether an occupied square is safe, i.e., is not en prise or can be
// exchanged with immediate loss.
func IsSafe(pos *board.Position, side board.Color, piece board.Piece, sq board.Square) bool {
	attackers := eval.SortByNominalValue(eval.FindCapture(pos, side.Opponent(), sq))
	if len(attackers) == 0 {
		return true // ok: no attackers
	}
	if !pos.IsDefended(side, sq) {
		return false // not ok: en prise
	}
	return eval.NominalValue(attackers[0].Piece) >= eval.NominalValue(piece)
}

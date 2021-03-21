// Package eval contains position evaluation logic and utilities.
package eval

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
)

// Evaluator is a static position evaluator.
type Evaluator interface {
	// Evaluate returns the position score.
	Evaluate(ctx context.Context, pos *board.Position, turn board.Color) Score
}

// Material returns the nominal material advantage balance for the side to move.
type Material struct{}

func (Material) Evaluate(ctx context.Context, pos *board.Position, turn board.Color) Score {
	var pawns Pawns
	for p := board.ZeroPiece; p < board.NumPieces; p++ {
		pawns += Pawns(pos.Piece(turn, p).PopCount()-pos.Piece(turn.Opponent(), p).PopCount()) * NominalValue(p)
	}
	return HeuristicScore(pawns)
}

// NominalValue the absolute nominal value in pawns of a piece. The King has an arbitrary value of 100 pawns.
func NominalValue(p board.Piece) Pawns {
	switch p {
	case board.Pawn:
		return 1
	case board.Bishop, board.Knight:
		return 3
	case board.Rook:
		return 5
	case board.Queen:
		return 9
	case board.King:
		return 100
	default:
		return 0
	}
}

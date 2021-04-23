// Package eval contains position evaluation logic and utilities.
package eval

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
)

// Evaluator is a static position evaluator.
type Evaluator interface {
	// Evaluate returns the position score in Pawns.
	Evaluate(ctx context.Context, b *board.Board) Pawns
}

// Material returns the nominal material advantage balance for the side to move.
type Material struct{}

func (Material) Evaluate(ctx context.Context, b *board.Board) Pawns {
	pos := b.Position()
	turn := b.Turn()

	var pawns Pawns
	for p := board.ZeroPiece; p < board.NumPieces; p++ {
		pawns += Pawns(pos.Piece(turn, p).PopCount()-pos.Piece(turn.Opponent(), p).PopCount()) * NominalValue(p)
	}
	return pawns
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

// NominalValueGain is the nominal material gain for a move.
func NominalValueGain(m board.Move) Pawns {
	switch m.Type {
	case board.CapturePromotion:
		return NominalValue(m.Capture) + NominalValue(m.Promotion) - NominalValue(board.Pawn)
	case board.Promotion:
		return NominalValue(m.Promotion) - NominalValue(board.Pawn)
	case board.Capture:
		return NominalValue(m.Capture)
	case board.EnPassant:
		return NominalValue(board.Pawn)
	default:
		return 0
	}
}

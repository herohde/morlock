// Package eval contains position evaluation logic and utilities.
package eval

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
)

// Evaluator is a static position evaluator.
type Evaluator interface {
	// Evaluate returns the position score.
	Evaluate(ctx context.Context, pos *board.Position, turn board.Color) board.Score
}

// Material returns the nominal material advantage balance.
type Material struct{}

func (Material) Evaluate(ctx context.Context, pos *board.Position, turn board.Color) board.Score {
	var score board.Score
	for p := board.ZeroPiece; p < board.NumPieces; p++ {
		score += board.Score(pos.Piece(board.White, p).PopCount()-pos.Piece(board.Black, p).PopCount()) * p.NominalValue()
	}
	return score
}

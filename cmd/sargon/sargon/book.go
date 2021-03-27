package sargon

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
)

// Book contains the sargon opening book of playing either e2e4 or d2d4. If black,
// SARGON plays e7e5 against a/b/c or e pawn moves. Otherwise, d2d4.
type Book struct{}

func (b Book) Find(ctx context.Context, pos string) ([]board.Move, error) {
	return nil, nil
}

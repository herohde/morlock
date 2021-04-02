package sargon

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
)

// Book contains the SARGON opening book of playing either e2e4 or d2d4. If black,
// SARGON plays e7e5 against a/b/c or e pawn moves. Otherwise, d7d5.
type Book struct {
	moves map[string][]board.Move
}

var (
	e2e4 = board.Move{Type: board.Normal, From: board.E2, To: board.E4}
	d2d4 = board.Move{Type: board.Normal, From: board.D2, To: board.D4}
	e7e5 = board.Move{Type: board.Normal, From: board.E7, To: board.E5}
	d7d5 = board.Move{Type: board.Normal, From: board.D7, To: board.D5}
)

func NewBook() *Book {
	moves := map[string][]board.Move{
		fen.Strip(fen.Initial): {e2e4, d2d4},
	}

	pos, turn, _, _, _ := fen.Decode(fen.Initial)
	for _, m := range pos.LegalMoves(turn) {
		next, _ := pos.Move(m)

		response := d7d5
		if isQueenSideOrKingPawn(m) {
			response = e7e5
		}

		key := fen.Strip(fen.Encode(next, turn.Opponent(), 0, 1))
		moves[key] = []board.Move{response}
	}

	return &Book{moves: moves}
}

func (b *Book) Find(ctx context.Context, pos string) ([]board.Move, error) {
	return b.moves[fen.Strip(pos)], nil
}

func isQueenSideOrKingPawn(m board.Move) bool {
	if m.Piece != board.Pawn {
		return false
	}
	switch m.From.File() {
	case board.FileA, board.FileB, board.FileC, board.FileE:
		return true
	default:
		return false
	}
}

package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// Selection defines move selection. It is required by quiescence search, but optional
// for full search. Selection turns true if the move just made should be explored.
type Selection func(ctx context.Context, move board.Move, b *board.Board) bool

// IsAnyMove is a trivial selection of all moves. Default for full search.
func IsAnyMove(ctx context.Context, m board.Move, b *board.Board) bool {
	return true
}

// NoMove is a trivial selection of no moves. Used to disable quiescence.
func NoMove(ctx context.Context, m board.Move, b *board.Board) bool {
	return false
}

// IsNotUnderPromotion selects any move, except under-promotions.
func IsNotUnderPromotion(ctx context.Context, m board.Move, b *board.Board) bool {
	return !m.IsPromotion() || m.Promotion == board.Queen
}

// IsQuickGain is a move selection for immediate material gain: promotions and captures.
func IsQuickGain(ctx context.Context, m board.Move, b *board.Board) bool {
	explore := m.IsPromotion()
	if m.IsCapture() {
		if eval.NominalValue(m.Piece) < eval.NominalValue(m.Capture) {
			explore = true
		}
		if !b.Position().IsAttacked(b.Turn().Opponent(), m.To) {
			explore = true
		}
	}
	return explore
}

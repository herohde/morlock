package turochamp

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
)

// IsConsiderableMove implements the selective "considerable moves" search:
//   (1) Re-captures are considerable.
//   (2) Capture of en prise pieces are considerable.
//   (3) Capture of higher value pieces are considerable.
//   (4) Checkmate are considerable.
func IsConsiderableMove(ctx context.Context, m board.Move, b *board.Board) bool {
	considerable := b.Position().IsCheckMate(b.Turn())
	if m.IsCapture() {
		if last, ok := b.SecondToLastMove(); ok && last.IsCapture() && m.To == last.To {
			considerable = true
		}
		if pieceValue(m.Piece) < pieceValue(m.Capture) {
			considerable = true
		}
		if !b.Position().IsAttacked(b.Turn().Opponent(), m.To) {
			considerable = true
		}
	}

	return considerable
}

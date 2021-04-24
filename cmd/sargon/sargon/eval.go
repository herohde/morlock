package sargon

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// Points implements the POINTS evaluation. It uses the full score for material and board
// control, given we do not have a representation size limit. As long as they are disjoint
// and the LIMIT 6 BRDC value is still blended in, they should reflect the original scheme.
type Points struct {
	side0 board.Color
	brdc0 eval.Pawns
}

func (p *Points) Reset(ctx context.Context, b *board.Board) {
	pins := FindKingQueenPins(b.Position())

	p.side0 = b.Turn()
	p.brdc0 = BoardControl(ctx, b, pins)
}

func (p *Points) Evaluate(ctx context.Context, b *board.Board) eval.Pawns {
	pins := FindKingQueenPins(b.Position())

	brdc := BoardControl(ctx, b, pins)
	mtrl, ptschk := Material(ctx, b, pins)
	if ptschk {
		return mtrl*4 + brdc/100
	}

	brdc0 := p.brdc0
	if b.Turn() != p.side0 {
		brdc0 = -brdc0
	}
	return mtrl*4 + eval.Limit(brdc-p.brdc0, 6) + brdc/100
}

// Notes
//
// XCHNG: exchange value.
//  initial defender x2 (<-- not in BYTE article), but perhaps only to allow RxB if defended. Not QxP or QxB. Seems
//  like a trick to cutoff futile exchanges that we can never bounce back from.
//  Use:  3/4 for PTSW2.
//
// if LOSS
//   -0.5
//   Save greatest loss to PTSL
//   Save PTSCHK if piece first lost just moved.
// if WIN:
//   - Save greatest 2 wins as PTSW1, PTSW2.
//   - Use PTSW2 if moving piece was lost.
//
//  PSTL:  if >0 then -1
//   - adjustment: (PTSW2 -1)/2 - PSTL.

// Material implements the MTRL heuristic without limit plus the ptschk (= moving into loss).
func Material(ctx context.Context, b *board.Board, pins Pins) (eval.Pawns, bool) {
	pos := b.Position()
	turn := b.Turn()

	// Material uses: 1,3,3,5,9,10

	mtrl := eval.Material{}.Evaluate(ctx, b)

	last, ok := b.LastMove()

	var ptsl, ptsw1, ptsw2 eval.Pawns
	ptschk := false

	pieces := b.Position().All()
	for pieces != 0 {
		sq := pieces.LastPopSquare()
		pieces ^= board.BitMask(sq)

		v := Exchange(pos, pins, turn.Opponent(), sq)
		switch {
		case v < ptsl:
			ptsl = v
			if ok && last.To == sq {
				ptschk = true // not cleared if later square is greater loss?
			}
		case ptsw1 < v:
			ptsw1, ptsw2 = v, ptsw1
		case ptsw2 < v:
			ptsw2 = v
		}
	}

	if ptschk {
		// Use PTSW2 if moving piece is moving into losing exchange. Assumed unguarded win.
		ptsw1, ptsw2 = ptsw2, 0
	}

	// NOTE(herohde) 4/22/2021: the interaction between doubling the exchange value, 4x MTRL and
	// limit 6 BRDC and ptschk is subtle. Following the assembly seem to put pawns en prise.
	// Instead follow the BYTE article for 3/4 of PTSW2?

	loss := ptsl
	if loss < 0 {
		loss = 2*ptsl + 1
	}
	win := ptsw2
	if win > 0 {
		win = (2*ptsw2 - 1) / 2
	}

	// We swap win/loss, because the evaluation here is from the points of the side to move. Sargon
	// rather evaluates the last move.

	mtrl -= loss + win
	return mtrl, ptschk
}

// BoardControl implements the BRDC heuristic without limit.
func BoardControl(ctx context.Context, b *board.Board, pins Pins) eval.Pawns {
	return Development(ctx, b) + Mobility(ctx, b, pins)
}

// Mobility implements the development aspects of the BRDC heuristic, without limit.
func Mobility(ctx context.Context, b *board.Board, pins Pins) eval.Pawns {
	pos := b.Position()
	turn := b.Turn()

	var pawns eval.Pawns
	for sq := board.ZeroSquare; sq < board.NumSquares; sq++ {
		att := FindAttackers(pos, pins, sq, turn)
		opp := FindAttackers(pos, pins, sq, turn.Opponent())

		pawns += eval.Pawns(NumAttackers(att) - NumAttackers(opp))
	}
	return pawns
}

// Development implements the development aspects of the BRDC heuristic, without limit. It
// covers the following w/ the symmetrical difference from the opponent:
//  (1) KNIGHT/BISHOP: -2 if not moved.
//  (2) ROOK/QUEEN:    -2 if MOVENO < 7 and moved.
//  (3) KING:          +6 if castled; -2 if moved, but not castled
func Development(ctx context.Context, b *board.Board) eval.Pawns {
	pos := b.Position()
	own := b.Turn()
	opp := own.Opponent()

	mask := b.HasMoved(1000)

	pawns := -2 * eval.Pawns((pos.Piece(own, board.Knight)&^mask).PopCount()-(pos.Piece(opp, board.Knight)&^mask).PopCount())
	pawns -= 2 * eval.Pawns((pos.Piece(own, board.Bishop)&^mask).PopCount()-(pos.Piece(opp, board.Bishop)&^mask).PopCount())
	if b.FullMoves() < 7 {
		pawns -= 2 * eval.Pawns((pos.Piece(own, board.Rook)&mask).PopCount()-(pos.Piece(opp, board.Rook)&mask).PopCount())
		pawns -= 2 * eval.Pawns((pos.Piece(own, board.Queen)&mask).PopCount()-(pos.Piece(opp, board.Queen)&mask).PopCount())
	}

	pawns += king(b.HasCastled(own), (pos.Piece(own, board.King)&mask) != 0)
	pawns -= king(b.HasCastled(opp), (pos.Piece(opp, board.King)&mask) != 0)
	return pawns
}

func king(castled, moved bool) eval.Pawns {
	switch {
	case castled:
		return 6
	case moved:
		return -2
	default:
		return 0
	}
}

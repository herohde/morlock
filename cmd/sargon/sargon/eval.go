package sargon

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// Points implements the POINTS evaluation.
type Points struct {
	ply0         board.Color
	mtrl0, brdc0 eval.Pawns
}

func (p *Points) Reset(ctx context.Context, b *board.Board) {
	pins := FindKingQueenPins(b.Position())

	p.ply0 = b.Turn()
	p.mtrl0 = Material(ctx, b, pins, 0)
	p.brdc0 = BoardControl(ctx, b, pins, 0)
}

func (p *Points) Evaluate(ctx context.Context, b *board.Board) eval.Pawns {
	pins := FindKingQueenPins(b.Position())

	mtrl0, brdc0 := p.mtrl0, p.brdc0
	if b.Turn() != p.ply0 {
		mtrl0, brdc0 = -mtrl0, -brdc0
	}

	mtrl := Material(ctx, b, pins, mtrl0)
	brdc := BoardControl(ctx, b, pins, brdc0)
	return mtrl*4 + brdc
}

// Material: 1,3,3,5,9,10

// Mobility : WHITE moves squares - BLACK move squares.

// Pins? Exchange. 2x victim.
// Pins for K+Q

// Move list: has moved. Not moved.

// Points: MTRL * 4 + BRDC
//  - MTRL: MATERIAL (XCHNG w/ pins for K/Q) Material difference. max=30 from ply0
//  - BRDC: BOARD CONTROL:   #Sq target move difference max=6 from ply0
//       * PAWN: no bonus
//       * KNIGHT/BISHOP -2/+2 if not moved.
//       * ROOK/QUEEN -2/+2 if MOVENO < 7 and _moved_.
//       * KING +6 if castle, -6 if opp.; -2 / +2 opp if moved, but not castled
//     + ATTACKERS - DEFENDERS for each square.

// PIN: pieces pinned to K or Q w/ source of pin. Can attack source.
// ATTACK: attack lists w/ transparent attackers and w/o pinned pieces (except if source).

// XCHNG: exchange value.
//  initial defender x2 (<-- not in BYTE article), but perhaps only to allow RxB if defended. Not QxP or QxB. Seems
//  like a trick to cutoff futile exchanges that we can never bounce back from.
//  Use:  3/4 for PTSW2.

// BASELINE: POINTS with 0 as ply0 value. So still limited to +/- 30/6.

// POINTS:
// if LOSS
//   -0.5
//   Save greatest loss to PTSL
//   Save PTSCHK if piece first lost just moved.
// if WIN:
//   - Save greatest 2 wins as PTSW1, PTSW2.
//   - Use PTSW2 if moving piece was lost.

//  PSTL:  if >0 then -1
//   - adjustment: (PSTW1 + PTSW2 -1)/2 - PSTL.   (omit x-1/2 if PTSW2 == 0)

// Q: Moving piece lost flag? Prevent undefended exchange.

// Does not under-promote.

// Quiescece +1ply if in check.

// Material implements the MTRL heuristic, limited to +/- 30 relative to its ply0 value.
func Material(ctx context.Context, b *board.Board, pins Pins, ply0 eval.Pawns) eval.Pawns {
	pos := b.Position()
	turn := b.Turn()

	mtrl := eval.Material{}.Evaluate(ctx, b)

	var ptsl, ptsw1, ptsw2 eval.Pawns
	for sq := board.ZeroSquare; sq < board.NumSquares; sq++ {
		v := Exchange(pos, pins, turn.Opponent(), sq)
		switch {
		case v < ptsl:
			ptsl = v
		case ptsw1 < v:
			ptsw1, ptsw2 = v, ptsw1
		case ptsw2 < v:
			ptsw2 = v
		}
	}

	if last, ok := b.LastMove(); ok && pos.IsAttacked(turn.Opponent(), last.To) {
		// Use PTSW2 if moving piece is subject to capture. Assumed unguarded win.
		ptsw1, ptsw2 = ptsw2, 0
	}
	mtrl -= ptsl + (3*ptsw1)/4
	return eval.Limit(mtrl-ply0, 30)
}

// BoardControl implements the BRDC heuristic limited to +/- 6 relative to its ply0 value.
func BoardControl(ctx context.Context, b *board.Board, pins Pins, ply0 eval.Pawns) eval.Pawns {
	brdc := Development(ctx, b) + Mobility(ctx, b, pins)
	return eval.Limit(brdc-ply0, 6)
}

// Mobility implements the development aspects of the BRDC heuristic, without limit.
func Mobility(ctx context.Context, b *board.Board, pins Pins) eval.Pawns {
	pos := b.Position()
	turn := b.Turn()

	var pawns eval.Pawns
	for sq := board.ZeroSquare; sq < board.NumSquares; sq++ {
		if !pos.IsEmpty(sq) {
			continue
		}
		attackers := FindAttackers(pos, pins, sq)
		pawns += eval.Pawns(NumAttackers(attackers, turn) - NumAttackers(attackers, turn.Opponent()))
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

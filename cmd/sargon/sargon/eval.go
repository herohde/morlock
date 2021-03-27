package sargon

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

type Points struct {
}

func (p Points) Evaluate(ctx context.Context, pos *board.Position, turn board.Color) eval.Score {
	panic("implement me")
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

// Q: Moving piece lost flag?

// Does not under-promote.

// Quiescece +1ply if in check.

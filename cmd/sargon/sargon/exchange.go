package sargon

import (
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// Attacker represents a non-pinned attacker of some square. The attacker
// may potentially have others stacked behind it. For example, if we have
// Rook -> Queen -> Target then the Rook is "behind" the queen. The Rook can
// only attack after the Queen has attacked in an exchange.
type Attacker struct {
	Piece  board.Placement
	Behind *Attacker
}

// NumAttackers returns the number of attackers for the given side.
func NumAttackers(attackers []*Attacker, turn board.Color) int {
	count := 0
	for _, att := range attackers {
		if att.Piece.Color != turn {
			continue
		}
		for att != nil {
			count++
			att = att.Behind
		}
	}
	return count
}

// FindAttackers returns all direct and indirect attackers to a given square.
func FindAttackers(pos *board.Position, pins Pins, sq board.Square) []*Attacker {
	var ret []*Attacker
	for _, piece := range board.KingQueenRookKnightBishop {
		attackboard := board.Attackboard(pos.Rotated(), sq, piece)

		for side := board.ZeroColor; side < board.NumColors; side++ {
			bb := attackboard & pos.Piece(side, piece)
			for bb != 0 {
				from := bb.LastPopSquare()
				bb ^= board.BitMask(from)

				stack, ok := addAttackerStack(pos, pos.Rotated(), pins, side, piece, from, sq)
				if ok {
					ret = append(ret, stack)
				}
			}
		}
	}

	for side := board.ZeroColor; side < board.NumColors; side++ {
		bb := board.PawnCaptureboard(side.Opponent(), board.BitMask(sq)) & pos.Piece(side, board.Pawn)
		for bb != 0 {
			from := bb.LastPopSquare()
			bb ^= board.BitMask(from)

			stack, ok := addAttackerStack(pos, pos.Rotated(), pins, side, board.Pawn, from, sq)
			if ok {
				ret = append(ret, stack)
			}
		}
	}

	return ret
}

func addAttackerStack(pos *board.Position, r board.RotatedBitboard, pins Pins, side board.Color, piece board.Piece, from, target board.Square) (*Attacker, bool) {
	if list := pins[from]; len(list) > 1 || (len(list) == 1 && list[0] != target) {
		return nil, false // skip: attacker is pinned
	}

	ret := &Attacker{
		Piece: board.Placement{
			Piece:  piece,
			Color:  side,
			Square: from,
		},
	}

	next := r.Xor(from)

	bb := board.EmptyBitboard
	if board.IsSameRankOrFile(from, target) {
		attackboard := board.RookAttackboard(next, target) &^ board.RookAttackboard(r, target)
		bb = attackboard & (pos.Piece(side, board.Queen) | pos.Piece(side, board.Rook))
	} else if board.IsSameDiagonal(from, target) {
		attackboard := board.BishopAttackboard(next, target) &^ board.BishopAttackboard(r, target)
		bb = attackboard & (pos.Piece(side, board.Queen) | pos.Piece(side, board.Bishop))
	}

	if bb != 0 {
		from = bb.LastPopSquare()
		_, piece, _ = pos.Square(from)

		ret.Behind, _ = addAttackerStack(pos, next, pins, side, piece, from, target)
	}

	return ret, true
}

// Pins is a map of pinned squared to list of attackers of opposing side.
type Pins map[board.Square][]board.Square

// FindKingQueenPins returns pinned squares and their attackers. If there is only one
// attacker, the pinned piece can attack that one. It is possible for a pinned piece
// to protect both the King and Queen.
func FindKingQueenPins(pos *board.Position) Pins {
	var pins []eval.Pin
	for side := board.ZeroColor; side < board.NumColors; side++ {
		for _, piece := range board.KingQueen {
			pins = append(pins, eval.FindPins(pos, side, piece)...)
		}
	}

	ret := map[board.Square][]board.Square{}
	for _, pin := range pins {
		ret[pin.Pinned] = append(ret[pin.Pinned], pin.Attacker)
	}
	return ret
}

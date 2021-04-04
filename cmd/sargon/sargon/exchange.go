package sargon

import (
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"sort"
)

// Exchange computes the exchange value of the square, if populated.
func Exchange(pos *board.Position, pins Pins, side board.Color, sq board.Square) eval.Pawns {
	cur, piece, ok := pos.Square(sq)
	if !ok || piece == board.King {
		return 0 // empty square or King: no exchange value
	}

	all := FindAttackers(pos, pins, sq)
	defenders := findSide(all, cur)
	attackers := findSide(all, cur.Opponent())

	var residue eval.Pawns // gain of exchange from cur.Opponent point-of-view

	defender := eval.NominalValue(piece)
	for len(attackers) > 0 {
		attacker := attackers[0]
		attackers = attackers[1:]

		// Opposing side will attack, if undefended or not a loss.

		willAttack := len(defenders) == 0 || val(attacker) <= defender
		willAttack = willAttack || (len(attackers) > 0 && val(attacker)+val(attackers[0]) <= defender+val(defenders[0]))
		if !willAttack {
			break
		}

		residue += defender
		defender = val(attacker)

		// Swap roles

		attackers, defenders = defenders, attackers
		residue = -residue
		cur = cur.Opponent()
	}

	if cur == side {
		return -residue
	}
	return residue
}

func findSide(attackers []*Attacker, turn board.Color) []*Attacker {
	// (1) Project side

	var ret []*Attacker
	for _, att := range attackers {
		if att.Piece.Color == turn {
			ret = append(ret, att)
		}
	}

	// (2) Flatten into attack list in value order, while respecting the Behind relation.

	sort.Slice(ret, byValue(ret))
	for i := 0; i < len(ret); i++ {
		att := ret[i]
		if att.Behind == nil {
			continue
		}

		ret = append(ret, att.Behind)
		sort.Slice(ret[i+1:], byValue(ret[i+1:]))
	}
	return ret
}

func byValue(list []*Attacker) func(i, j int) bool {
	return func(i, j int) bool {
		return val(list[i]) < val(list[j])
	}
}

func val(att *Attacker) eval.Pawns {
	return eval.NominalValue(att.Piece.Piece)
}

// Attacker represents a non-pinned attacker of some square. The attacker
// may potentially have others stacked behind it. For example, if we have
// Rook -> Queen -> Target then the Rook is "behind" the queen. The Rook can
// only attack after the Queen has attacked in an exchange.
type Attacker struct {
	Piece  board.Placement
	Behind *Attacker
}

func (a *Attacker) String() string {
	return fmt.Sprintf("%v|%v", a.Piece, a.Behind)
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
	if piece == board.King {
		return ret, true // nobody can be behind the King in an exchange
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

package eval

import "github.com/herohde/morlock/pkg/board"

// Pin represents a pinned piece. A pinned piece cannot attack anything but
// the attacker itself, if the relative value of attacker/target is high enough.
type Pin struct {
	Attacker, Pinned, Target board.Square
}

// FindPins returns all pins targeting the given piece.
func FindPins(pos *board.Position, side board.Color, piece board.Piece) []Pin {
	var ret []Pin

	bb := pos.Piece(side, piece)
	for bb != 0 {
		target := bb.LastPopSquare()
		bb ^= board.BitMask(target)

		// (1) Rook/Queen pins

		rooks := board.RookAttackboard(pos.Rotated(), target)
		pins := rooks & pos.Color(side)
		for pins != 0 {
			pinned := pins.LastPopSquare()
			pins ^= board.BitMask(pinned)

			attackers := pos.Piece(side.Opponent(), board.Queen) | pos.Piece(side.Opponent(), board.Rook)

			candidate := (board.RookAttackboard(pos.Rotated().Xor(pinned), target) &^ rooks) & attackers
			if candidate != 0 {
				attacker := candidate.LastPopSquare()

				p := Pin{Attacker: attacker, Pinned: pinned, Target: target}
				ret = append(ret, p)
			}
		}

		// (1) Bishop/Queen pins

		bishops := board.BishopAttackboard(pos.Rotated(), target)
		pins = bishops & pos.Color(side)
		for pins != 0 {
			pinned := pins.LastPopSquare()
			pins ^= board.BitMask(pinned)

			attackers := pos.Piece(side.Opponent(), board.Queen) | pos.Piece(side.Opponent(), board.Bishop)

			candidate := (board.BishopAttackboard(pos.Rotated().Xor(pinned), target) &^ bishops) & attackers
			if candidate != 0 {
				attacker := candidate.LastPopSquare()

				p := Pin{Attacker: attacker, Pinned: pinned, Target: target}
				ret = append(ret, p)
			}
		}
	}

	return ret
}

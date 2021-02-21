package board

// Piece represents a chess piece (King, Pawn, etc) with no color. 3 bits.
type Piece uint8

const (
	NoPiece Piece = iota
	Pawn
	Bishop
	Knight
	Rook
	Queen
	King
)

func ParsePiece(r rune) (Piece, bool) {
	switch r {
	case 'p', 'P':
		return Pawn, true
	case 'b', 'B':
		return Bishop, true
	case 'n', 'N':
		return Knight, true
	case 'r', 'R':
		return Rook, true
	case 'q', 'Q':
		return Queen, true
	case 'k', 'K':
		return King, true
	default:
		return NoPiece, false
	}
}

func (p Piece) IsValid() bool {
	return Pawn <= p && p <= King
}

func (p Piece) String() string {
	switch p {
	case NoPiece:
		return " "
	case Pawn:
		return "p"
	case Bishop:
		return "b"
	case Knight:
		return "n"
	case Rook:
		return "r"
	case Queen:
		return "q"
	case King:
		return "k"
	default:
		return "?"
	}
}

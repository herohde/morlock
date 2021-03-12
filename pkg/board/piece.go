package board

// Piece represents a chess piece (King, Pawn, etc) with no color. Zero indicates "No Piece". 3 bits.
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

const (
	ZeroPiece Piece = 1
	NumPieces Piece = 7
)

var (
	KingQueenRookKnightBishop = []Piece{King, Queen, Rook, Knight, Bishop}
	QueenRookKnightBishop     = []Piece{Queen, Rook, Knight, Bishop}
	QueenRookKnightBishopPawn = []Piece{Queen, Rook, Knight, Bishop, Pawn}
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

// NominalValue returns the absolute nominal value in centi-pawns. The King
// has an arbitrary value of 100 pawns.
func (p Piece) NominalValue() Score {
	switch p {
	case Pawn:
		return 100
	case Bishop, Knight:
		return 300
	case Rook:
		return 500
	case Queen:
		return 900
	case King:
		return 10000
	default:
		return 0
	}
}

func (p Piece) String() string {
	switch p {
	case NoPiece:
		return "-"
	case Pawn:
		return "P"
	case Bishop:
		return "B"
	case Knight:
		return "N"
	case Rook:
		return "R"
	case Queen:
		return "Q"
	case King:
		return "K"
	default:
		return "?"
	}
}

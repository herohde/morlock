package board

import "fmt"

// MoveType indicates the type of move. The no-progress counter is reset with any non-Normal move.
type MoveType uint8

const (
	Normal    MoveType = iota
	Push               // Pawn move
	Jump               // Pawn 2-square move
	EnPassant          // Implicitly a pawn capture
	QueenSideCastle
	KingSideCastle
	Capture
	Promotion
	CapturePromotion
)

func (m MoveType) String() string {
	switch m {
	case Normal:
		return "Normal"
	case Push:
		return "Push"
	case Jump:
		return "Jump"
	case EnPassant:
		return "EnPassant"
	case QueenSideCastle:
		return "QueenSideCastle"
	case KingSideCastle:
		return "KingSideCastle"
	case Capture:
		return "Capture"
	case Promotion:
		return "Promotion"
	case CapturePromotion:
		return "CapturePromotion"
	default:
		return "?"
	}
}

// TODO(herohde) 2/21/2021: add remarks, like "dubious", to represent standard notation?

// Move represents a not-necessarily legal move along with contextual metadata. 64bits.
type Move struct {
	Type      MoveType
	From, To  Square
	Piece     Piece // moved piece
	Promotion Piece // desired piece for promotion, if any.
	Capture   Piece // captured piece, if any. Not set if EnPassant.
	Score     Score
}

// ParseMove parses a move in pure algebraic coordinate notation, such as "a2a4" or "a7a8q".
// The parsed move does not contain contextual information like castling or en passant.
func ParseMove(str string) (Move, error) {
	runes := []rune(str)

	if len(runes) < 4 || len(runes) > 5 {
		return Move{}, fmt.Errorf("invalid move: '%v'", str)
	}

	from, err := ParseSquare(runes[0], runes[1])
	if err != nil {
		return Move{}, fmt.Errorf("invalid from: '%v': %v", str, err)
	}
	to, err := ParseSquare(runes[2], runes[3])
	if err != nil {
		return Move{}, fmt.Errorf("invalid to: '%v': %v", str, err)
	}

	if len(runes) == 5 {
		promo, ok := ParsePiece(runes[4])
		if !ok || promo == Pawn || promo == King {
			return Move{}, fmt.Errorf("invalid promotion: '%v'", str)
		}
		return Move{From: from, To: to, Promotion: promo}, nil
	}

	return Move{From: from, To: to}, nil
}

func (m Move) Equals(o Move) bool {
	return m.From == o.From && m.To == o.To && m.Promotion == o.Promotion
}

func (m Move) String() string {
	switch m.Type {
	case Promotion:
		return fmt.Sprintf("%v-%v=%v", m.From, m.To, m.Promotion)
	case CapturePromotion:
		return fmt.Sprintf("%v*%v=%v", m.From, m.To, m.Promotion)
	case EnPassant:
		return fmt.Sprintf("%v*%v e.p.", m.From, m.To)
	case KingSideCastle:
		return fmt.Sprintf("0-0")
	case QueenSideCastle:
		return fmt.Sprintf("0-0-0")
	case Capture:
		return fmt.Sprintf("%v%v*%v", ignorePawn(m.Piece), m.From, m.To)
	default:
		return fmt.Sprintf("%v%v-%v", ignorePawn(m.Piece), m.From, m.To)
	}
}

func ignorePawn(piece Piece) string {
	if piece == Pawn {
		return ""
	}
	return piece.String()
}

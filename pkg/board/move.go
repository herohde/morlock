package board

import (
	"fmt"
	"github.com/seekerror/stdlib/pkg/util/slicex"
	"strings"
)

// MoveType indicates the type of move. The no-progress counter is reset with any non-Normal move.
type MoveType uint8

const (
	Normal    MoveType = 1 + iota
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

// IsInvalid true iff the move is of invalid type. Convenience function.
func (m Move) IsInvalid() bool {
	return m.Type == 0
}

// IsCapture returns true iff the move is a Capture or CapturePromotion. Convenience function.
func (m Move) IsCapture() bool {
	return m.Type == CapturePromotion || m.Type == Capture
}

// IsCaptureOrEnPassant returns true iff the move is a Capture, CapturePromotion or EnPassant. Convenience function.
func (m Move) IsCaptureOrEnPassant() bool {
	return m.Type == CapturePromotion || m.Type == Capture || m.Type == EnPassant
}

// IsPromotion returns true iff the move is a Promotion or CapturePromotion. Convenience function.
func (m Move) IsPromotion() bool {
	return m.Type == CapturePromotion || m.Type == Promotion
}

// IsUnderPromotion returns true iff the move is a Promotion, but not to a Queen. Convenience function.
func (m Move) IsUnderPromotion() bool {
	return m.IsPromotion() && m.Promotion != Queen
}

// IsNotUnderPromotion returns false if under-promotion. Convenience function for move selection.
func (m Move) IsNotUnderPromotion() bool {
	return !m.IsUnderPromotion()
}

// IsCastle returns true iff the move is a KingSideCastle or QueenSideCastle. Convenience function.
func (m Move) IsCastle() bool {
	return m.Type == KingSideCastle || m.Type == QueenSideCastle
}

// EnPassantTarget return the e.p target square, if a Jump move. For e2-e4, it turns e3.
func (m Move) EnPassantTarget() (Square, bool) {
	if m.Type != Jump {
		return 0, false
	}

	if m.To.Rank() == Rank4 { // White
		return NewSquare(m.To.File(), Rank3), true
	} else {
		return NewSquare(m.To.File(), Rank6), true
	}
}

// EnPassantCapture return the e.p capture square, if a EnPassant move. For d4*e3 e.p, it turns e4.
func (m Move) EnPassantCapture() (Square, bool) {
	if m.Type != EnPassant {
		return 0, false
	}

	if m.To.Rank() == Rank3 { // Black
		return NewSquare(m.To.File(), Rank4), true
	} else {
		return NewSquare(m.To.File(), Rank5), true
	}
}

// CastlingRookMove returns the implicit rook move (from, to), if a KingSideCastle or QueenSideCastle move.
func (m Move) CastlingRookMove() (Square, Square, bool) {
	switch {
	case m.Type == KingSideCastle && m.From == E1:
		return H1, F1, true
	case m.Type == QueenSideCastle && m.From == E1:
		return A1, D1, true
	case m.Type == KingSideCastle && m.From == E8:
		return H8, F8, true
	case m.Type == QueenSideCastle && m.From == E8:
		return A8, D8, true
	default:
		return 0, 0, false
	}
}

// CastlingRightsLost returns the castling rights that are definitely not present after this move.
// If king moves, rights are lost. Ditto if rook moves or is captured.
func (m Move) CastlingRightsLost() Castling {
	switch {
	case m.From == E1:
		return WhiteKingSideCastle | WhiteQueenSideCastle
	case m.From == A1 || m.To == A1:
		return WhiteQueenSideCastle
	case m.From == H1 || m.To == H1:
		return WhiteKingSideCastle
	case m.From == E8:
		return BlackKingSideCastle | BlackQueenSideCastle
	case m.From == A8 || m.To == A8:
		return BlackQueenSideCastle
	case m.From == H8 || m.To == H8:
		return BlackKingSideCastle
	default:
		return NoCastlingRights
	}
}

func (m Move) Equals(o Move) bool {
	return m.From == o.From && m.To == o.To && m.Promotion == o.Promotion
}

func (m Move) String() string {
	if m.IsInvalid() {
		return "invalid"
	}

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

// MovePredicateFn is a move predicate.
type MovePredicateFn func(move Move) bool

// FindMoves returns moves that satisfy a predicate from a list of moves.
func FindMoves(moves []Move, fn MovePredicateFn) []Move {
	return slicex.MapIf(moves, func(m Move) (Move, bool) {
		return m, fn(m)
	})
}

// PrintMoves prints a list of moves.
func PrintMoves(list []Move) string {
	return FormatMoves(list, func(m Move) string {
		return m.String()
	})
}

// FormatMoves formats a list of moves.
func FormatMoves(list []Move, fn func(Move) string) string {
	var ret []string
	for _, m := range list {
		ret = append(ret, fn(m))
	}
	return strings.Join(ret, " ")
}

func ignorePawn(piece Piece) string {
	if piece == Pawn || piece == NoPiece {
		return ""
	}
	return piece.String()
}

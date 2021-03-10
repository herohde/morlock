package board

import "strings"

// Castling represents the set of castling rights. 4 bits.
type Castling uint8

const (
	WhiteKingSideCastle Castling = 1 << iota
	WhiteQueenSideCastle
	BlackKingSideCastle
	BlackQueenSideCastle
)

const (
	NoCastlingRights  = Castling(0)
	FullCastingRights = WhiteKingSideCastle | WhiteQueenSideCastle | BlackKingSideCastle | BlackQueenSideCastle

	ZeroCastling = NoCastlingRights
	NumCastling  = FullCastingRights + 1
)

// IsAllowed returns true iff all the given rights are allowed.
func (c Castling) IsAllowed(right Castling) bool {
	return c&right != 0
}

func (c Castling) String() string {
	if c == 0 {
		return "-"
	}

	var sb strings.Builder
	if c.IsAllowed(WhiteKingSideCastle) {
		sb.WriteString("K")
	}
	if c.IsAllowed(WhiteQueenSideCastle) {
		sb.WriteString("Q")
	}
	if c.IsAllowed(BlackKingSideCastle) {
		sb.WriteString("k")
	}
	if c.IsAllowed(BlackQueenSideCastle) {
		sb.WriteString("q")
	}
	return sb.String()
}

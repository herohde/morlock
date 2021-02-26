// Package fen contains utilities for read and writing positions in FEN notation.
package fen

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/herohde/morlock/pkg/board"
)

const (
	Initial = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
)

// Decode returns a new position and game status from a FEN description.
//
// Example:
//   "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
func Decode(fen string) (*board.Position, board.Color, int, int, error) {
	// A FEN record contains six fields. The separator between fields is a
	// space. The fields are:

	parts := strings.Split(strings.TrimSpace(fen), " ")
	if len(parts) != 6 {
		return nil, 0, 0, 0, fmt.Errorf("invalid number of sections in FEN: '%v'", fen)
	}

	// (1) Piece placement (from white's perspective). Each rank is described,
	// starting with rank 8 and ending with rank 1; within each rank, the
	// contents of each square are described from file a through file h.

	var pieces []board.Placement

	sq := board.A8
	for _, r := range []rune(parts[0]) {
		switch {
		case r == '/':
			// "/" separate ranks. Cosmetic.

		case unicode.IsDigit(r):
			// Blank squares are noted using digits 1 through 8 (the number of blank squares).

			sq -= board.Square(r - '0')

		case unicode.IsLetter(r):
			// Following the Standard Algebraic Notation (SAN), each piece is -
			// identified by a single letter taken from the standard English names -
			// (pawn = "P", knight = "N", bishop = "B", rook = "R", queen = "Q" and -
			// king = "K")[1]. White pieces are designated using upper-case letters -
			// ("PNBRQK") while Black take lowercase ("pnbrqk").

			color, piece, ok := parsePiece(r)
			if !ok {
				return nil, 0, 0, 0, fmt.Errorf("invalid piece '%v' in FEN: '%v'", r, fen)
			}
			pieces = append(pieces, board.Placement{Square: sq, Color: color, Piece: piece})
			sq--

		default:
			return nil, 0, 0, 0, fmt.Errorf("invalid character in FEN: '%v'", fen)
		}
	}
	if sq+1 != board.H1 {
		return nil, 0, 0, 0, fmt.Errorf("invalid number of squares in FEN: '%v'", fen)
	}

	// (2) Active color. "w" means white moves next, "b" means black.

	active, ok := parseColor(parts[1])
	if !ok {
		return nil, 0, 0, 0, fmt.Errorf("invalid active color in FEN: '%v'", fen)
	}

	// (3) Castling availability. If neither side can castle, this is
	// "-". Otherwise, this has one or more letters: "K" (White can castle
	// kingside), "Q" (White can castle queenside), "k" (Black can castle
	// kingside), and/or "q" (Black can castle queenside).

	castling, ok := parseCastling(parts[2])
	if !ok {
		return nil, 0, 0, 0, fmt.Errorf("invalid castling in FEN: '%v'", fen)
	}

	// (4) En passant target square in algebraic notation. If there's no en
	// passant target square, this is "-". If a pawn has just made a
	// 2-square move, this is the position "behind" the pawn.

	var ep board.Square
	if parts[3] != "-" {
		sq, err := board.ParseSquareStr(parts[3])
		if err != nil {
			return nil, 0, 0, 0, fmt.Errorf("invalid en passant in FEN: '%v'", fen)
		}
		ep = sq
	}

	// (5) Halfmove clock: This is the number of halfmoves since the last pawn
	// advance or capture. This is used to determine if a draw can be
	// claimed under the fifty move rule.

	np, err := strconv.Atoi(parts[4])
	if err != nil || np < 0 {
		return nil, 0, 0, 0, fmt.Errorf("invalid halfmove in FEN: '%v'", fen)
	}

	// (6) Fullmove number: The number of the full move. It starts at 1, and is
	// incremented after Black's move.

	fm, err := strconv.Atoi(parts[5])
	if err != nil || fm < 0 {
		return nil, 0, 0, 0, fmt.Errorf("invalid full moves in FEN: '%v'", fen)
	}

	pos, _ := board.NewPosition(pieces, castling, ep)
	return pos, active, np, fm, nil
}

// Encode encodes the position and game data in FEN notation.
func Encode(pos *board.Position, c board.Color, noprogress, fullmoves int) string {
	var sb strings.Builder
	for r := board.ZeroRank; r < board.NumRanks; r++ {
		blanks := 0
		for f := board.ZeroFile; f < board.NumFiles; f++ {
			color, piece, ok := pos.Square(board.NewSquare(board.NumFiles-f-1, board.NumRanks-r-1))
			if !ok {
				blanks++
				continue
			}

			if blanks > 0 {
				sb.WriteString(strconv.Itoa(blanks))
				blanks = 0
			}

			sb.WriteRune(printPiece(color, piece))
		}

		if blanks > 0 {
			sb.WriteString(strconv.Itoa(blanks))
			blanks = 0
		}

		if r < board.NumRanks-1 {
			sb.WriteString("/")
		}
	}

	turn := printColor(c)
	castling := printCastling(pos.Castling())

	ep := "-"
	if sq, ok := pos.EnPassant(); ok {
		ep = sq.String()
	}

	return fmt.Sprintf("%v %v %v %v %v %v", sb.String(), turn, castling, ep, noprogress, fullmoves)
}

func parseCastling(str string) (board.Castling, bool) {
	var ret board.Castling

	if str == "-" {
		return ret, true
	}
	for _, r := range []rune(str) {
		switch r {
		case 'K':
			ret |= board.WhiteKingSideCastle
		case 'Q':
			ret |= board.WhiteQueenSideCastle
		case 'k':
			ret |= board.BlackKingSideCastle
		case 'q':
			ret |= board.BlackQueenSideCastle
		default:
			return 0, false
		}
	}
	return ret, true
}

func printCastling(c board.Castling) string {
	if c == 0 {
		return "-"
	}

	ret := ""
	if c.IsAllowed(board.WhiteKingSideCastle) {
		ret += "K"
	}
	if c.IsAllowed(board.WhiteQueenSideCastle) {
		ret += "Q"
	}
	if c.IsAllowed(board.BlackKingSideCastle) {
		ret += "k"
	}
	if c.IsAllowed(board.BlackQueenSideCastle) {
		ret += "q"
	}
	return ret
}

func parseColor(str string) (board.Color, bool) {
	switch str {
	case "w", "W":
		return board.White, true
	case "b", "B":
		return board.Black, true
	default:
		return 0, false
	}
}

func printColor(c board.Color) string {
	if c == board.White {
		return "w"
	}
	return "b"
}

func parsePiece(r rune) (board.Color, board.Piece, bool) {
	switch r {
	case 'P':
		return board.White, board.Pawn, true
	case 'B':
		return board.White, board.Bishop, true
	case 'N':
		return board.White, board.Knight, true
	case 'R':
		return board.White, board.Rook, true
	case 'Q':
		return board.White, board.Queen, true
	case 'K':
		return board.White, board.King, true

	case 'p':
		return board.Black, board.Pawn, true
	case 'b':
		return board.Black, board.Bishop, true
	case 'n':
		return board.Black, board.Knight, true
	case 'r':
		return board.Black, board.Rook, true
	case 'q':
		return board.Black, board.Queen, true
	case 'k':
		return board.Black, board.King, true

	default:
		return 0, 0, false
	}
}

func printPiece(c board.Color, p board.Piece) rune {
	if c == board.White {
		switch p {
		case board.Pawn:
			return 'P'
		case board.Bishop:
			return 'B'
		case board.Knight:
			return 'N'
		case board.Rook:
			return 'R'
		case board.Queen:
			return 'Q'
		case board.King:
			return 'K'
		default:
			return '?'
		}
	}

	switch p {
	case board.Pawn:
		return 'p'
	case board.Bishop:
		return 'b'
	case board.Knight:
		return 'n'
	case board.Rook:
		return 'r'
	case board.Queen:
		return 'q'
	case board.King:
		return 'k'
	default:
		return '?'
	}
}

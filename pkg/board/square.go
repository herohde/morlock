package board

import "fmt"

// Square represents a square on the board, ordered H1=0, G1=1 .., A8=63. This numbering
// matches a 64-bit interpretation as a bitboard:
//
//  A8 = 63, B8 = 62, C8 = 61, D8 = 60, E8 = 59, F8 = 58, G8 = 57, H8 = 56,
//  A7 = 55, B7 = 54, C7 = 53, D7 = 52, E7 = 51, F7 = 50, G7 = 49, H7 = 48,
//  A6 = 47, B6 = 46, C6 = 45, D6 = 44, E6 = 43, F6 = 42, G6 = 41, H6 = 40,
//  A5 = 39, B5 = 38, C5 = 37, D5 = 36, E5 = 35, F5 = 34, G5 = 33, H5 = 32,
//  A4 = 31, B4 = 30, C4 = 29, D4 = 28, E4 = 27, F4 = 26, G4 = 25, H4 = 24,
//  A3 = 23, B3 = 22, C3 = 21, D3 = 20, E3 = 19, F3 = 18, G3 = 17, H3 = 16,
//  A2 = 15, B2 = 14, C2 = 13, D2 = 12, E2 = 11, F2 = 10, G2 =  9, H2 =  8,
//  A1 =  7, B1 =  6, C1 =  5, D1 =  4, E1 =  3, F1 =  2, G1 =  1, H1 =  0
//
// A square is a bit-index into the bitboard layout. 6 bits.
type Square uint8

const (
	H1 Square = iota
	G1
	F1
	E1
	D1
	C1
	B1
	A1

	H2
	G2
	F2
	E2
	D2
	C2
	B2
	A2

	H3
	G3
	F3
	E3
	D3
	C3
	B3
	A3

	H4
	G4
	F4
	E4
	D4
	C4
	B4
	A4

	H5
	G5
	F5
	E5
	D5
	C5
	B5
	A5

	H6
	G6
	F6
	E6
	D6
	C6
	B6
	A6

	H7
	G7
	F7
	E7
	D7
	C7
	B7
	A7

	H8
	G8
	F8
	E8
	D8
	C8
	B8
	A8
)

// Iteration helpers to enable "for i := ZeroSquare; i<NumSquares; i++".
const (
	ZeroSquare Square = 0
	NumSquares Square = 64
)

func NewSquare(f File, r Rank) Square {
	return ((Square(r) & 0x7) << 3) | (Square(f) & 0x7)
}

func ParseSquare(f, r rune) (Square, error) {
	file, ok := ParseFile(f)
	if !ok {
		return 0, fmt.Errorf("invalid file: %v", f)
	}
	rank, ok := ParseRank(r)
	if !ok {
		return 0, fmt.Errorf("invalid rank: %v", r)
	}
	return NewSquare(file, rank), nil
}

func ParseSquareStr(str string) (Square, error) {
	runes := []rune(str)
	if len(runes) != 2 {
		return 0, fmt.Errorf("invalid square: %v", str)
	}
	return ParseSquare(runes[0], runes[1])
}

func (s Square) IsValid() bool {
	return s <= A8
}

func (s Square) Rank() Rank {
	return Rank((s >> 3) & 0x7)
}

func (s Square) File() File {
	return File(s & 0x7)
}

func (s Square) String() string {
	return fmt.Sprintf("%v%v", s.File(), s.Rank())
}

// Rank represents a chess board rank from Rank1=0, ..Rank8=7. 3bits.
type Rank uint8

const (
	Rank1 Rank = iota
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
)

const (
	ZeroRank Rank = 0
	NumRanks Rank = 8
)

func ParseRank(r rune) (Rank, bool) {
	switch r {
	case '1':
		return Rank1, true
	case '2':
		return Rank2, true
	case '3':
		return Rank3, true
	case '4':
		return Rank4, true
	case '5':
		return Rank5, true
	case '6':
		return Rank6, true
	case '7':
		return Rank7, true
	case '8':
		return Rank8, true
	default:
		return 0, false
	}
}

func (r Rank) IsValid() bool {
	return r <= Rank8
}

func (r Rank) V() int {
	return int(r)
}

func (r Rank) String() string {
	switch r {
	case Rank1:
		return "1"
	case Rank2:
		return "2"
	case Rank3:
		return "3"
	case Rank4:
		return "4"
	case Rank5:
		return "5"
	case Rank6:
		return "6"
	case Rank7:
		return "7"
	case Rank8:
		return "8"
	default:
		return "?"
	}
}

// File represents a chess board file from FileH=0, ..FileA=7. The numbering is reversed
// to match Square. 3bits.
type File uint8

const (
	FileH File = iota
	FileG
	FileF
	FileE
	FileD
	FileC
	FileB
	FileA
)

const (
	ZeroFile File = 0
	NumFiles File = 8
)

func ParseFile(r rune) (File, bool) {
	switch r {
	case 'a', 'A':
		return FileA, true
	case 'b', 'B':
		return FileB, true
	case 'c', 'C':
		return FileC, true
	case 'd', 'D':
		return FileD, true
	case 'e', 'E':
		return FileE, true
	case 'f', 'F':
		return FileF, true
	case 'g', 'G':
		return FileG, true
	case 'h', 'H':
		return FileH, true
	default:
		return 0, false
	}
}

func (f File) IsValid() bool {
	return f <= FileA
}

func (f File) V() int {
	return int(f)
}

func (f File) String() string {
	switch f {
	case FileA:
		return "A"
	case FileB:
		return "B"
	case FileC:
		return "C"
	case FileD:
		return "D"
	case FileE:
		return "E"
	case FileF:
		return "F"
	case FileG:
		return "G"
	case FileH:
		return "H"
	default:
		return "?"
	}
}

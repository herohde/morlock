package board

import (
	"fmt"
	"math/bits"
	"strings"
)

// Bitboard is a bit-wise representation of the chess board. Each bit represents the appearance
// of some piece on that square. (bit 63 = A8 and bit 0 = H1). It relies on CPU-support for
// certain operations, such as popcount and bitscan.
type Bitboard uint64

const (
	EmptyBitboard Bitboard = 0
)

func (b Bitboard) IsSet(sq Square) bool {
	return b&BitMask(sq) != 0
}

// PopCount returns the population count of the bitboard, i.e., number of 1s.
func (b Bitboard) PopCount() int {
	return bits.OnesCount64(uint64(b))
}

// LastPopSquare returns the index of the least-significant 1. Returns 64 if zero.
func (b Bitboard) LastPopSquare() Square {
	return Square(bits.TrailingZeros64(uint64(b)))
}

func (b Bitboard) String() string {
	var sb strings.Builder
	for i := ZeroSquare; i < NumSquares; i++ {
		if i != 0 && i%8 == 0 {
			sb.WriteRune('/')
		}
		if b.IsSet(NumSquares - 1 - i) {
			sb.WriteRune('X')
		} else {
			sb.WriteRune('-')
		}
	}
	return sb.String()
}

// BitMask returns a bitboard with the given square populated.
func BitMask(sq Square) Bitboard {
	return Bitboard(1 << sq)
}

// BitRank returns a bitboard for the given rank.
func BitRank(r Rank) Bitboard {
	return Bitboard(0xff << (r << 3))
}

// BitFile returns a bitboard for the given file.
func BitFile(f File) Bitboard {
	return Bitboard(0x0101010101010101 << f)
}

// PawnCaptureboard returns all potential pawn captures for the given color.
func PawnCaptureboard(c Color, pawns Bitboard) Bitboard {
	if c == White {
		return ((pawns << 9) &^ BitFile(FileH)) | ((pawns << 7) &^ BitFile(FileA))
	} else {
		return ((pawns >> 9) &^ BitFile(FileA)) | ((pawns >> 7) &^ BitFile(FileH))
	}
}

// PawnMoveboard returns all potential pawn sigle-step moves for the given color.
func PawnMoveboard(all Bitboard, c Color, pawns Bitboard) Bitboard {
	if c == White {
		return (pawns << 8) & ^all
	} else {
		return (pawns >> 8) & ^all
	}
}

// PawnPromotionRank returns the mask of the promotion rank for the given color, i.e.,
// Rank8 for White or Rank1 for Black.
func PawnPromotionRank(c Color) Bitboard {
	if c == White {
		return BitRank(Rank8)
	} else {
		return BitRank(Rank1)
	}
}

// PawnJumpRank returns the mask of the target rank for jump moves for the given color,
// i.e., Rank4 for White or Rank5 for Black.
func PawnJumpRank(c Color) Bitboard {
	if c == White {
		return BitRank(Rank4)
	} else {
		return BitRank(Rank5)
	}
}

// Attackboard returns all potential moves/attacks for an officer (= non-Pawn) at the given square.
func Attackboard(bb RotatedBitboard, sq Square, piece Piece) Bitboard {
	switch piece {
	case King:
		return KingAttackboard(sq)
	case Queen:
		return QueenAttackboard(bb, sq)
	case Rook:
		return RookAttackboard(bb, sq)
	case Bishop:
		return BishopAttackboard(bb, sq)
	case Knight:
		return KnightAttackboard(sq)
	default:
		panic("invalid piece")
	}
}

// KingAttackboard returns all potential moves/attacks for a King at the given square.
func KingAttackboard(sq Square) Bitboard {
	return king[sq]
}

var king [NumSquares]Bitboard

func init() {
	for sq := ZeroSquare; sq < NumSquares; sq++ {
		// Build mask w/ crop: x -> xxx -> xxx/xxx/xxx -> xxx/x-x/xxx

		tmp := BitMask(sq)
		tmp |= ((tmp << 1) &^ BitFile(FileH)) | ((tmp >> 1) &^ BitFile(FileA))
		tmp |= tmp<<8 | tmp>>8
		tmp = tmp &^ BitMask(sq)

		king[sq] = tmp
	}
}

// KnighAttackboard returns all potential moves/attacks for a Knight at the given square.
func KnightAttackboard(sq Square) Bitboard {
	return knight[sq]
}

var knight [NumSquares]Bitboard

func init() {
	for sq := ZeroSquare; sq < NumSquares; sq++ {
		// Build mask w/ crop: x-x + x----x -> --x-x--/-x---x-/-------/-x---x-/--x-x--

		one := ((BitMask(sq) << 1) &^ BitFile(FileH)) | ((BitMask(sq) >> 1) &^ BitFile(FileA))
		two := ((BitMask(sq) << 2) &^ (BitFile(FileG) | BitFile(FileH))) | ((BitMask(sq) >> 2) &^ (BitFile(FileA) | BitFile(FileB)))
		tmp := one<<16 | one>>16 | two<<8 | two>>8

		knight[sq] = tmp
	}
}

// RotatedBitboard represents the piece-agnostic population of the board as so-called "rotated bitboards".
// It is designed to map files/diagonals into adjacent memory cells. It is conceptually simpler to
// view the transformations as rotations, but we are free to "shuffle" the files/diagonals as we
// please: the 'rot90' is really a flip. In the diagonal case we have to hold additional information
// about the length and offset of the desired diagonal, since that information is not constant.
type RotatedBitboard struct {
	rot, rot90, rot45L, rot45R Bitboard
}

func NewRotatedBitboard(bb Bitboard) RotatedBitboard {
	var ret RotatedBitboard
	for sq := ZeroSquare; sq < NumSquares; sq++ {
		if bb.IsSet(sq) {
			ret = ret.Xor(sq)
		}
	}
	return ret
}

// Mask returns the bitboard mask (in normal orientation).
func (r RotatedBitboard) Mask() Bitboard {
	return r.rot
}

// Xor returns the rotated bitboard xor the square mask.
func (r RotatedBitboard) Xor(sq Square) RotatedBitboard {
	return RotatedBitboard{
		rot:    r.rot ^ BitMask(sq),
		rot90:  r.rot90 ^ BitMask(rot90[sq]),
		rot45L: r.rot45L ^ BitMask(rot45L[sq]),
		rot45R: r.rot45R ^ BitMask(rot45R[sq]),
	}
}

func (r RotatedBitboard) String() string {
	return fmt.Sprintf("%v [rot90=%v, rot45L=%v, rot45R=%v]", r.rot, r.rot90, r.rot45L, r.rot45R)
}

const (
	// numStates is the maximum number of states on any rotation line (horizontal, vertical, diagonal).
	numStates = 256
)

// rot90 represents the 90 degree "rotated" square.
//
// 63 62 61 60 59 58 57 56          63 55 47 39 31 23 15  7
// 55 54 53 52 51 50 49 48          62 54 46 38 30 22 14  6
// 47 46 45 44 43 42 41 40  rot90   61 53 45 37 29 21 13  5
// 39 38 37 36 35 34 33 32 |------> 60 52 44 36 28 20 12  4
// 31 30 29 28 27 26 25 24          59 51 43 35 27 19 11  3
// 23 22 21 20 19 18 17 16          58 50 42 34 26 18 10  2
// 15 14 13 12 11 10  9  8          57 49 41 33 25 17  9  1
//  7  6  5  4  3  2  1  0          56 48 40 32 24 16  8  0
//
// We know that the mask is 0xff and the offset is file<<3.

var rot90 = [NumSquares]Square{
	0, 8, 16, 24, 32, 40, 48, 56,
	1, 9, 17, 25, 33, 41, 49, 57,
	2, 10, 18, 26, 34, 42, 50, 58,
	3, 11, 19, 27, 35, 43, 51, 59,
	4, 12, 20, 28, 36, 44, 52, 60,
	5, 13, 21, 29, 37, 45, 53, 61,
	6, 14, 22, 30, 38, 46, 54, 62,
	7, 15, 23, 31, 39, 47, 55, 63,
}

// RookAttackboard returns all potential moves/attacks for a Rook at the given square.
func RookAttackboard(bb RotatedBitboard, sq Square) Bitboard {
	rank := bb.rot >> (sq.Rank() << 3) & 0xff
	file := bb.rot90 >> (sq.File() << 3) & 0xff
	return rookrank[sq][rank] | rookfile[sq][file]
}

var (
	rookrank [NumSquares][numStates]Bitboard // (pos, rank state) -> bitboard
	rookfile [NumSquares][numStates]Bitboard // (pos, file state) -> bitboard
)

func init() {
	// Build mask by raytracing each direction.
	//
	// For example,
	//    Rook:    --R-----  (= Rook on index 2 of rank/file)
	//    State:   -XX---X-  (= Pieces on Rank/File)
	//    Attack:  -X-XXXX-  (= Rook moves/attacks)
	//
	// Attackboard is then obtained by shifting to the actual rank/file of the position. We
	// could store the table more compactly, if we were willing to shift the result on lookup.

	for sq := ZeroSquare; sq < NumSquares; sq++ {
		for state := EmptyBitboard; state < numStates; state++ {
			tmp := EmptyBitboard

			// Right: R--->X
			for i := Square(sq.File()) + 1; i < 8; i++ {
				tmp |= BitMask(i + Square(sq.Rank()<<3))
				if BitMask(i)&state != 0 {
					break
				}
			}

			// Left: X<-R
			for i := int(sq.File()) - 1; i > -1; i-- {
				tmp |= BitMask(Square(i) + Square(sq.Rank()<<3))
				if BitMask(Square(i))&state != 0 {
					break
				}
			}

			rookrank[sq][state] = tmp
		}
	}

	for sq := ZeroSquare; sq < NumSquares; sq++ {
		for state := EmptyBitboard; state < numStates; state++ {
			tmp := EmptyBitboard

			// Down: R-->X (rot90)
			for i := Square(sq.Rank()) + 1; i < 8; i++ {
				tmp |= BitMask(Square(sq.File()) + i<<3)
				if BitMask(i)&state != 0 {
					break
				}
			}

			// Up: X<-R (rot90)
			for i := int(sq.Rank()) - 1; i > -1; i-- {
				tmp |= BitMask(Square(sq.File()) + Square(i<<3))
				if BitMask(Square(i))&state != 0 {
					break
				}
			}

			rookfile[sq][state] = tmp
		}
	}
}

// rot45L represents the 45 degree clockwise rotated square.
//
// 63 62 61 60 59 58 57 56          35 42 48 53 57 60 62 63
// 55 54 53 52 51 50 49 48          27 34 41 47 52 56 59 61
// 47 46 45 44 43 42 41 40  rot45L  20 26 33 40 46 51 55 58
// 39 38 37 36 35 34 33 32 |------> 14 19 25 32 39 45 50 54
// 31 30 29 28 27 26 25 24           9 13 18 24 31 38 44 49
// 23 22 21 20 19 18 17 16           5  8 12 17 23 30 37 43
// 15 14 13 12 11 10  9  8           2  4  7 11 16 22 29 36
//  7  6  5  4  3  2  1  0           0  1  3  6 10 15 21 28
//
// 63 62 61 60 59 58 57 56           8  7  6  5  4  3  2  1           255 127  63  31  15   7   3   1
// 55 54 53 52 51 50 49 48           7  8  7  6  5  4  3  2           127 255 127  63  31  15   7   3
// 47 46 45 44 43 42 41 40  len45L   6  7  8  7  6  5  4  3  2^len-1   63 127 255 127  63  31  15   7
// 39 38 37 36 35 34 33 32 |------>  5  6  7  8  7  6  5  4 |------->  31  63 127 255 127  63  31  15
// 31 30 29 28 27 26 25 24           4  5  6  7  8  7  6  5            15  31  63 127 255 127  63  31
// 23 22 21 20 19 18 17 16           3  4  5  6  7  8  7  6             7  15  31  63 127 255 127  63
// 15 14 13 12 11 10  9  8           2  3  4  5  6  7  8  7             3   7  15  31  63 127 255 127
//  7  6  5  4  3  2  1  0           1  2  3  4  5  6  7  8             1   3   7  15  31  63 127 255
//
// However, since we are really only interested in the bitmask, we only store
// the composition, mask45L := len45L o 2^len-1.
//
// 63 62 61 60 59 58 57 56          28 36 43 49 54 58 61 63
// 55 54 53 52 51 50 49 48          21 28 36 43 49 54 58 61
// 47 46 45 44 43 42 41 40  off45L  15 21 28 36 43 49 54 58
// 39 38 37 36 35 34 33 32 |------> 10 15 21 28 36 43 49 54
// 31 30 29 28 27 26 25 24           6 10 15 21 28 36 43 49
// 23 22 21 20 19 18 17 16           3  6 10 15 21 28 36 43
// 15 14 13 12 11 10  9  8           1  3  6 10 15 21 28 36
//  7  6  5  4  3  2  1  0           0  1  3  6 10 15 21 28

var rot45L = [NumSquares]Square{
	28, 21, 15, 10, 6, 3, 1, 0,
	36, 29, 22, 16, 11, 7, 4, 2, // hard-to-find  bug: 36 was 35
	43, 37, 30, 23, 17, 12, 8, 5,
	49, 44, 38, 31, 24, 18, 13, 9,
	54, 50, 45, 39, 32, 25, 19, 14,
	58, 55, 51, 46, 40, 33, 26, 20,
	61, 59, 56, 52, 47, 41, 34, 27,
	63, 62, 60, 57, 53, 48, 42, 35,
}

var mask45L = [NumSquares]int{
	255, 127, 63, 31, 15, 7, 3, 1,
	127, 255, 127, 63, 31, 15, 7, 3,
	63, 127, 255, 127, 63, 31, 15, 7,
	31, 63, 127, 255, 127, 63, 31, 15,
	15, 31, 63, 127, 255, 127, 63, 31,
	7, 15, 31, 63, 127, 255, 127, 63,
	3, 7, 15, 31, 63, 127, 255, 127,
	1, 3, 7, 15, 31, 63, 127, 255,
}

var off45L = [NumSquares]int{
	28, 21, 15, 10, 6, 3, 1, 0,
	36, 28, 21, 15, 10, 6, 3, 1,
	43, 36, 28, 21, 15, 10, 6, 3,
	49, 43, 36, 28, 21, 15, 10, 6,
	54, 49, 43, 36, 28, 21, 15, 10,
	58, 54, 49, 43, 36, 28, 21, 15,
	61, 58, 54, 49, 43, 36, 28, 21,
	63, 61, 58, 54, 49, 43, 36, 28,
}

// rot45R represents the 45 degree counter-clockwise rotated square.
//
// 63 62 61 60 59 58 57 56          63 62 60 57 53 48 42 35
// 55 54 53 52 51 50 49 48          61 59 56 52 47 41 34 27
// 47 46 45 44 43 42 41 40  rot45R  58 55 51 46 40 33 26 20
// 39 38 37 36 35 34 33 32 |------> 54 50 45 39 32 25 19 14
// 31 30 29 28 27 26 25 24          49 44 38 31 24 18 13  9
// 23 22 21 20 19 18 17 16          43 37 30 23 17 12  8  5
// 15 14 13 12 11 10  9  8          36 29 22 16 11  7  4  2
//  7  6  5  4  3  2  1  0          28 21 15 10  6  3  1  0
//
// 63 62 61 60 59 58 57 56            1   3   7  15  31  63 127 255
// 55 54 53 52 51 50 49 48            3   7  15  31  63 127 255 127
// 47 46 45 44 43 42 41 40  mask45R   7  15  31  63 127 255 127  63
// 39 38 37 36 35 34 33 32 |------>  15  31  63 127 255 127  63  31
// 31 30 29 28 27 26 25 24           31  63 127 255 127  63  31  15
// 23 22 21 20 19 18 17 16           63 127 255 127  63  31  15   7
// 15 14 13 12 11 10  9  8          127 255 127  63  31  15   7   3
//  7  6  5  4  3  2  1  0          255 127  63  31  15   7   3   1
//
// 63 62 61 60 59 58 57 56          63 61 58 54 49 43 36 28
// 55 54 53 52 51 50 49 48          61 58 54 49 43 36 28 21
// 47 46 45 44 43 42 41 40  off45R  58 54 49 43 36 28 21 15
// 39 38 37 36 35 34 33 32 |------> 54 49 43 36 28 21 15 10
// 31 30 29 28 27 26 25 24          49 43 36 28 21 15 10  6
// 23 22 21 20 19 18 17 16          43 36 28 21 15 10  6  3
// 15 14 13 12 11 10  9  8          36 28 21 15 10  6  3  1
//  7  6  5  4  3  2  1  0          28 21 15 10  6  3  1  0

var rot45R = [NumSquares]Square{
	0, 1, 3, 6, 10, 15, 21, 28,
	2, 4, 7, 11, 16, 22, 29, 36,
	5, 8, 12, 17, 23, 30, 37, 43,
	9, 13, 18, 24, 31, 38, 44, 49,
	14, 19, 25, 32, 39, 45, 50, 54,
	20, 26, 33, 40, 46, 51, 55, 58,
	27, 34, 41, 47, 52, 56, 59, 61,
	35, 42, 48, 53, 57, 60, 62, 63,
}

var mask45R = [NumSquares]int{
	1, 3, 7, 15, 31, 63, 127, 255,
	3, 7, 15, 31, 63, 127, 255, 127,
	7, 15, 31, 63, 127, 255, 127, 63,
	15, 31, 63, 127, 255, 127, 63, 31,
	31, 63, 127, 255, 127, 63, 31, 15,
	63, 127, 255, 127, 63, 31, 15, 7,
	127, 255, 127, 63, 31, 15, 7, 3,
	255, 127, 63, 31, 15, 7, 3, 1,
}

var off45R = [NumSquares]int{
	0, 1, 3, 6, 10, 15, 21, 28,
	1, 3, 6, 10, 15, 21, 28, 36,
	3, 6, 10, 15, 21, 28, 36, 43,
	6, 10, 15, 21, 28, 36, 43, 49,
	10, 15, 21, 28, 36, 43, 49, 54,
	15, 21, 28, 36, 43, 49, 54, 58,
	21, 28, 36, 43, 49, 54, 58, 61,
	28, 36, 43, 49, 54, 58, 61, 63,
}

// BishopAttackboard returns all potential moves/attacks for a Bishop at the given square.
func BishopAttackboard(bb RotatedBitboard, sq Square) Bitboard {
	diagL := int(bb.rot45L>>off45L[sq]) & mask45L[sq]
	diagR := int(bb.rot45R>>off45R[sq]) & mask45R[sq]
	return bishopL[sq][diagL] | bishopR[sq][diagR]
}

var (
	bishopL, bishopR [NumSquares][numStates]Bitboard // (pos, state) -> bitboard
)

func init() {
	// Build mask by raytracing each direction, similar to Rook.
	//
	// ------
	// -X----
	// --B---
	// ------
	// ----X-
	//
	// Bishop: --B--
	// State:  -XX-X
	// Attack: -X-XX

	for sq := ZeroSquare; sq < NumSquares; sq++ {
		for state := EmptyBitboard; state <= Bitboard(mask45L[sq]); state++ {
			tmp := EmptyBitboard

			// UpLeft: X<--B (rot45L)
			for i := 1; i < min(8-sq.Rank(), 8-sq.File()); i++ {
				tmp |= BitMask(Square(sq.Rank().V()+i)<<3 + Square(sq.File().V()+i))
				if BitMask(Square(min(sq.Rank(), sq.File())+i))&state != 0 {
					break
				}
			}

			// DownRight: B-->X (rot45L)
			for i := 1; i < min(sq.Rank(), sq.File())+1; i++ {
				tmp |= BitMask(Square(sq.Rank().V()-i)<<3 + Square(sq.File().V()-i))
				if BitMask(Square(min(sq.Rank(), sq.File())-i))&state != 0 {
					break
				}
			}

			bishopL[sq][state] = tmp
		}
	}

	for sq := ZeroSquare; sq < NumSquares; sq++ {
		for state := EmptyBitboard; state <= Bitboard(mask45R[sq]); state++ {
			tmp := EmptyBitboard

			// UpRight: B-->X (rot45R)
			for i := 1; i < min(8-sq.Rank(), sq.File()+1); i++ {
				tmp |= BitMask(Square(sq.Rank().V()+i)<<3 + Square(sq.File().V()-i))
				if BitMask(Square(min(sq.Rank(), 7-sq.File())+i))&state != 0 {
					break
				}
			}

			// DownLeft: X<-R (rot45R)
			for i := 1; i < min(sq.Rank()+1, 8-sq.File()); i++ {
				tmp |= BitMask(Square(sq.Rank().V()-i)<<3 + Square(sq.File().V()+i))
				if BitMask(Square(min(sq.Rank(), 7-sq.File())-i))&state != 0 {
					break
				}
			}

			bishopR[sq][state] = tmp
		}
	}
}

// QueenAttackboard returns all potential moves/attacks for a Queen at the given square. Convenience function.
func QueenAttackboard(bb RotatedBitboard, sq Square) Bitboard {
	return RookAttackboard(bb, sq) | BishopAttackboard(bb, sq)
}

func min(r Rank, f File) int {
	if int(r) < int(f) {
		return int(r)
	}
	return int(f)
}

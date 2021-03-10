package board

import "math/rand"

// ZobristHash is a position hash based on piece-squares. It is intended for
// 3-fold repetition draw detection and hashes "identical" positions under
// that rule to the same hash value.
//
// See also: https://research.cs.wisc.edu/techreports/1970/TR88.pdf.
type ZobristHash uint64

// ZobristTable is a pseudo-randomized table for computing a position hash.
type ZobristTable struct {
	pieces    [NumColors][NumPieces][NumSquares]ZobristHash
	castling  [NumCastling]ZobristHash
	enpassant [NumSquares]ZobristHash
	turn      [NumColors]ZobristHash
}

func NewZobristTable(seed int64) *ZobristTable {
	ret := &ZobristTable{}

	r := rand.New(rand.NewSource(seed))

	for c := ZeroColor; c < NumColors; c++ {
		for p := ZeroPiece; p < NumPieces; p++ {
			for sq := ZeroSquare; sq < NumSquares; sq++ {
				ret.pieces[c][p][sq] = ZobristHash(r.Uint64())
			}
		}
		ret.turn[c] = ZobristHash(r.Uint64())
	}
	for i := ZeroCastling; i < NumCastling; i++ {
		ret.castling[i] = ZobristHash(r.Uint64())
	}
	for sq := ZeroSquare; sq < NumSquares; sq++ {
		if sq.Rank() == Rank3 || sq.Rank() == Rank6 {
			ret.enpassant[sq] = ZobristHash(r.Uint64())
		}
	}
	return ret
}

// Hash computes the zobrist hash for the given position.
func (z *ZobristTable) Hash(pos *Position, turn Color) ZobristHash {
	var hash ZobristHash

	for sq := ZeroSquare; sq < NumSquares; sq++ {
		if c, p, ok := pos.Square(sq); ok {
			hash ^= z.pieces[c][p][sq]
		}
	}
	hash ^= z.castling[pos.Castling()]
	if ep, ok := pos.EnPassant(); ok {
		hash ^= z.enpassant[ep]
	}
	hash ^= z.turn[turn]

	return hash
}

// Move computes a hash for the position after the (legal) move incrementally. Cheaper than
// computing it for the new position directly.
func (z *ZobristTable) Move(h ZobristHash, pos *Position, m Move) ZobristHash {
	hash := h

	turn, _, _ := pos.Square(m.From)

	// (1) Undo existing metastatus

	hash ^= z.castling[pos.Castling()]
	if ep, ok := pos.EnPassant(); ok {
		hash ^= z.enpassant[ep]
	}
	hash ^= z.turn[turn]

	// (2) Update hash based on moved pieces and new status

	hash ^= z.pieces[turn][m.Piece][m.From]

	switch m.Type {
	case Capture:
		hash ^= z.pieces[turn.Opponent()][m.Capture][m.To]
		hash ^= z.pieces[turn][m.Piece][m.To]

	case Promotion:
		hash ^= z.pieces[turn][m.Promotion][m.To]

	case CapturePromotion:
		hash ^= z.pieces[turn.Opponent()][m.Capture][m.To]
		hash ^= z.pieces[turn][m.Promotion][m.To]

	case EnPassant:
		hash ^= z.pieces[turn][m.Piece][m.To]
		epc, _ := m.EnPassantCapture()
		hash ^= z.pieces[turn.Opponent()][Pawn][epc]

	case KingSideCastle, QueenSideCastle:
		hash ^= z.pieces[turn][m.Piece][m.To]
		from, to, _ := m.CastlingRookMove()
		hash ^= z.pieces[turn][Rook][from]
		hash ^= z.pieces[turn][Rook][to]

	default:
		hash ^= z.pieces[turn][m.Piece][m.To]
	}

	hash ^= z.castling[pos.Castling()&m.CastlingRightsLost()]
	ept, _ := m.EnPassantTarget()
	hash ^= z.enpassant[ept]
	hash ^= z.turn[turn.Opponent()]

	return hash
}

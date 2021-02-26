package board

import (
	"fmt"
	"strings"
)

// Placement defines a piece placement.
type Placement struct {
	Square Square
	Color  Color
	Piece  Piece
}

func (p Placement) String() string {
	return fmt.Sprintf("%v@%v", printPiece(p.Color, p.Piece), p.Square)
}

// Position represents a board position suitable for move generation. It includes castling and
// en passant, but not game metadata to determine various Draw conditions.
type Position struct {
	pieces  [NumColors][NumPieces]Bitboard // Zero piece contains all pieces for color.
	rotated RotatedBitboard

	castling  Castling
	enpassant Square // zero if last move was not a Jump
}

func NewPosition(pieces []Placement, castling Castling, ep Square) (*Position, error) {
	ret := &Position{castling: castling, enpassant: ep}

	for _, p := range pieces {
		if !ret.IsEmpty(p.Square) {
			return nil, fmt.Errorf("duplicate placement: %v", p)
		}
		ret.xor(p.Square, p.Color, p.Piece)
	}

	if ret.pieces[White][King].PopCount() != 1 || ret.pieces[Black][King].PopCount() != 1 {
		return nil, fmt.Errorf("invalid number of kings")
	}
	if (KingAttackboard(ret.pieces[White][King].LastPopSquare()) & ret.pieces[Black][King]) != 0 {
		return nil, fmt.Errorf("kings cannot be adjacent")
	}
	return ret, nil
}

// Castling returns the castling rights.
func (p *Position) Castling() Castling {
	return p.castling
}

// EnPassant return the target en passant square, if previous move was a Jump. For example,
// after e2e4, the en passant target square is e3 whether or not black has pawns on d4 or f4.
func (p *Position) EnPassant() (Square, bool) {
	return p.enpassant, p.enpassant != ZeroSquare
}

// Square returns the content of the given square. Returns false is no piece present.
func (p *Position) Square(sq Square) (Color, Piece, bool) {
	if p.IsEmpty(sq) {
		return 0, 0, false
	}

	for c := ZeroColor; c < NumColors; c++ {
		if !p.pieces[c][NoPiece].IsSet(sq) {
			continue
		}
		for piece := ZeroPiece; piece < NumPieces; piece++ {
			if p.pieces[c][piece].IsSet(sq) {
				return c, piece, true
			}
		}
	}
	return 0, 0, false
}

// IsEmpty returns true iff the square is empty.
func (p *Position) IsEmpty(sq Square) bool {
	return !p.rotated.Mask().IsSet(sq)
}

// IsAttacked returns true iff the square is attacked by the opposing color. Does not include en passant.
func (p *Position) IsAttacked(c Color, sq Square) bool {
	opp := c.Opponent()

	if bishops := p.pieces[opp][Bishop] | p.pieces[opp][Queen]; bishops != 0 && BishopAttackboard(p.rotated, sq)&bishops != 0 {
		return true
	}
	if knights := p.pieces[opp][Knight]; knights != 0 && KnightAttackboard(sq)&knights != 0 {
		return true
	}
	if rooks := p.pieces[opp][Rook] | p.pieces[opp][Queen]; rooks != 0 && RookAttackboard(p.rotated, sq)&rooks != 0 {
		return true
	}
	if kings := p.pieces[opp][King]; kings != 0 && KingAttackboard(sq)&kings != 0 {
		return true
	}
	return PawnCaptureboard(opp, p.pieces[opp][Pawn])&BitMask(sq) != 0
}

// IsChecked returns true iff the color is in check. Convenient for IsAttacked(King).
func (p *Position) IsChecked(c Color) bool {
	return p.IsAttacked(c, p.pieces[c][King].LastPopSquare())
}

// TODO(herohde) 2/24/2021: consider incremental move generation via go/channels.

func (p *Position) PseudoLegalMoves() []Move {
	var ret []Move

	//	mask := ^p.pieces[p.turn][all] // cannot capture own pieces

	/*
		origin := p.pieces[p.turn][King]
		for origin != 0 {
			sq := origin.PopIndex()
			origin ^= BitMask(Square(sq))

		}
	*/
	return ret
}

func (p *Position) String() string {
	var sb strings.Builder
	for i := ZeroSquare; i < NumSquares; i++ {
		if i != 0 && i%8 == 0 {
			sb.WriteRune('/')
		}
		if color, piece, ok := p.Square(NumSquares - i - 1); ok {
			sb.WriteString(printPiece(color, piece))
		} else {
			sb.WriteRune('-')
		}
	}

	ep := "-"
	if p.enpassant != ZeroSquare {
		ep = p.enpassant.String()
	}

	return fmt.Sprintf("%v %v(%v)", sb.String(), p.castling, ep)
}

func (p *Position) xor(sq Square, color Color, piece Piece) {
	p.rotated = p.rotated.Xor(sq)
	p.pieces[color][NoPiece] ^= BitMask(sq)
	p.pieces[color][piece] ^= BitMask(sq)
}

func printPiece(c Color, p Piece) string {
	if c == White {
		return strings.ToUpper(p.String())
	}
	return strings.ToLower(p.String())
}

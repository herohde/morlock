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

	return ret, nil
}

// Move attempts to make a pseudo-legal move. The attempted move is assumed to be
// pseudo-legal and generated from the position. Returns false if not legal.
func (p *Position) Move(m Move) (*Position, bool) {
	ret := *p

	// (1) Remove piece from "from" square.

	turn, piece, ok := p.Square(m.From)
	if !ok {
		return nil, false
	}
	ret.xor(m.From, turn, piece)

	// (2) Remove any captured piece.

	if m.Type == Capture || m.Type == CapturePromotion {
		ret.xor(m.To, turn.Opponent(), m.Capture)
	}

	// (3) Add piece to "to" square.

	if m.Type == Promotion || m.Type == CapturePromotion {
		piece = m.Promotion
	}
	ret.xor(m.To, turn, piece)

	// (4) Handle special moves/captures.

	switch m.Type {
	case EnPassant:
		var capture Square
		if turn == White {
			capture = NewSquare(m.To.File(), Rank5)
		} else {
			capture = NewSquare(m.To.File(), Rank4)
		}
		ret.xor(capture, turn.Opponent(), Pawn)

	case KingSideCastle, QueenSideCastle:
		for _, sq := range safeCastlingSquares(turn, m.Type) {
			if p.IsAttacked(turn, sq) {
				return nil, false
			}
		}
		for _, sq := range rookSquares(turn, m.Type) {
			ret.xor(sq, turn, Rook)
		}
	}

	// (5) Update EnPassant status.

	if m.Type == Jump {
		if turn == White {
			ret.enpassant = NewSquare(m.From.File(), Rank3)
		} else {
			ret.enpassant = NewSquare(m.From.File(), Rank6)
		}
	} else {
		ret.enpassant = ZeroSquare
	}

	// (6) Update Castling status. If king moves, rights are lost. Ditto if rook moves or is captured.

	if m.From == E1 || m.From == A1 || m.To == A1 {
		ret.castling &^= WhiteQueenSideCastle
	}
	if m.From == E1 || m.From == H1 || m.To == H1 {
		ret.castling &^= WhiteKingSideCastle
	}
	if m.From == E8 || m.From == A8 || m.To == A8 {
		ret.castling &^= BlackQueenSideCastle
	}
	if m.From == E8 || m.From == H8 || m.To == H8 {
		ret.castling &^= BlackKingSideCastle
	}

	// (7) Validate that move does not leave own king in check.

	if ret.IsChecked(turn) {
		return nil, false
	}
	return &ret, true
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
	if pos := p.pieces[c][King].LastPopSquare(); pos != NumSquares {
		return p.IsAttacked(c, pos)
	}
	return false
}

var (
	whiteKingSideCastlingMask  = BitMask(G1) | BitMask(F1)
	whiteQueenSideCastlingMask = BitMask(B1) | BitMask(C1) | BitMask(D1)
	blackKingSideCastlingMask  = BitMask(G8) | BitMask(F8)
	blackQueenSideCastlingMask = BitMask(B8) | BitMask(C8) | BitMask(D8)
)

// PseudoLegalMoves returns a list of all pseudo-legal moves. The move may not respect
// either side being in check, which must be validated subsequently.
func (p *Position) PseudoLegalMoves(turn Color) []Move {
	mask := ^p.pieces[turn][NoPiece] // cannot capture own pieces

	captures := p.pieces[turn.Opponent()][NoPiece]
	moves := ^captures
	jumps := PawnJumpRank(turn)
	promos := PawnPromotionRank(turn)

	var ret []Move

	queens := p.pieces[turn][Queen]
	for queens != EmptyBitboard {
		from := queens.LastPopSquare()
		queens ^= BitMask(from)

		attackboard := (RookAttackboard(p.rotated, from) | BishopAttackboard(p.rotated, from)) & mask
		p.emitMove(Normal, Queen, from, attackboard&moves, &ret)
		p.emitMove(Capture, Queen, from, attackboard&captures, &ret)
	}

	rooks := p.pieces[turn][Rook]
	for rooks != EmptyBitboard {
		from := rooks.LastPopSquare()
		rooks ^= BitMask(from)

		attackboard := RookAttackboard(p.rotated, from) & mask
		p.emitMove(Normal, Rook, from, attackboard&moves, &ret)
		p.emitMove(Capture, Rook, from, attackboard&captures, &ret)
	}

	bishops := p.pieces[turn][Bishop]
	for bishops != EmptyBitboard {
		from := bishops.LastPopSquare()
		bishops ^= BitMask(from)

		attackboard := BishopAttackboard(p.rotated, from) & mask
		p.emitMove(Normal, Bishop, from, attackboard&moves, &ret)
		p.emitMove(Capture, Bishop, from, attackboard&captures, &ret)
	}

	knights := p.pieces[turn][Knight]
	for knights != EmptyBitboard {
		from := knights.LastPopSquare()
		knights ^= BitMask(from)

		attackboard := KnightAttackboard(from) & mask
		p.emitMove(Normal, Knight, from, attackboard&moves, &ret)
		p.emitMove(Capture, Knight, from, attackboard&captures, &ret)
	}

	pawns := p.pieces[turn][Pawn]
	for pawns != EmptyBitboard {
		from := pawns.LastPopSquare()
		origin := BitMask(from)
		pawns ^= origin

		captureboard := PawnCaptureboard(turn, origin) & mask
		pushboard := PawnMoveboard(p.rotated.rot, turn, origin)
		jumpboard := PawnMoveboard(p.rotated.rot, turn, pushboard) & jumps

		p.emitMove(Capture, Pawn, from, captureboard&captures&^promos, &ret)
		p.emitMove(Push, Pawn, from, pushboard&^promos, &ret)
		p.emitMove(Jump, Pawn, from, jumpboard, &ret)

		p.emitPromo(CapturePromotion, Pawn, from, captureboard&captures&promos, &ret)
		p.emitPromo(Promotion, Pawn, from, pushboard&promos, &ret)

		if p.enpassant != ZeroSquare {
			p.emitMove(EnPassant, Pawn, from, captureboard&BitMask(p.enpassant), &ret)
		}
	}

	if king := p.pieces[turn][King]; king != EmptyBitboard {
		from := king.LastPopSquare()

		attackboard := KingAttackboard(from) & mask
		p.emitMove(Normal, King, from, attackboard&moves, &ret)
		p.emitMove(Capture, King, from, attackboard&captures, &ret)

		if turn == White {
			if p.castling.IsAllowed(WhiteKingSideCastle) && (whiteKingSideCastlingMask&p.rotated.rot) == 0 && p.pieces[turn][Rook]&BitMask(H1) != 0 {
				p.emitMove(KingSideCastle, King, from, BitMask(G1), &ret)
			}
			if p.castling.IsAllowed(WhiteQueenSideCastle) && (whiteQueenSideCastlingMask&p.rotated.rot) == 0 && p.pieces[turn][Rook]&BitMask(A1) != 0 {
				p.emitMove(QueenSideCastle, King, from, BitMask(C1), &ret)
			}
		} else {
			if p.castling.IsAllowed(BlackKingSideCastle) && (blackKingSideCastlingMask&p.rotated.rot) == 0 && p.pieces[turn][Rook]&BitMask(H8) != 0 {
				p.emitMove(KingSideCastle, King, from, BitMask(G8), &ret)
			}
			if p.castling.IsAllowed(BlackQueenSideCastle) && (blackQueenSideCastlingMask&p.rotated.rot) == 0 && p.pieces[turn][Rook]&BitMask(A8) != 0 {
				p.emitMove(QueenSideCastle, King, from, BitMask(C8), &ret)
			}
		}
	}

	return ret
}

func (p *Position) emitMove(t MoveType, piece Piece, from Square, attackboard Bitboard, out *[]Move) {
	for attackboard != EmptyBitboard {
		to := attackboard.LastPopSquare()
		attackboard ^= BitMask(to)

		capture := NoPiece
		if t == Capture {
			_, capture, _ = p.Square(to)
		}
		*out = append(*out, Move{Type: t, Piece: piece, From: from, To: to, Capture: capture})
	}
}

func (p *Position) emitPromo(t MoveType, piece Piece, from Square, attackboard Bitboard, out *[]Move) {
	for attackboard != EmptyBitboard {
		to := attackboard.LastPopSquare()
		attackboard ^= BitMask(to)

		capture := NoPiece
		if t == CapturePromotion {
			_, capture, _ = p.Square(to)
		}

		// Emit under-promotions as well.

		for _, pc := range []Piece{Queen, Rook, Knight, Bishop} {
			*out = append(*out, Move{Type: t, Piece: piece, From: from, To: to, Capture: capture, Promotion: pc})
		}
	}
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

// safeCastlingSquares returns the squares that must not be in check to castle.
// Does not include the king to square.
func safeCastlingSquares(c Color, t MoveType) []Square {
	if c == White {
		switch t {
		case KingSideCastle:
			return []Square{E1, F1}
		case QueenSideCastle:
			return []Square{E1, D1}
		default:
			return nil
		}
	} else {
		switch t {
		case KingSideCastle:
			return []Square{E8, F8}
		case QueenSideCastle:
			return []Square{E8, D8}
		default:
			return nil
		}
	}
}

// rookSquares return the rook squares for castling.
func rookSquares(c Color, t MoveType) []Square {
	if c == White {
		switch t {
		case KingSideCastle:
			return []Square{F1, H1}
		case QueenSideCastle:
			return []Square{A1, D1}
		default:
			return nil
		}
	} else {
		switch t {
		case KingSideCastle:
			return []Square{F8, H8}
		case QueenSideCastle:
			return []Square{A8, D8}
		default:
			return nil
		}
	}
}

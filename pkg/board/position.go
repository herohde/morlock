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

	if m.IsCapture() {
		ret.xor(m.To, turn.Opponent(), m.Capture)
	}

	// (3) Add piece to "to" square.

	if m.IsPromotion() {
		piece = m.Promotion
	}
	ret.xor(m.To, turn, piece)

	// (4) Handle special moves/captures.

	switch m.Type {
	case EnPassant:
		capture, _ := m.EnPassantCapture()
		ret.xor(capture, turn.Opponent(), Pawn)

	case KingSideCastle, QueenSideCastle:
		for _, sq := range safeCastlingSquares(turn, m.Type) {
			if p.IsAttacked(turn, sq) {
				return nil, false
			}
		}

		from, to, _ := m.CastlingRookMove()
		ret.xor(from, turn, Rook)
		ret.xor(to, turn, Rook)
	}

	// (5) Update EnPassant and castling status.

	ret.enpassant, _ = m.EnPassantTarget()
	ret.castling &^= m.CastlingRightsLost()

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

// Rotated returns the rotated bitboard.
func (p *Position) Rotated() RotatedBitboard {
	return p.rotated
}

// All returns a bitboard contains all pirces.
func (p *Position) All() Bitboard {
	return p.rotated.rot
}

// Color returns the bitboard for a given color.
func (p *Position) Color(c Color) Bitboard {
	return p.pieces[c][NoPiece]
}

// Piece returns the bitboard for a given color/piece.
func (p *Position) Piece(c Color, piece Piece) Bitboard {
	return p.pieces[c][piece]
}

// PieceSquares returns the squares for a given color/piece.
func (p *Position) PieceSquares(c Color, piece Piece) []Square {
	return p.pieces[c][piece].ToSquares()
}

// KingSquare returns the square for a given color. Must be valid and unique.
func (p *Position) KingSquare(c Color) Square {
	return p.pieces[c][King].LastPopSquare()
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

// IsDefended returns true iff the square is defended by the color.
func (p *Position) IsDefended(c Color, sq Square) bool {
	return p.IsAttacked(c.Opponent(), sq)
}

// IsAttacked returns true iff the square is attacked by the opposing color. Does not include en passant.
func (p *Position) IsAttacked(c Color, sq Square) bool {
	opp := c.Opponent()

	for _, piece := range KingQueenRookKnightBishop {
		if pieces := p.pieces[opp][piece]; pieces != 0 && Attackboard(p.rotated, sq, piece)&pieces != 0 {
			return true
		}
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

// IsCheckMate returns true iff the color is checkmate. Convenient for IsChecked && len(LegalMoves)==0.
func (p *Position) IsCheckMate(c Color) bool {
	return p.IsChecked(c) && len(p.LegalMoves(c)) == 0
}

var (
	whiteSquareMask = Bitboard(0xaaaaaaaaaaaaaaaa)
)

// HasInsufficientMaterial returns true iff there is not sufficient material for either side to win.
// The cases are: K v K, KN v K, KB v KB (or KBB v K) w/ Bishops on same square color. Assumes 2 kings.
func (p *Position) HasInsufficientMaterial() bool {
	switch p.rotated.rot.PopCount() {
	case 2:
		return true
	case 3:
		weak := p.pieces[White][Knight] | p.pieces[Black][Knight] | p.pieces[White][Bishop] | p.pieces[Black][Bishop]
		return weak.PopCount() == 1

	case 4:
		bishops := p.pieces[White][Bishop] | p.pieces[Black][Bishop]
		return bishops.PopCount() == 2 && (whiteSquareMask&bishops).PopCount() != 1

	default:
		return false
	}
}

var (
	whiteKingSideCastlingMask  = BitMask(G1) | BitMask(F1)
	whiteQueenSideCastlingMask = BitMask(B1) | BitMask(C1) | BitMask(D1)
	blackKingSideCastlingMask  = BitMask(G8) | BitMask(F8)
	blackQueenSideCastlingMask = BitMask(B8) | BitMask(C8) | BitMask(D8)
)

// LegalMoves returns a list of all legal moves. Convenience function.
func (p *Position) LegalMoves(turn Color) []Move {
	var ret []Move
	for _, m := range p.PseudoLegalMoves(turn) {
		if _, ok := p.Move(m); ok {
			ret = append(ret, m)
		}
	}
	return ret
}

// PseudoLegalMoves returns a list of all pseudo-legal moves. The move may not respect
// either side being in check, which must be validated subsequently.
func (p *Position) PseudoLegalMoves(turn Color) []Move {
	mask := ^p.pieces[turn][NoPiece] // cannot capture own pieces

	captures := p.pieces[turn.Opponent()][NoPiece]
	moves := ^captures
	jumps := PawnJumpRank(turn)
	promos := PawnPromotionRank(turn)

	ret := make([]Move, 0, 50)

	for _, piece := range QueenRookKnightBishop {
		pieces := p.pieces[turn][piece]
		for pieces != EmptyBitboard {
			from := pieces.LastPopSquare()
			pieces ^= BitMask(from)

			attackboard := Attackboard(p.rotated, from, piece) & mask
			p.emitMove(turn, Normal, piece, from, attackboard&moves, &ret)
			p.emitMove(turn, Capture, piece, from, attackboard&captures, &ret)
		}
	}

	pawns := p.pieces[turn][Pawn]
	for pawns != EmptyBitboard {
		from := pawns.LastPopSquare()
		origin := BitMask(from)
		pawns ^= origin

		captureboard := PawnCaptureboard(turn, origin) & mask
		pushboard := PawnMoveboard(p.rotated.rot, turn, origin)
		jumpboard := PawnMoveboard(p.rotated.rot, turn, pushboard) & jumps

		p.emitMove(turn, Capture, Pawn, from, captureboard&captures&^promos, &ret)
		p.emitMove(turn, Push, Pawn, from, pushboard&^promos, &ret)
		p.emitMove(turn, Jump, Pawn, from, jumpboard, &ret)

		p.emitPromo(turn, CapturePromotion, Pawn, from, captureboard&captures&promos, &ret)
		p.emitPromo(turn, Promotion, Pawn, from, pushboard&promos, &ret)

		if p.enpassant != ZeroSquare {
			p.emitMove(turn, EnPassant, Pawn, from, captureboard&BitMask(p.enpassant), &ret)
		}
	}

	if king := p.pieces[turn][King]; king != EmptyBitboard {
		from := king.LastPopSquare()

		attackboard := KingAttackboard(from) & mask
		p.emitMove(turn, Normal, King, from, attackboard&moves, &ret)
		p.emitMove(turn, Capture, King, from, attackboard&captures, &ret)

		if turn == White {
			if p.castling.IsAllowed(WhiteKingSideCastle) && (whiteKingSideCastlingMask&p.rotated.rot) == 0 && p.pieces[turn][Rook]&BitMask(H1) != 0 {
				p.emitMove(turn, KingSideCastle, King, from, BitMask(G1), &ret)
			}
			if p.castling.IsAllowed(WhiteQueenSideCastle) && (whiteQueenSideCastlingMask&p.rotated.rot) == 0 && p.pieces[turn][Rook]&BitMask(A1) != 0 {
				p.emitMove(turn, QueenSideCastle, King, from, BitMask(C1), &ret)
			}
		} else {
			if p.castling.IsAllowed(BlackKingSideCastle) && (blackKingSideCastlingMask&p.rotated.rot) == 0 && p.pieces[turn][Rook]&BitMask(H8) != 0 {
				p.emitMove(turn, KingSideCastle, King, from, BitMask(G8), &ret)
			}
			if p.castling.IsAllowed(BlackQueenSideCastle) && (blackQueenSideCastlingMask&p.rotated.rot) == 0 && p.pieces[turn][Rook]&BitMask(A8) != 0 {
				p.emitMove(turn, QueenSideCastle, King, from, BitMask(C8), &ret)
			}
		}
	}

	return ret
}

func (p *Position) emitMove(turn Color, t MoveType, piece Piece, from Square, attackboard Bitboard, out *[]Move) {
	for attackboard != EmptyBitboard {
		to := attackboard.LastPopSquare()
		attackboard ^= BitMask(to)

		capture := NoPiece
		if t == Capture {
			capture = p.captureAt(to, turn)
		}
		*out = append(*out, Move{Type: t, Piece: piece, From: from, To: to, Capture: capture})
	}
}

func (p *Position) emitPromo(turn Color, t MoveType, piece Piece, from Square, attackboard Bitboard, out *[]Move) {
	for attackboard != EmptyBitboard {
		to := attackboard.LastPopSquare()
		attackboard ^= BitMask(to)

		capture := NoPiece
		if t == CapturePromotion {
			capture = p.captureAt(to, turn)
		}

		// Emit under-promotions as well.

		for _, pc := range QueenRookKnightBishop {
			*out = append(*out, Move{Type: t, Piece: piece, From: from, To: to, Capture: capture, Promotion: pc})
		}
	}
}

func (p *Position) captureAt(sq Square, turn Color) Piece {
	for piece := ZeroPiece; piece < NumPieces; piece++ {
		if p.pieces[turn.Opponent()][piece].IsSet(sq) {
			return piece
		}
	}
	return NoPiece
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

package board_test

import (
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestPseudoLegalMoves(t *testing.T) {

	t.Run("pawns", func(t *testing.T) {
		tests := []struct {
			turn      board.Color
			pieces    []board.Placement
			enpassant board.Square
			expected  []board.Move
		}{
			{ // Empty board
				board.White,
				nil,
				board.ZeroSquare,
				nil,
			},
			{ // Pawn @ E2,G6
				board.White,
				[]board.Placement{
					{board.E2, board.White, board.Pawn},
					{board.G5, board.White, board.Pawn},
				},
				board.ZeroSquare,
				[]board.Move{
					{Type: board.Push, Piece: board.Pawn, From: board.E2, To: board.E3},
					{Type: board.Jump, Piece: board.Pawn, From: board.E2, To: board.E4},
					{Type: board.Push, Piece: board.Pawn, From: board.G5, To: board.G6},
				},
			},
			{ // Pawn @ C7,G6
				board.Black,
				[]board.Placement{
					{board.C7, board.Black, board.Pawn},
					{board.G6, board.Black, board.Pawn},
				},
				board.ZeroSquare,
				[]board.Move{
					{Type: board.Push, Piece: board.Pawn, From: board.G6, To: board.G5},
					{Type: board.Push, Piece: board.Pawn, From: board.C7, To: board.C6},
					{Type: board.Jump, Piece: board.Pawn, From: board.C7, To: board.C5},
				},
			},
			{ // Pawn @ E2,H6 -- obstructed w/ capture
				board.White,
				[]board.Placement{
					{board.E2, board.White, board.Pawn},
					{board.E4, board.Black, board.Bishop},
					{board.D3, board.Black, board.Knight},
					{board.D4, board.Black, board.Rook},
					{board.H5, board.White, board.Pawn},
					{board.G6, board.Black, board.Bishop},
					{board.H6, board.Black, board.Knight},
					{board.A6, board.Black, board.Rook},
				},
				board.ZeroSquare,
				[]board.Move{
					{Type: board.Capture, Piece: board.Pawn, From: board.E2, To: board.D3, Capture: board.Knight},
					{Type: board.Push, Piece: board.Pawn, From: board.E2, To: board.E3},
					{Type: board.Capture, Piece: board.Pawn, From: board.H5, To: board.G6, Capture: board.Bishop},
				},
			},
			{ // Pawn @ D7
				board.White,
				[]board.Placement{
					{board.D7, board.White, board.Pawn},
				},
				board.ZeroSquare,
				[]board.Move{
					{Type: board.Promotion, Piece: board.Pawn, From: board.D7, To: board.D8, Promotion: board.Queen},
					{Type: board.Promotion, Piece: board.Pawn, From: board.D7, To: board.D8, Promotion: board.Rook},
					{Type: board.Promotion, Piece: board.Pawn, From: board.D7, To: board.D8, Promotion: board.Knight},
					{Type: board.Promotion, Piece: board.Pawn, From: board.D7, To: board.D8, Promotion: board.Bishop},
				},
			},
			{ // Pawn @ C4,E4,F4 -- en passant
				board.Black,
				[]board.Placement{
					{board.C4, board.Black, board.Pawn},
					{board.D4, board.White, board.Pawn},
					{board.E4, board.Black, board.Pawn},
					{board.F4, board.Black, board.Pawn},
				},
				board.D3,
				[]board.Move{
					{Type: board.Push, Piece: board.Pawn, From: board.F4, To: board.F3},
					{Type: board.Push, Piece: board.Pawn, From: board.E4, To: board.E3},
					{Type: board.EnPassant, Piece: board.Pawn, From: board.E4, To: board.D3},
					{Type: board.Push, Piece: board.Pawn, From: board.C4, To: board.C3},
					{Type: board.EnPassant, Piece: board.Pawn, From: board.C4, To: board.D3},
				},
			},
		}

		for _, tt := range tests {
			pos, err := board.NewPosition(tt.pieces, 0, tt.enpassant)
			require.NoError(t, err)

			actual := pos.PseudoLegalMoves(tt.turn)
			assert.Equal(t, printMoves(tt.expected), printMoves(actual))
		}
	})

	t.Run("officers", func(t *testing.T) {
		tests := []struct {
			pieces   []board.Placement
			expected []board.Move
		}{
			{ // King @ A3
				[]board.Placement{
					{board.A3, board.White, board.King},
					{board.B3, board.Black, board.Rook},
					{board.A2, board.Black, board.Bishop},
				},
				[]board.Move{
					{Type: board.Normal, Piece: board.King, From: board.A3, To: board.B2},
					{Type: board.Normal, Piece: board.King, From: board.A3, To: board.B4},
					{Type: board.Normal, Piece: board.King, From: board.A3, To: board.A4},
					{Type: board.Capture, Piece: board.King, From: board.A3, To: board.A2, Capture: board.Bishop},
					{Type: board.Capture, Piece: board.King, From: board.A3, To: board.B3, Capture: board.Rook},
				},
			},
			{ // Knight @ A3
				[]board.Placement{
					{board.A3, board.White, board.Knight},
					{board.B1, board.Black, board.Rook},
					{board.B2, board.Black, board.Bishop},
					{board.C2, board.Black, board.Queen},
				},
				[]board.Move{
					{Type: board.Normal, Piece: board.Knight, From: board.A3, To: board.C4},
					{Type: board.Normal, Piece: board.Knight, From: board.A3, To: board.B5},
					{Type: board.Capture, Piece: board.Knight, From: board.A3, To: board.B1, Capture: board.Rook},
					{Type: board.Capture, Piece: board.Knight, From: board.A3, To: board.C2, Capture: board.Queen},
				},
			},
			{ // Bishop @ G3 -- partly obstructed
				[]board.Placement{
					{board.G3, board.White, board.Bishop},
					{board.F2, board.Black, board.Rook},
					{board.E5, board.Black, board.Rook},
				},
				[]board.Move{
					{Type: board.Normal, Piece: board.Bishop, From: board.G3, To: board.H2},
					{Type: board.Normal, Piece: board.Bishop, From: board.G3, To: board.H4},
					{Type: board.Normal, Piece: board.Bishop, From: board.G3, To: board.F4},
					{Type: board.Capture, Piece: board.Bishop, From: board.G3, To: board.F2, Capture: board.Rook},
					{Type: board.Capture, Piece: board.Bishop, From: board.G3, To: board.E5, Capture: board.Rook},
				},
			},
			{ // Bishop @ D3
				[]board.Placement{
					{board.D3, board.White, board.Bishop},
					{board.C2, board.Black, board.Rook},
					{board.C4, board.Black, board.Rook},
					{board.F5, board.Black, board.Rook},
				},
				[]board.Move{
					{Type: board.Normal, Piece: board.Bishop, From: board.D3, To: board.F1},
					{Type: board.Normal, Piece: board.Bishop, From: board.D3, To: board.E2},
					{Type: board.Normal, Piece: board.Bishop, From: board.D3, To: board.E4},
					{Type: board.Capture, Piece: board.Bishop, From: board.D3, To: board.C2, Capture: board.Rook},
					{Type: board.Capture, Piece: board.Bishop, From: board.D3, To: board.C4, Capture: board.Rook},
					{Type: board.Capture, Piece: board.Bishop, From: board.D3, To: board.F5, Capture: board.Rook},
				},
			},
			{ // Rook @ D3
				[]board.Placement{
					{board.D3, board.White, board.Rook},
					{board.B3, board.Black, board.Rook},
					{board.E3, board.Black, board.Bishop},
					{board.D5, board.Black, board.Queen},
				},
				[]board.Move{
					{Type: board.Normal, Piece: board.Rook, From: board.D3, To: board.D1},
					{Type: board.Normal, Piece: board.Rook, From: board.D3, To: board.D2},
					{Type: board.Normal, Piece: board.Rook, From: board.D3, To: board.C3},
					{Type: board.Normal, Piece: board.Rook, From: board.D3, To: board.D4},
					{Type: board.Capture, Piece: board.Rook, From: board.D3, To: board.E3, Capture: board.Bishop},
					{Type: board.Capture, Piece: board.Rook, From: board.D3, To: board.B3, Capture: board.Rook},
					{Type: board.Capture, Piece: board.Rook, From: board.D3, To: board.D5, Capture: board.Queen},
				},
			},
			{ // Queen @ D3 -- union of bishop/rook above
				[]board.Placement{
					{board.D3, board.White, board.Queen},
					{board.C2, board.Black, board.Rook},
					{board.C4, board.Black, board.Rook},
					{board.F5, board.Black, board.Rook},
					{board.B3, board.Black, board.Rook},
					{board.E3, board.Black, board.Bishop},
					{board.D5, board.Black, board.Queen},
				},
				[]board.Move{
					{Type: board.Normal, Piece: board.Queen, From: board.D3, To: board.F1},
					{Type: board.Normal, Piece: board.Queen, From: board.D3, To: board.D1},
					{Type: board.Normal, Piece: board.Queen, From: board.D3, To: board.E2},
					{Type: board.Normal, Piece: board.Queen, From: board.D3, To: board.D2},
					{Type: board.Normal, Piece: board.Queen, From: board.D3, To: board.C3},
					{Type: board.Normal, Piece: board.Queen, From: board.D3, To: board.E4},
					{Type: board.Normal, Piece: board.Queen, From: board.D3, To: board.D4},
					{Type: board.Capture, Piece: board.Queen, From: board.D3, To: board.C2, Capture: board.Rook},
					{Type: board.Capture, Piece: board.Queen, From: board.D3, To: board.E3, Capture: board.Bishop},
					{Type: board.Capture, Piece: board.Queen, From: board.D3, To: board.B3, Capture: board.Rook},
					{Type: board.Capture, Piece: board.Queen, From: board.D3, To: board.C4, Capture: board.Rook},
					{Type: board.Capture, Piece: board.Queen, From: board.D3, To: board.F5, Capture: board.Rook},
					{Type: board.Capture, Piece: board.Queen, From: board.D3, To: board.D5, Capture: board.Queen},
				},
			},
		}

		for _, tt := range tests {
			pos, err := board.NewPosition(tt.pieces, 0, 0)
			require.NoError(t, err)

			actual := pos.PseudoLegalMoves(board.White)
			assert.Equal(t, printMoves(tt.expected), printMoves(actual))
		}
	})

	t.Run("castling", func(t *testing.T) {
		tests := []struct {
			turn     board.Color
			pieces   []board.Placement
			castling board.Castling
			expected []board.Move
		}{
			{ // No rights
				board.White,
				[]board.Placement{
					{board.E1, board.White, board.King},
					{board.H1, board.White, board.Rook},
					{board.A1, board.White, board.Rook},
				},
				0,
				nil,
			},
			{ // Full rights.
				board.White,
				[]board.Placement{
					{board.E1, board.White, board.King},
					{board.H1, board.White, board.Rook},
					{board.A1, board.White, board.Rook},
				},
				board.FullCastingRights,
				[]board.Move{
					{Type: board.KingSideCastle, Piece: board.King, From: board.E1, To: board.G1},
					{Type: board.QueenSideCastle, Piece: board.King, From: board.E1, To: board.C1},
				},
			},
			{ // Obstructed
				board.Black,
				[]board.Placement{
					{board.E8, board.Black, board.King},
					{board.H8, board.Black, board.Rook},
					{board.G8, board.White, board.Bishop},
					{board.A8, board.Black, board.Rook},
				},
				board.FullCastingRights,
				[]board.Move{
					{Type: board.QueenSideCastle, Piece: board.King, From: board.E8, To: board.C8},
				},
			},
			{ // Partial rights.
				board.Black,
				[]board.Placement{
					{board.E8, board.Black, board.King},
					{board.H8, board.Black, board.Rook},
					{board.A8, board.Black, board.Rook},
				},
				board.BlackQueenSideCastle | board.WhiteKingSideCastle,
				[]board.Move{
					{Type: board.QueenSideCastle, Piece: board.King, From: board.E8, To: board.C8},
				},
			},
		}

		for _, tt := range tests {
			pos, err := board.NewPosition(tt.pieces, tt.castling, 0)
			require.NoError(t, err)

			actual := filterMoves(pos.PseudoLegalMoves(tt.turn), func(move board.Move) bool {
				return move.Type == board.KingSideCastle || move.Type == board.QueenSideCastle
			})
			assert.Equal(t, printMoves(tt.expected), printMoves(actual))
		}
	})
}

func TestPerft1(t *testing.T) {
	tests := []struct {
		fen      string
		expected int
	}{
		// FEN: http://www.talkchess.com/forum3/viewtopic.php?t=48616. Missed Bc5xb4 due to BB mask off by one.
		{"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/1PB1P1b1/P1NP1N2/2P1QPPP/R4RK1 b - b3 0 10", 45},
	}

	for _, tt := range tests {
		pos, turn, _, _, err := fen.Decode(tt.fen)
		assert.NoError(t, err)

		moves := pos.PseudoLegalMoves(turn)
		assert.Equal(t, len(moves), tt.expected)
	}
}

func filterMoves(ms []board.Move, fn func(move board.Move) bool) []board.Move {
	var list []board.Move
	for _, m := range ms {
		if fn(m) {
			list = append(list, m)
		}
	}
	return list
}

func printMoves(ms []board.Move) string {
	var list []string
	for _, m := range ms {
		list = append(list, m.String())
	}
	return strings.Join(list, "\n")
}

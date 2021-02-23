package board_test

import (
	"testing"

	"github.com/herohde/morlock/pkg/board"
	"github.com/stretchr/testify/assert"
)

func TestBitboard(t *testing.T) {

	t.Run("popcount", func(t *testing.T) {
		tests := []struct {
			bb       board.Bitboard
			expected int
		}{
			{board.EmptyBitboard, 0},
			{board.BitMask(board.G4), 1},
			{board.BitMask(board.G3) | board.BitMask(board.G4), 2},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.bb.PopCount(), tt.expected)
		}
	})

	t.Run("string", func(t *testing.T) {
		tests := []struct {
			bb       board.Bitboard
			expected string
		}{
			{board.EmptyBitboard, "--------/--------/--------/--------/--------/--------/--------/--------"},
			{board.BitMask(board.H1), "--------/--------/--------/--------/--------/--------/--------/-------X"},
			{board.BitMask(board.G3) | board.BitMask(board.G4), "--------/--------/--------/--------/------X-/------X-/--------/--------"},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.bb.String(), tt.expected)
		}
	})

	t.Run("king", func(t *testing.T) {
		tests := []struct {
			sq       board.Square
			expected string
		}{
			{board.H1, "--------/--------/--------/--------/--------/--------/------XX/------X-"},
			{board.D1, "--------/--------/--------/--------/--------/--------/--XXX---/--X-X---"},
			{board.D3, "--------/--------/--------/--------/--XXX---/--X-X---/--XXX---/--------"},
			{board.A3, "--------/--------/--------/--------/XX------/-X------/XX------/--------"},
			{board.B7, "XXX-----/X-X-----/XXX-----/--------/--------/--------/--------/--------"},
			{board.A8, "-X------/XX------/--------/--------/--------/--------/--------/--------"},
			{board.H8, "------X-/------XX/--------/--------/--------/--------/--------/--------"},
		}

		for _, tt := range tests {
			assert.Equal(t, board.KingAttackboard(tt.sq).String(), tt.expected)
		}
	})

	t.Run("knight", func(t *testing.T) {
		tests := []struct {
			sq       board.Square
			expected string
		}{
			{board.H1, "--------/--------/--------/--------/--------/------X-/-----X--/--------"},
			{board.D1, "--------/--------/--------/--------/--------/--X-X---/-X---X--/--------"},
			{board.D3, "--------/--------/--------/--X-X---/-X---X--/--------/-X---X--/--X-X---"},
			{board.A3, "--------/--------/--------/-X------/--X-----/--------/--X-----/-X------"},
			{board.B7, "---X----/--------/---X----/X-X-----/--------/--------/--------/--------"},
			{board.A8, "--------/--X-----/-X------/--------/--------/--------/--------/--------"},
			{board.H8, "--------/-----X--/------X-/--------/--------/--------/--------/--------"},
		}

		for _, tt := range tests {
			assert.Equal(t, board.KnightAttackboard(tt.sq).String(), tt.expected)
		}
	})

	t.Run("rook", func(t *testing.T) {
		tests := []struct {
			bb       board.Bitboard
			sq       board.Square
			expected string
		}{
			{board.EmptyBitboard, board.H1, "-------X/-------X/-------X/-------X/-------X/-------X/-------X/XXXXXXX-"},
			{board.EmptyBitboard, board.D3, "---X----/---X----/---X----/---X----/---X----/XXX-XXXX/---X----/---X----"},
			{board.EmptyBitboard, board.A6, "X-------/X-------/-XXXXXXX/X-------/X-------/X-------/X-------/X-------"},

			{board.BitMask(board.H2), board.H1, "--------/--------/--------/--------/--------/--------/-------X/XXXXXXX-"},
			{board.BitRank(board.Rank2), board.H1, "--------/--------/--------/--------/--------/--------/-------X/XXXXXXX-"},
			{board.BitMask(board.H2) | board.BitMask(board.D1), board.H1, "--------/--------/--------/--------/--------/--------/-------X/---XXXX-"},
			{board.BitMask(board.B4) | board.BitMask(board.G4), board.E4, "----X---/----X---/----X---/----X---/-XXX-XX-/----X---/----X---/----X---"},
			{board.BitMask(board.E2) | board.BitMask(board.E7), board.E4, "--------/----X---/----X---/----X---/XXXX-XXX/----X---/----X---/--------"},
		}

		for _, tt := range tests {
			assert.Equal(t, board.RookAttackboard(board.NewRotatedBitboard(tt.bb), tt.sq).String(), tt.expected)
		}
	})

}

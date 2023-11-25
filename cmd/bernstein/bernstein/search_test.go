package bernstein_test

import (
	"github.com/herohde/morlock/cmd/bernstein/bernstein"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFindPlausibleMoves(t *testing.T) {
	tests := []struct {
		pos      string
		limit    int
		expected string
	}{
		{fen.Initial, 7, "Nb1-a3 Nb1-c3 Ng1-f3 Ng1-h3 e2-e4 e2-e3 d2-d4"},
		{"r1bqk2r/pppp1ppp/2nbpn2/6B1/3P4/2PB1N2/PP3PPP/RN1Q1RK1 b kq - 5 7", 6, "h7-h5 Nf6-g4 Nf6-h5 Nf6-d5 Nf6-g8 0-0"}, // bad: 2023 game1 move7
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.pos)
		require.NoError(t, err)

		actual := bernstein.FindPlausibleMoves(b)
		assert.Equal(t, tt.expected, board.PrintMoves(actual[:tt.limit]), "board: %v", b)
	}
}

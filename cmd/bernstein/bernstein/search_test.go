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
		// 2023 game1
		{"r1bqk2r/pppp1ppp/2nbpn2/6B1/3P4/2PB1N2/PP3PPP/RN1Q1RK1 b kq - 5 7", 1, "0-0"}, // move 7: "h7-h5 Nf6-g4 Nf6-h5 Nf6-d5 Nf6-g8 0-0" due to bad loss
		// 2023 game2
		{"rnbqk1nr/ppp2ppp/3p4/2b1p3/1PP1P3/P4N2/3P1PPP/RNBQKB1R b KQkq - 0 5", 7, "Bc5-d4 Bc5-b6 Bc8-g4 Ng8-h6 Ng8-f6 Bc8-e6 Nb8-c6"}, // move 5: search did not pick B
		{"rn1q1rk1/ppp2ppp/3p1n2/8/1PP1p1b1/PQ3N2/4BPPP/RNB2RK1 b - - 1 12", 7, "e4*f3 Bg4*f3 Nb8-c6 Nb8-a6 Nb8-d7 c7-c5 d6-d5"},       // move 12: no capture
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.pos)
		require.NoError(t, err)

		actual := bernstein.FindPlausibleMoves(b)
		assert.Equal(t, tt.expected, board.PrintMoves(actual[:tt.limit]), "board: %v", b)
	}
}

package turochamp_test

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/turochamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQuiescence(t *testing.T) {
	tests := []struct {
		fen      string
		moves    []string
		nodes    uint64
		expected eval.Score // not really in centi-pawns
	}{
		{fen.Initial, nil, 1, eval.HeuristicScore(0.8)},
		{fen.Initial, []string{"d2d4"}, 1, eval.HeuristicScore(0.8)},                      // no captures, so equal to startpos w/ black
		{"kr6/pppppppp/8/8/8/8/6Q1/7K w - - 0 1", nil, 2, eval.HeuristicScore(-1200.23)},  // 1 undefended, 1 defended
		{"k7/pppppp1p/6b1/7P/8/8/8/7K w - - 0 1", nil, 3, eval.HeuristicScore(-10500.19)}, // lower value, 1x recapture, 1xcutoff
		{"k7/pppppnpn/8/n6Q/8/8/8/7K w - - 0 1", nil, 4, eval.HeuristicScore(-1200.20)},   // 3x undefended Knight
		{"k7/p2ppnpn/8/r6Q/8/8/8/7K w - - 0 1", nil, 4, eval.HeuristicScore(-0.18)},       //  3x undefended, picks Rook for equal material
		{"2b2rk1/r1Pp2p1/ppn1p3/q3N1Bp/3P4/2NQR2P/PPP2PP1/R5K1 b - - 4 18", nil, 2, eval.HeuristicScore(-1122.23)},
	}

	qs := turochamp.Quiescence{Eval: turochamp.Eval{}}

	for _, tt := range tests {
		pos, turn, np, fm, err := fen.Decode(tt.fen)
		require.NoError(t, err)

		b := board.NewBoard(board.NewZobristTable(0), pos, turn, np, fm)
		for _, m := range tt.moves {
			move, err := board.ParseMove(m)
			require.NoError(t, err)
			b.PushMove(move)
		}

		nodes, actual := qs.QuietSearch(context.Background(), b, eval.NegInfScore, eval.InfScore, make(chan struct{}))
		assert.Equal(t, nodes, tt.nodes, "failed: %v", tt.fen)
		assert.Equal(t, actual, tt.expected, "failed: %v", tt.fen)
	}
}

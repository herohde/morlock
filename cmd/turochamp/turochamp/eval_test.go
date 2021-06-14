package turochamp_test

import (
	"context"
	"github.com/herohde/morlock/cmd/turochamp/turochamp"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMaterial(t *testing.T) {
	tests := []struct {
		fen      string
		expected eval.Pawns
	}{
		{fen.Initial, 0},
		{"k7/8/8/8/8/8/8/7K w - - 0 1", 0},
		{"kq6/8/8/8/8/8/8/7K w - - 0 1", -20},
		{"kq6/8/8/8/8/8/8/6PK w - - 0 1", -10},
		{"kb6/8/8/8/8/8/8/6QK w - - 0 1", 2.8571},
		{"kq6/8/8/8/8/8/8/5PQK w - - 0 1", 1.1},
		{"kq6/8/8/8/8/8/8/4PPQK w - - 0 1", 1.2},
		{"kq6/8/8/8/8/8/8/3PPPQK w - - 0 1", 1.3},
		{"kq6/8/8/8/8/8/8/2PPPPQK w - - 0 1", 1.4},
		{"kq6/8/8/8/8/8/8/1PPPPPQK w - - 0 1", 1.5},
		{"kq6/8/8/8/8/8/8/PPPPPPQK w - - 0 1", 1.6},
		{"kqqq4/8/8/8/8/8/8/1PPPQQQK w - - 0 1", 1.1},
		{"rnbqkbnr/ppppppp1/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 1.025},

		{"k7/8/8/8/8/8/8/7K b - - 0 1", 0},
		{"kq6/8/8/8/8/8/8/7K b - - 0 1", 20},
		{"kq6/8/8/8/8/8/8/6PK b - - 0 1", 10},
		{"kb6/8/8/8/8/8/8/6QK b - - 0 1", -2.8571},
		{"kq6/8/8/8/8/8/8/5PQK b - - 0 1", -1.10},
		{"kq6/8/8/8/8/8/8/4PPQK b - - 0 1", -1.20},
		{"kq6/8/8/8/8/8/8/3PPPQK b - - 0 1", -1.30},

		{"rnbqkbnr/pppppppp/8/8/8/8/8/7K b - - 0 1", 82},
		{"rnbqkbnr/qqqqqqqq/8/8/8/8/8/7K b - - 0 1", 226},
		{"rnbqkbnr/qqqqqqqq/8/8/8/8/8/6PK b - - 0 1", 113},
		{"rnbqkbnr/qqqqqqqq/8/8/8/8/8/5PPK b - - 0 1", 56.5},
		{"rnbqkbnr/qqqqqqqq/8/8/8/8/8/4PPPK b - - 0 1", 37.6667},
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.fen)
		require.NoError(t, err)

		actual := turochamp.Material{}.Evaluate(context.Background(), b)
		assert.Equal(t, actual.String(), tt.expected.String())
	}
}

func TestPositionPlay(t *testing.T) {
	tests := []struct {
		fen   string
		moves []string
		w, b  eval.Pawns
	}{
		{fen.Initial, nil, 10.20, 10.20},
		{fen.Initial, []string{"e2e3"}, 14.60, 10.20}, // +4.4 (ignores opponent progress)
		{fen.Initial, []string{"e2e4"}, 14.40, 10.20}, // +4.2
		{fen.Initial, []string{"d2d3"}, 12.90, 10.20},
		{fen.Initial, []string{"d2d4"}, 13.50, 10.20},

		{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1", nil, 10.20, 10.20},
		{"kr6/pppppppp/8/8/8/8/PPPPPPPP/6RK w - - 0 1", nil, 4.00, 4.00},
		{"k7/8/8/8/8/8/8/7K w - - 0 1", nil, -2.90, -2.90},
		{"kb6/8/8/8/8/8/8/6QK w - - 0 1", nil, 1.8, 0.90},
		{"rnbqkbnr/qqqqqqqq/8/8/8/8/8/7K b - - 0 1", nil, -4.40, 42.10},
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.fen, tt.moves...)
		require.NoError(t, err)

		actual := turochamp.PositionPlay(b, board.White)
		assert.Equalf(t, actual.String(), tt.w.String(), "white: %v", b)

		actual2 := turochamp.PositionPlay(b, board.Black)
		assert.Equalf(t, actual2.String(), tt.b.String(), "black: %v", b)
	}
}

func TestEval(t *testing.T) {
	tests := []struct {
		fen      string
		expected eval.Pawns // not really in centi-pawns
	}{
		{fen.Initial, 0},
		{"k7/8/8/8/8/8/8/7K w - - 0 1", 0},
		{"kb6/8/8/8/8/8/8/6QK w - - 0 1", 2860.09},
		{"kb6/8/8/8/8/8/8/6QK b - - 0 1", -2860.09},
		{"rnbqkbnr/qqqqqqqq/8/8/8/8/8/7K b - - 0 1", 226004.66},

		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 0},
		{"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", -0.01},
		{"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 1018.69},
		{"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 0.34},
		{"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 0},
		{"r4rk1/2p1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 1020.10},
		{"r4rk1/4qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 1050.07},
		{"r4rk1/4q1pp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 1081.59},
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.fen)
		require.NoError(t, err)

		actual := turochamp.Eval{}.Evaluate(context.Background(), b)
		assert.Equal(t, actual.String(), tt.expected.String())
	}
}

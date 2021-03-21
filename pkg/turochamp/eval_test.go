package turochamp_test

import (
	"context"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/turochamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMaterial(t *testing.T) {
	tests := []struct {
		fen      string
		expected eval.Score
	}{
		{fen.Initial, eval.HeuristicScore(0)},
		{"k7/8/8/8/8/8/8/7K w - - 0 1", eval.HeuristicScore(0)},
		{"kq6/8/8/8/8/8/8/7K w - - 0 1", eval.HeuristicScore(-20)},
		{"kq6/8/8/8/8/8/8/6PK w - - 0 1", eval.HeuristicScore(-10)},
		{"kb6/8/8/8/8/8/8/6QK w - - 0 1", eval.HeuristicScore(2.8571)},
		{"kq6/8/8/8/8/8/8/5PQK w - - 0 1", eval.HeuristicScore(1.1)},
		{"kq6/8/8/8/8/8/8/4PPQK w - - 0 1", eval.HeuristicScore(1.2)},
		{"kq6/8/8/8/8/8/8/3PPPQK w - - 0 1", eval.HeuristicScore(1.3)},
		{"kq6/8/8/8/8/8/8/2PPPPQK w - - 0 1", eval.HeuristicScore(1.4)},
		{"kq6/8/8/8/8/8/8/1PPPPPQK w - - 0 1", eval.HeuristicScore(1.5)},
		{"kq6/8/8/8/8/8/8/PPPPPPQK w - - 0 1", eval.HeuristicScore(1.6)},
		{"kqqq4/8/8/8/8/8/8/1PPPQQQK w - - 0 1", eval.HeuristicScore(1.1)},
		{"rnbqkbnr/ppppppp1/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", eval.HeuristicScore(1.025)},

		{"k7/8/8/8/8/8/8/7K b - - 0 1", eval.HeuristicScore(0)},
		{"kq6/8/8/8/8/8/8/7K b - - 0 1", eval.HeuristicScore(20)},
		{"kq6/8/8/8/8/8/8/6PK b - - 0 1", eval.HeuristicScore(10)},
		{"kb6/8/8/8/8/8/8/6QK b - - 0 1", eval.HeuristicScore(-2.8571)},
		{"kq6/8/8/8/8/8/8/5PQK b - - 0 1", eval.HeuristicScore(-1.10)},
		{"kq6/8/8/8/8/8/8/4PPQK b - - 0 1", eval.HeuristicScore(-1.20)},
		{"kq6/8/8/8/8/8/8/3PPPQK b - - 0 1", eval.HeuristicScore(-1.30)},

		{"rnbqkbnr/pppppppp/8/8/8/8/8/7K b - - 0 1", eval.HeuristicScore(82)},
		{"rnbqkbnr/qqqqqqqq/8/8/8/8/8/7K b - - 0 1", eval.HeuristicScore(226)},
		{"rnbqkbnr/qqqqqqqq/8/8/8/8/8/6PK b - - 0 1", eval.HeuristicScore(113)},
		{"rnbqkbnr/qqqqqqqq/8/8/8/8/8/5PPK b - - 0 1", eval.HeuristicScore(56.5)},
		{"rnbqkbnr/qqqqqqqq/8/8/8/8/8/4PPPK b - - 0 1", eval.HeuristicScore(37.6667)},
	}

	for _, tt := range tests {
		pos, turn, _, _, err := fen.Decode(tt.fen)
		require.NoError(t, err)

		actual := turochamp.Material{}.Evaluate(context.Background(), pos, turn)
		assert.Equal(t, actual.String(), tt.expected.String())
	}
}

func TestPositionPlay(t *testing.T) {
	tests := []struct {
		fen      string
		expected eval.Score
	}{
		{fen.Initial, eval.HeuristicScore(8)},
		{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1", eval.HeuristicScore(8)},
		{"kr6/pppppppp/8/8/8/8/PPPPPPPP/6RK w - - 0 1", eval.HeuristicScore(2.3)},
		{"kr6/pppppppp/8/8/8/8/PPPPPPPP/6RK b - - 0 1", eval.HeuristicScore(2.3)},
		{"k7/8/8/8/8/8/8/7K w - - 0 1", eval.HeuristicScore(-3)},
		{"k7/8/8/8/8/8/8/7K b - - 0 1", eval.HeuristicScore(-3)},
		{"kb6/8/8/8/8/8/8/6QK w - - 0 1", eval.HeuristicScore(2.5)},
		{"kb6/8/8/8/8/8/8/6QK b - - 0 1", eval.HeuristicScore(0.6)},
		{"rnbqkbnr/qqqqqqqq/8/8/8/8/8/7K b - - 0 1", eval.HeuristicScore(40.4)},
	}

	for _, tt := range tests {
		pos, turn, _, _, err := fen.Decode(tt.fen)
		require.NoError(t, err)

		actual := turochamp.PositionPlay{}.Evaluate(context.Background(), pos, turn)
		assert.Equal(t, actual.String(), tt.expected.String())
	}
}

func TestEval(t *testing.T) {
	tests := []struct {
		fen      string
		expected eval.Score // not really in centi-pawns
	}{
		{fen.Initial, eval.HeuristicScore(0.8)},
		{"k7/8/8/8/8/8/8/7K w - - 0 1", eval.HeuristicScore(-0.3)},
		{"kb6/8/8/8/8/8/8/6QK w - - 0 1", eval.HeuristicScore(2860.25)},
		{"kb6/8/8/8/8/8/8/6QK b - - 0 1", eval.HeuristicScore(-2859.94)},
		{"rnbqkbnr/qqqqqqqq/8/8/8/8/8/7K b - - 0 1", eval.HeuristicScore(226004.05)},

		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", eval.HeuristicScore(2.34)},
		{"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", eval.HeuristicScore(0.38)},
		{"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", eval.HeuristicScore(1020.97)},
		{"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", eval.HeuristicScore(1.97)},
		{"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", eval.HeuristicScore(2.65)},
		{"r4rk1/2p1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", eval.HeuristicScore(1022.65)},
		{"r4rk1/4qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", eval.HeuristicScore(1052.65)},
		{"r4rk1/4q1pp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", eval.HeuristicScore(1082.67)},
	}

	for _, tt := range tests {
		pos, turn, _, _, err := fen.Decode(tt.fen)
		require.NoError(t, err)

		actual := turochamp.Eval{}.Evaluate(context.Background(), pos, turn)
		assert.Equal(t, actual.String(), tt.expected.String())
	}
}

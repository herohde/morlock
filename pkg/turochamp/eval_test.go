package turochamp_test

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/turochamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMaterial(t *testing.T) {
	tests := []struct {
		fen      string
		expected board.Score // x100
	}{
		{fen.Initial, 0},
		{"k7/8/8/8/8/8/8/7K w - - 0 1", 0},
		{"kq6/8/8/8/8/8/8/7K w - - 0 1", -1000},
		{"kq6/8/8/8/8/8/8/6PK w - - 0 1", -1000},
		{"kb6/8/8/8/8/8/8/6QK w - - 0 1", 285},
		{"kq6/8/8/8/8/8/8/5PQK w - - 0 1", 110},
		{"kq6/8/8/8/8/8/8/4PPQK w - - 0 1", 120},
		{"kq6/8/8/8/8/8/8/3PPPQK w - - 0 1", 130},
		{"kq6/8/8/8/8/8/8/2PPPPQK w - - 0 1", 140},
		{"kq6/8/8/8/8/8/8/1PPPPPQK w - - 0 1", 150},
		{"kq6/8/8/8/8/8/8/PPPPPPQK w - - 0 1", 160},
		{"kqqq4/8/8/8/8/8/8/1PPPQQQK w - - 0 1", 110},
		{"rnbqkbnr/ppppppp1/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 102},

		{"k7/8/8/8/8/8/8/7K b - - 0 1", 0},
		{"kq6/8/8/8/8/8/8/7K b - - 0 1", -1000},
		{"kq6/8/8/8/8/8/8/6PK b - - 0 1", -1000},
		{"kb6/8/8/8/8/8/8/6QK b - - 0 1", 285},
		{"kq6/8/8/8/8/8/8/5PQK b - - 0 1", 110},
		{"kq6/8/8/8/8/8/8/4PPQK b - - 0 1", 120},
		{"kq6/8/8/8/8/8/8/3PPPQK b - - 0 1", 130},
	}

	for _, tt := range tests {
		pos, turn, _, _, err := fen.Decode(tt.fen)
		require.NoError(t, err)

		actual := turochamp.Material{}.Evaluate(context.Background(), pos, turn)
		assert.Equal(t, actual, tt.expected)
	}
}

func TestPositionPlay(t *testing.T) {
	tests := []struct {
		fen      string
		expected board.Score // x10
	}{
		{fen.Initial, 80},
		{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1", -80},
		{"kr6/pppppppp/8/8/8/8/PPPPPPPP/6RK w - - 0 1", 23},
		{"kr6/pppppppp/8/8/8/8/PPPPPPPP/6RK b - - 0 1", -23},
		{"k7/8/8/8/8/8/8/7K w - - 0 1", -29},
		{"k7/8/8/8/8/8/8/7K b - - 0 1", 29},
		{"kb6/8/8/8/8/8/8/6QK w - - 0 1", 24},
		{"kb6/8/8/8/8/8/8/6QK b - - 0 1", -6},
	}

	for _, tt := range tests {
		pos, turn, _, _, err := fen.Decode(tt.fen)
		require.NoError(t, err)

		actual := turochamp.PositionPlay{}.Evaluate(context.Background(), pos, turn)
		assert.Equal(t, actual, tt.expected)
	}
}

func TestEval(t *testing.T) {
	tests := []struct {
		fen      string
		expected board.Score // not really in centi-pawns
	}{
		{fen.Initial, 80},
		{"k7/8/8/8/8/8/8/7K w - - 0 1", -29},
		{"kb6/8/8/8/8/8/8/6QK w - - 0 1", 285 + 10000 + 24},
		{"kb6/8/8/8/8/8/8/6QK b - - 0 1", 285 + 10000 - 6},
	}

	for _, tt := range tests {
		pos, turn, _, _, err := fen.Decode(tt.fen)
		require.NoError(t, err)

		actual := turochamp.Eval{}.Evaluate(context.Background(), pos, turn)
		assert.Equal(t, actual, tt.expected)
	}
}

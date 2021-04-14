package engine_test

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBook(t *testing.T) {
	ctx := context.Background()

	book, err := engine.NewBook([]engine.Line{
		{"e2e4", "d7d5", "d2d4"},
		{"e2e4", "d7d6"},
		{"d2d4", "d7d6"},
	})
	require.NoError(t, err)

	tests := []struct {
		pos   string
		moves string
	}{
		{fen.Initial, "d2-d4 e2-e4"},
		{"rnbqkbnr/pppppppp/8/8/3P4/8/PPP1PPPP/RNBQKBNR b KQkq d3 0 1", "d7-d6"},
	}

	for _, tt := range tests {
		list, err := book.Find(ctx, tt.pos)
		assert.NoError(t, err)
		assert.Equal(t, board.PrintMoves(list), tt.moves)
	}
}

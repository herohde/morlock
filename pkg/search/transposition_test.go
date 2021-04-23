package search_test

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestTranspositionTable(t *testing.T) {
	ctx := context.Background()

	// (1) Test that we use MSB for size only.

	tt := search.NewTranspositionTable(ctx, 0x1000)
	assert.Equal(t, tt.Size(), uint64(0x1000))
	tt2 := search.NewTranspositionTable(ctx, 0x1f00)
	assert.Equal(t, tt2.Size(), uint64(0x1000))

	// (2) Test read/write.

	a := board.ZobristHash(rand.Uint64())

	_, _, _, _, notok := tt.Read(a)
	assert.False(t, notok)

	m := board.Move{From: board.G4, To: board.G8, Promotion: board.Queen}
	s := eval.HeuristicScore(2)
	_ = tt.Write(a, search.ExactBound, 5, 2, s, m)

	bound, depth, score, move, ok := tt.Read(a)
	assert.True(t, ok)
	assert.Equal(t, bound, search.ExactBound)
	assert.Equal(t, depth, 2)
	assert.Equal(t, score, s)
	assert.Equal(t, move, m)

	_, _, _, _, notok = tt.Read(a ^ 0xff0000)
	assert.False(t, notok)

	// (2) Test replacement.

	norepl := tt.Write(a, search.ExactBound, 2, 3, eval.HeuristicScore(5), m)
	assert.False(t, norepl)

	repl := tt.Write(a, search.ExactBound, 4, 3, eval.HeuristicScore(5), m)
	assert.True(t, repl)
}

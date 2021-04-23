package search_test

import (
	"context"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAlphaBeta(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		fen      string
		depth    int
		expected eval.Score
	}{
		{fen.Initial, 4, eval.ZeroScore},
		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 4, eval.ZeroScore},
		{"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 4, eval.ZeroScore},
		{"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 4, eval.HeuristicScore(-6)},
		{"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 4, eval.HeuristicScore(2)},
		{"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 4, eval.HeuristicScore(-1)},

		{"k7/7R/6R1/8/8/8/8/7K w - - 0 1", 1, eval.HeuristicScore(10)},
		{"k7/7R/6R1/8/8/8/8/7K w - - 0 1", 2, eval.MateInXScore(1)},
		{"k7/7R/6R1/8/8/8/8/7K w - - 0 1", 3, eval.MateInXScore(1)},
		{"k7/7R/7R/8/8/8/8/7K w - - 0 1", 4, eval.MateInXScore(3)},
	}

	minimax := search.Minimax{Eval: eval.Material{}}
	pvs := search.AlphaBeta{Eval: search.ZeroPly{Eval: eval.Material{}}}

	t.Run("correctness", func(t *testing.T) {
		for _, tt := range tests {
			b, err := fen.NewBoard(tt.fen)
			require.NoError(t, err)

			n, actual, _, _ := pvs.Search(ctx, search.EmptyContext, b, tt.depth, make(chan struct{}))
			assert.Lessf(t, n, uint64(16000), "too many nodes: %v", tt.fen)
			assert.Equalf(t, actual, tt.expected, "failed: %v", tt.fen)
		}
	})

	t.Run("minimax", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping minimax comparison test")
		}

		for _, tt := range tests {
			b, err := fen.NewBoard(tt.fen)
			require.NoError(t, err)

			n, actual, _, _ := pvs.Search(ctx, search.EmptyContext, b, tt.depth, make(chan struct{}))
			n2, actual2, _, _ := pvs.Search(ctx, &search.Context{TT: search.NewTranspositionTable(ctx, 64<<20)}, b, tt.depth, make(chan struct{}))
			m, expected, _, _ := minimax.Search(ctx, search.EmptyContext, b, tt.depth, make(chan struct{}))
			t.Logf("POS: %v; NODES: %v /tt:%v (minimax %v)", tt.fen, n, n2, m)

			assert.LessOrEqualf(t, n, m, "more than minimax nodes: %v", tt.fen)
			assert.Equalf(t, actual, actual2, "tt failed: %v", tt.fen)
			assert.Equalf(t, actual, expected, "failed: %v", tt.fen)
		}
	})
}

func BenchmarkAlphaBeta1(b *testing.B) {
	pos, _ := fen.NewBoard("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	s := search.AlphaBeta{Eval: search.ZeroPly{Eval: eval.Material{}}}

	for i := 0; i < b.N; i++ {
		s.Search(context.Background(), &search.Context{TT: search.NoTranspositionTable{}}, pos, 4, make(chan struct{}))
	}
}

func BenchmarkAlphaBeta1_TT(b *testing.B) {
	ctx := context.Background()
	pos, _ := fen.NewBoard("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	s := search.AlphaBeta{Eval: search.ZeroPly{Eval: eval.Material{}}}

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tt := search.NewTranspositionTable(ctx, 64<<20)
		b.StartTimer()
		s.Search(ctx, &search.Context{TT: tt}, pos, 4, make(chan struct{}))
	}
}

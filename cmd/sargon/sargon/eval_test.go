package sargon_test

import (
	"context"
	"github.com/herohde/morlock/cmd/sargon/sargon"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDevelopment(t *testing.T) {
	tests := []struct {
		moves    []string
		expected eval.Pawns
	}{
		{[]string{}, 0},
		{[]string{"e2e4", "e7e5"}, 0},
		{[]string{"g1f3", "e7e5"}, 2},
		{[]string{"e2e4", "e7e5", "d1e2", "d7d6"}, -2},
		{[]string{"e2e4", "e7e5", "f1e2", "d7d6"}, 2},
		{[]string{"e2e4", "e7e5", "e1e2", "d7d6"}, -2},
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(fen.Initial, tt.moves...)
		require.NoError(t, err)

		actual := sargon.Development(context.Background(), b)
		assert.Equal(t, actual, tt.expected, "failed: %v", b.Position())
	}
}

func TestMobility(t *testing.T) {
	tests := []struct {
		fen      string
		expected eval.Pawns
	}{
		{fen.Initial, 0},
		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 8},
		{"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", -2},
		{"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 5},
		{"k7/8/8/8/8/8/8/6K1 w - - 0 1", 2},
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.fen)
		require.NoError(t, err)

		pins := sargon.FindKingQueenPins(b.Position())
		actual := sargon.Mobility(context.Background(), b, pins)
		assert.Equal(t, actual, tt.expected, "failed: %v", b.Position())
	}
}

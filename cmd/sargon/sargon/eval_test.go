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

func TestMaterial(t *testing.T) {
	tests := []struct {
		fen      string
		moves    []string
		expected eval.Pawns
	}{
		{fen.Initial, []string{}, 0},
		{"kr4QR/pr6/2B5/8/8/8/8/7K w - - 0 1", []string{}, 8}, // +6: +5, -4*3/4=-3
		{"kr4QR/pr6/2B5/8/8/8/8/7K b - - 0 1", []string{}, -5.75},
		// In game37, Qh1 seems broken after the below position. Maybe when ply0 is different color?
		{"r7/2p1k1pp/Q6n/p2qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 w - - 7 18", []string{"a6b5"}, -6},
		{"r7/2p1k1pp/7n/pQ1qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 b - - 8 18", []string{}, -12.75},
		{"r7/2p1k1pp/Q6n/p2qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 w - - 7 18", []string{"a6b5", "d5h1"}, 16},
		{"r7/2p1k1pp/Q6n/p2qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 w - - 7 18", []string{"a6b5", "d5b5"}, 8}, // <- clearly better
		// In game38, f5c2 seems broken. Bishop is moving into a losing exchange.
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5"}, 1},
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5", "g3e5"}, -3.75},
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5", "g3e5", "f5c2"}, 6},
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5", "g3e5", "d6c6"}, 5},
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.fen, tt.moves...)
		require.NoError(t, err)

		pins := sargon.FindKingQueenPins(b.Position())
		actual := sargon.Material(context.Background(), b, pins, 0)
		assert.Equal(t, actual, tt.expected, "failed: %v", b.Position())
	}
}

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
		{"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", -1},
		{"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 6},
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

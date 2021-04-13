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

func TestPoints(t *testing.T) {
	tests := []struct {
		fen      string
		moves    []string
		expected eval.Pawns
	}{
		{fen.Initial, []string{}, 0},
		{"kr4QR/pr6/2B5/8/8/8/8/7K w - - 0 1", []string{}, 32.36},
		{"kr4QR/pr6/2B5/8/8/8/8/7K b - - 0 1", []string{}, -23.36},
		// In game37, Qh1 seems broken after the below position. Maybe when ply0 is different color?
		{"r7/2p1k1pp/Q6n/p2qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 w - - 7 18", []string{"a6b5"}, -23.96},
		{"r7/2p1k1pp/7n/pQ1qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 b - - 8 18", []string{}, -50.96},
		{"r7/2p1k1pp/Q6n/p2qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 w - - 7 18", []string{"a6b5", "d5h1"}, 64.03},
		{"r7/2p1k1pp/Q6n/p2qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 w - - 7 18", []string{"a6b5", "d5b5"}, 31.97}, // <- clearly better
		// In game38, f5c2 seems broken. Bishop is moving into a losing exchange.
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5"}, 3.83},
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5", "g3e5"}, -14.85},
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5", "g3e5", "f5c2"}, 23.88},
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5", "g3e5", "d6c6"}, 19.89}, // <- clearly better
		// In game41, Ne3 seems broken.
		{"rnb1k2r/ppppbppp/3q4/8/2BBP1n1/5N1P/PPP2PP1/RN1Q1RK1 b kq - 0 8", []string{"g4e3"}, 2.92},
		{"rnb1k2r/ppppbppp/3q4/8/2BBP1n1/5N1P/PPP2PP1/RN1Q1RK1 b kq - 0 8", []string{"g4h6"}, 3.96},
		// In game 43, e2e4 seems broken when B is en prise.
		{"rnbqkbnr/ppppp1pp/8/5pB1/3P4/8/PPP1PPPP/RN1QKBNR b KQkq - 1 2", []string{"h7h6"}, 0.15},
		{"rnbqkbnr/ppppp1pp/8/5pB1/3P4/8/PPP1PPPP/RN1QKBNR b KQkq - 1 2", []string{"h7h6", "e2e4"}, 11.76},
		{"rnbqkbnr/ppppp1pp/8/5pB1/3P4/8/PPP1PPPP/RN1QKBNR b KQkq - 1 2", []string{"h7h6", "g5h4"}, -0.13}, // <- clearly better
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.fen, tt.moves...)
		require.NoError(t, err)

		actual := (&sargon.Points{}).Evaluate(context.Background(), b)
		assert.Equal(t, actual, tt.expected, "failed: %v", b.Position())
	}
}

func BenchmarkPoints1(b *testing.B) {
	pos, _ := fen.NewBoard("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	points := &sargon.Points{}

	for i := 0; i < b.N; i++ {
		points.Evaluate(context.Background(), pos)
	}
}

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
		// In game41, Ne3 seems broken.
		{"rnb1k2r/ppppbppp/3q4/8/2BBP1n1/5N1P/PPP2PP1/RN1Q1RK1 b kq - 0 8", []string{"g4e3"}, 0.75},
		{"rnb1k2r/ppppbppp/3q4/8/2BBP1n1/5N1P/PPP2PP1/RN1Q1RK1 b kq - 0 8", []string{"g4h6"}, 1},
		// In game 43, e2e4 seems broken when B is en prise.
		{"rnbqkbnr/ppppp1pp/8/5pB1/3P4/8/PPP1PPPP/RN1QKBNR b KQkq - 1 2", []string{"h7h6"}, 0},
		{"rnbqkbnr/ppppp1pp/8/5pB1/3P4/8/PPP1PPPP/RN1QKBNR b KQkq - 1 2", []string{"h7h6", "e2e4"}, 3},
		{"rnbqkbnr/ppppp1pp/8/5pB1/3P4/8/PPP1PPPP/RN1QKBNR b KQkq - 1 2", []string{"h7h6", "g5h4"}, 0},
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.fen, tt.moves...)
		require.NoError(t, err)

		pins := sargon.FindKingQueenPins(b.Position())
		actual := sargon.Material(context.Background(), b, pins)
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
		{"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 7},
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

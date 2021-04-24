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
		{"kr5R/8/8/8/8/8/8/7K w - - 0 1", []string{}, 6.12},
		{"kr5R/8/8/8/8/8/8/7K b - - 0 1", []string{}, 29.88},
		{"kr4QR/pr6/2B5/8/8/8/8/7K w - - 0 1", []string{}, 66.39},
		{"kr4QR/pr6/2B5/8/8/8/8/7K b - - 0 1", []string{}, -8.39},
		// In game37, Qh1 seems broken after the below position. Maybe when ply0 is different color?
		{"r7/2p1k1pp/Q6n/p2qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 w - - 7 18", []string{"a6b5"}, -30.06},
		{"r7/2p1k1pp/7n/pQ1qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 b - - 8 18", []string{}, -30.06},
		{"r7/2p1k1pp/Q6n/p2qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 w - - 7 18", []string{"a6b5", "d5h1"}, 96.16},
		{"r7/2p1k1pp/Q6n/p2qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 w - - 7 18", []string{"a6b5", "d5b5"}, 38.07}, // <- clearly better
		// In game38, f5c2 seems broken. Bishop is moving into a losing exchange.
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5"}, 4.01},
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5", "g3e5"}, -16.99},
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5", "g3e5", "f5c2"}, 45.01},
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5", "g3e5", "d6c6"}, 38.02}, // <- clearly better
		// In game41, Ne3 seems broken.
		{"rnb1k2r/ppppbppp/3q4/8/2BBP1n1/5N1P/PPP2PP1/RN1Q1RK1 b kq - 0 8", []string{"g4e3"}, 19.98},
		{"rnb1k2r/ppppbppp/3q4/8/2BBP1n1/5N1P/PPP2PP1/RN1Q1RK1 b kq - 0 8", []string{"g4h6"}, 7.03},
		// In game 43, e2e4 seems broken when B is en prise.
		{"rnbqkbnr/ppppp1pp/8/5pB1/3P4/8/PPP1PPPP/RN1QKBNR b KQkq - 1 2", []string{"h7h6"}, 6.1},
		{"rnbqkbnr/ppppp1pp/8/5pB1/3P4/8/PPP1PPPP/RN1QKBNR b KQkq - 1 2", []string{"h7h6", "e2e4"}, 19.85},
		{"rnbqkbnr/ppppp1pp/8/5pB1/3P4/8/PPP1PPPP/RN1QKBNR b KQkq - 1 2", []string{"h7h6", "g5h4"}, -6.09}, // <- clearly better
		// Sargon 1ply moved into pawn en prise.
		{"rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2", []string{"f7f5"}, 4.04},
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
		ptsck    bool
	}{
		{fen.Initial, []string{}, 0, false},
		{fen.Initial, []string{"e2e4", "e7e5", "g1f3", "f7f5"}, 1, true},
		{fen.Initial, []string{"e2e4", "e7e5", "g1f3", "b8c6"}, 0, false},
		{"kr5R/8/8/8/8/8/8/7K w - - 0 1", []string{}, 0, false}, // white will move en prise rook
		{"kr5R/8/8/8/8/8/8/7K b - - 0 1", []string{}, 9, false},
		{"kr4QR/pr6/2B5/8/8/8/8/7K w - - 0 1", []string{}, 15, false},
		{"kr4QR/pr6/2B5/8/8/8/8/7K b - - 0 1", []string{}, -0.5, false},
		// In game37, Qh1 seems broken after the below position.
		{"r7/2p1k1pp/Q6n/p2qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 w - - 7 18", []string{"a6b5"}, -6, false},
		{"r7/2p1k1pp/7n/pQ1qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 b - - 8 18", []string{}, -6, false}, // == above, given last move irrelevant
		{"r7/2p1k1pp/Q6n/p2qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 w - - 7 18", []string{"a6b5", "d5h1"}, 24, true},
		{"r7/2p1k1pp/Q6n/p2qPp2/3p4/N5P1/PPP1PP1P/3R1RK1 w - - 7 18", []string{"a6b5", "d5b5"}, 8, false}, // <- clearly better
		// In game38, f5c2 seems broken. Bishop is moving into a losing exchange.
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5"}, 1, true},
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5", "g3e5"}, -4.5, false}, // loss of (rook-1)/2
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5", "g3e5", "f5c2"}, 11, false},
		{"rn2kbnr/ppp1pp2/3q3p/3p1bp1/3P4/2N2NB1/PPP1PPPP/R2QKB1R b KQkq - 1 6", []string{"e7e5", "g3e5", "d6c6"}, 9, false},
		// In game41, Ne3 seems broken.
		{"rnb1k2r/ppppbppp/3q4/8/2BBP1n1/5N1P/PPP2PP1/RN1Q1RK1 b kq - 0 8", []string{"g4e3"}, 5, true},
		{"rnb1k2r/ppppbppp/3q4/8/2BBP1n1/5N1P/PPP2PP1/RN1Q1RK1 b kq - 0 8", []string{"g4h6"}, 1, false},
		// In game 43, e2e4 seems broken when B is en prise.
		{"rnbqkbnr/ppppp1pp/8/5pB1/3P4/8/PPP1PPPP/RN1QKBNR b KQkq - 1 2", []string{"h7h6"}, 0, false},
		{"rnbqkbnr/ppppp1pp/8/5pB1/3P4/8/PPP1PPPP/RN1QKBNR b KQkq - 1 2", []string{"h7h6", "e2e4"}, 5, true},
		{"rnbqkbnr/ppppp1pp/8/5pB1/3P4/8/PPP1PPPP/RN1QKBNR b KQkq - 1 2", []string{"h7h6", "g5h4"}, 0, false},
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.fen, tt.moves...)
		require.NoError(t, err)

		pins := sargon.FindKingQueenPins(b.Position())
		actual, ptschk := sargon.Material(context.Background(), b, pins)
		assert.Equal(t, actual, tt.expected, "failed: %v", b.Position())
		assert.Equal(t, ptschk, tt.ptsck, "failed ptschk: %v", b.Position())
	}
}

func TestDevelopment(t *testing.T) {
	tests := []struct {
		moves    []string
		expected eval.Pawns
	}{
		{[]string{}, 0},
		{[]string{"e2e4", "e7e5"}, 0},
		{[]string{"e2e4", "e7e5", "g1f3"}, -2},
		{[]string{"e2e4", "e7e5", "g1f3", "f7f5"}, 2},
		{[]string{"e2e4", "e7e5", "g1f3", "b8c6"}, 0},
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
		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 9},
		{"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", -1},
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

package sargon_test

import (
	"github.com/herohde/morlock/cmd/sargon/sargon"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExchange(t *testing.T) {
	tests := []struct {
		fen      string
		sq       board.Square
		expected eval.Pawns
	}{
		{fen.Initial, board.E2, 0},
		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", board.C3, -2},
		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", board.E6, 0},
		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", board.D5, 0},
		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", board.G2, 0},
		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", board.A6, 3},
		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", board.D7, 0},
		{"kr4QR/pr6/2B5/8/8/8/8/7K w - - 0 1", board.B7, 2},
		{"kr4QR/pr6/2B5/8/8/8/8/7K w - - 0 1", board.B8, 5},
		{"kr4QR/pr6/2B5/8/8/8/8/7K b - - 0 1", board.B8, -5},
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.fen)
		require.NoError(t, err)

		pins := sargon.FindKingQueenPins(b.Position())
		actual := sargon.Exchange(b.Position(), pins, b.Turn(), tt.sq)
		assert.Equal(t, actual, tt.expected, "failed at %v: %v", tt.sq, b.Position())
	}
}

var empty sargon.Pins = map[board.Square][]board.Square{}

func TestFindKingQueenPins(t *testing.T) {
	tests := []struct {
		fen      string
		expected sargon.Pins
	}{
		{fen.Initial, empty},
		{"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", empty},
		{"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", map[board.Square][]board.Square{
			board.F4: {board.B4},
			board.B5: {board.H5},
		}},
		{"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", empty},
		{"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", empty}, // no Q on Q
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.fen)
		require.NoError(t, err)

		actual := sargon.FindKingQueenPins(b.Position())
		t.Logf("PINS %v: %v", tt.fen, actual)

		assert.Equal(t, len(actual), len(tt.expected))
		for sq, att := range actual {
			assert.Equal(t, len(att), len(tt.expected[sq]))
		}
	}
}

func TestFindAttackers(t *testing.T) {
	tests := []struct {
		fen      string
		sq       board.Square
		expected string
	}{
		{
			"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
			board.C3,
			"--------/--------/--------/--------/-X------/-----X--/-X-X----/--------",
		},
		{
			"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
			board.D7,
			"----X---/----X---/-X---X--/----X---/--------/--------/--------/--------",
		},
		{
			"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
			board.D5,
			"--------/--------/-X--XX--/--------/----X---/--X--X--/--------/--------",
		},
		{
			"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
			board.A5,
			"--------/--------/--------/--------/--------/--------/--------/--------",
		},
		{
			"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
			board.G4, // B behind Q
			"--------/--------/-----X--/----X---/--------/-----X--/----X---/--------",
		},
	}

	for _, tt := range tests {
		b, err := fen.NewBoard(tt.fen)
		require.NoError(t, err)

		pins := sargon.FindKingQueenPins(b.Position())
		attackers := sargon.FindAttackers(b.Position(), pins, tt.sq, board.White)
		attackers = append(attackers, sargon.FindAttackers(b.Position(), pins, tt.sq, board.Black)...)

		actual := attSquares(attackers)
		assert.Equal(t, actual.String(), tt.expected)
	}
}

func attSquares(list []*sargon.Attacker) board.Bitboard {
	bb := board.EmptyBitboard
	for _, att := range list {
		bb |= attSquare(att)
	}
	return bb
}

func attSquare(att *sargon.Attacker) board.Bitboard {
	if att == nil {
		return board.EmptyBitboard
	}
	return attSquare(att.Behind) | board.BitMask(att.Piece.Square)
}

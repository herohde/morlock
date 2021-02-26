package fen_test

import (
	"testing"

	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	tests := []string{
		fen.Initial,
		"4k3/2pppp2/8/4P1K1/4PP2/3P4/8/8 w - - 0 1",
		"rnbqkbnr/pppppppp/8/8/8/5P2/PPPPP1PP/RNBQKBNR w KQkq - 0 1",
	}

	for _, tt := range tests {
		p, c, np, fm, err := fen.Decode(tt)
		require.NoError(t, err)

		assert.Equal(t, tt, fen.Encode(p, c, np, fm))
	}

}

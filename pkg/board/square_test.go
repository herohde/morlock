package board_test

import (
	"testing"

	"github.com/herohde/morlock/pkg/board"
	"github.com/stretchr/testify/assert"
)

func TestRank(t *testing.T) {
	assert.True(t, board.Rank1.IsValid())
	assert.True(t, board.Rank3.IsValid())
	assert.True(t, board.Rank8.IsValid())
	assert.False(t, board.Rank(8).IsValid())

	assert.Equal(t, board.Rank1.String(), "1")
	assert.Equal(t, board.Rank7.String(), "7")
	assert.Equal(t, board.Rank(4).String(), "5")
}

func TestFile(t *testing.T) {
	assert.True(t, board.FileA.IsValid())
	assert.True(t, board.FileB.IsValid())
	assert.True(t, board.FileH.IsValid())
	assert.False(t, board.File(8).IsValid())

	assert.Equal(t, board.FileA.String(), "A")
	assert.Equal(t, board.FileG.String(), "G")
	assert.Equal(t, board.File(3).String(), "E")
}

func TestSquare(t *testing.T) {
	assert.Equal(t, board.C2, board.NewSquare(board.FileC, board.Rank2))
	assert.Equal(t, board.G5, board.NewSquare(board.FileG, board.Rank5))

	assert.True(t, board.H1.IsValid())
	assert.True(t, board.D4.IsValid())
	assert.True(t, board.A8.IsValid())
	assert.False(t, board.Square(64).IsValid())

	assert.Equal(t, board.H1.String(), "H1")
	assert.Equal(t, board.A1.String(), "A1")
	assert.Equal(t, board.Square(3).String(), "E1")
}

package bernstein_test

import (
	"github.com/herohde/morlock/cmd/bernstein/bernstein"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEval(t *testing.T) {
	tests := []struct {
		pos string

		mobility int
		control  int
		defence  int
		material int
	}{
		{fen.Initial, 20, 22, 5, 39},
		{"k7/8/8/8/8/8/8/K7 w - - 0 1", 3, 3, 3, 0},                                 // K
		{"k7/7R/8/8/8/8/8/K7 w - - 0 1", 3 + 7 + 7, 3 + 7 + 5 /* -2 for k */, 3, 5}, // K+R
		{"k7/7R/8/8/8/8/8/K7 b - - 0 1", 1 /* -2 due to R */, 1, 1, 0},              // K with limited mobility
		{"k7/p6R/8/8/8/8/8/K7 b - - 0 1", 1 + 2, 1 + 1, 1, 1},                       // K+P
		{"k7/1p5R/8/8/8/8/8/K7 b - - 0 1", 2 + 2, 2 + 2, 2, 1},                      // K+P w/ block
		{"k7/7R/1R6/8/8/8/8/K7 b - - 0 1", 0, 0, 0, 0},                              // stalemate
	}

	for _, tt := range tests {
		pos, side, _, _, _ := fen.Decode(tt.pos)

		mobility := bernstein.Mobility(pos, side)
		assert.Equal(t, tt.mobility, mobility, "mobility: %v", pos)
		control := bernstein.Control(pos, side)
		assert.Equal(t, tt.control, control, "control: %v", pos)
		defense := bernstein.KingDefense(pos, side)
		assert.Equal(t, tt.defence, defense, "defence: %v", pos)
		material := bernstein.Material(pos, side)
		assert.Equal(t, tt.material, material, "material: %v", pos)
	}
}

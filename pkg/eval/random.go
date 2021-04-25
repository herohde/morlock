package eval

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"math/rand"
)

// Randomize adds a small amount of randomness to evaluations. The limit specifies how many
// millipawns to add/remove.
func Randomize(eval Evaluator, limit int, seed int64) Evaluator {
	return &random{
		limit: limit,
		rand:  rand.New(rand.NewSource(seed)),
		eval:  eval,
	}
}

type random struct {
	limit int
	rand  *rand.Rand
	eval  Evaluator
}

func (r *random) Evaluate(ctx context.Context, b *board.Board) Pawns {
	return r.eval.Evaluate(ctx, b) + Pawns(r.rand.Intn(r.limit))/1000
}

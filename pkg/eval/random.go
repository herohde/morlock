package eval

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"math/rand"
)

// Randomize adds a small amount of randomness to evaluations. The limit specifies how many
// millipawns to add/remove in the range [-limit/2; limit/2]. Does nothing if limit is zero.
func Randomize(eval Evaluator, limit int, seed int64) Evaluator {
	if limit <= 0 {
		return eval
	}

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
	noise := Pawns(r.rand.Intn(r.limit)-r.limit/2) / 1000
	return r.eval.Evaluate(ctx, b) + noise
}

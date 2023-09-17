package eval

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"math/rand"
)

// Random is a randomized noise generator. It is used to a small amount of randomness to evaluations. The
// limit specifies how many millipawns to add/remove in the range [-limit/2; limit/2]. The default value
// always returns zero.
type Random struct {
	rand  *rand.Rand
	limit int
}

func NewRandom(limit int, seed int64) Random {
	return Random{
		limit: limit,
		rand:  rand.New(rand.NewSource(seed)),
	}
}

func (n Random) Evaluate(ctx context.Context, b *board.Board) Pawns {
	if n.limit <= 0 {
		return 0
	}
	return Pawns(n.rand.Intn(n.limit)-n.limit/2) / 1000
}

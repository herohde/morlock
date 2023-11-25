package search

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// Exploration defines move selection and priority in a given position. Limited exploration is required
// by quiescence search and can be used for forward pruning in full search. Default: explore all
// moves in MVVLVA order.
type Exploration func(ctx context.Context, b *board.Board) (board.MovePriorityFn, board.MovePredicateFn)

func FullExploration(ctx context.Context, b *board.Board) (board.MovePriorityFn, board.MovePredicateFn) {
	return MVVLVA, IsAnyMove
}

// Selection returns a move order and priority for exploring the given moves.
func Selection(list []board.Move) (board.MovePriorityFn, board.MovePredicateFn) {
	rank := map[board.Move]board.MovePriority{}
	for i, m := range list {
		rank[m] = board.MovePriority(len(list) - i)
	}

	priority := func(move board.Move) board.MovePriority {
		return rank[move]
	}
	pick := func(move board.Move) bool {
		_, ok := rank[move]
		return ok
	}
	return priority, pick
}

// MVVLVA implements the MVV-LVA move priority.
func MVVLVA(m board.Move) board.MovePriority {
	if p := board.MovePriority(100 * eval.NominalValueGain(m)); p > 0 {
		return p - board.MovePriority(eval.NominalValue(m.Piece))
	}
	return 0
}

// IsAnyMove selects all moves.
func IsAnyMove(m board.Move) bool {
	return true
}

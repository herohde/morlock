package search

import (
	"container/heap"
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// Priority represents the move order priority.
type Priority int16

// MoveList is move priority queue for move ordering.
type MoveList struct {
	h moveHeap
}

// NewMoveList returns a new move list with the given priorities.
func NewMoveList(moves []board.Move, fn func(move board.Move) Priority) *MoveList {
	h := moveHeap(make([]elm, len(moves)))
	for i, m := range moves {
		h[i] = elm{m: m, val: fn(m)}
	}
	heap.Init(&h)
	return &MoveList{h: h}
}

// Next returns the next move. It is the highest priority move in the list.
func (ml *MoveList) Next() (board.Move, bool) {
	if ml.Size() == 0 {
		return board.Move{}, false
	}
	ret := heap.Pop(&ml.h).(elm)
	return ret.m, true
}

func (ml *MoveList) Size() int {
	return ml.h.Len()
}

func (ml *MoveList) String() string {
	if ml.Size() == 0 {
		return "[size=0]"
	}
	return fmt.Sprintf("[top=%v, size=%v]", ml.h[0].m, ml.Size())
}

type elm struct {
	m   board.Move
	val Priority
}

type moveHeap []elm

func (h moveHeap) Len() int {
	return len(h)
}

func (h moveHeap) Less(i, j int) bool {
	return h[i].val > h[j].val
}

func (h moveHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *moveHeap) Push(x interface{}) {
	panic("fixed size heap")
}

func (h *moveHeap) Pop() interface{} {
	n := len(*h)
	ret := (*h)[n-1]
	*h = (*h)[0 : n-1]
	return ret
}

// MVVLVA returns the MVV-LVA priority.
func MVVLVA(m board.Move) Priority {
	if p := Priority(100 * eval.NominalValueGain(m)); p > 0 {
		return p - Priority(eval.NominalValue(m.Piece))
	}
	return 0
}

// First puts the given move first. Otherwise uses MVVLVA.
type First board.Move

func (f First) MVVLVA(m board.Move) Priority {
	if m.Equals(board.Move(f)) {
		return 1000
	}
	return MVVLVA(m)
}

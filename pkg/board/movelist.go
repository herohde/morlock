package board

import (
	"container/heap"
	"fmt"
	"math"
	"sort"
)

// MovePriority represents the move order priority.
type MovePriority int16

// MovePriorityFn assigns a priority to moves
type MovePriorityFn func(move Move) MovePriority

// First puts the given move first. Otherwise uses the given function.
func First(first Move, fn MovePriorityFn) MovePriorityFn {
	return func(m Move) MovePriority {
		if first.Equals(m) {
			return math.MaxInt16
		}
		return fn(m)
	}
}

// SortByPriority sorts the moves by priority, preserving order for same priority.
func SortByPriority(moves []Move, fn MovePriorityFn) {
	sort.SliceStable(moves, func(i, j int) bool {
		return fn(moves[i]) > fn(moves[j])
	})
}

// MoveList is move priority queue for move ordering.
type MoveList struct {
	h moveHeap
}

// NewMoveList returns a new move list with the given priorities.
func NewMoveList(moves []Move, fn MovePriorityFn) *MoveList {
	h := moveHeap(make([]elm, len(moves)))
	for i, m := range moves {
		h[i] = elm{m: m, val: fn(m)}
	}
	heap.Init(&h)
	return &MoveList{h: h}
}

// Next returns the next move. It is the highest priority move in the list.
func (ml *MoveList) Next() (Move, bool) {
	if ml.Size() == 0 {
		return Move{}, false
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
	m   Move
	val MovePriority
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

package search

import (
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
)

// Bound represents the bound of a -- possibly inexact -- search score.
type Bound uint8

const (
	ExactBound Bound = 0
	UpperBound
	LowerBound
)

// node represents 32bytes.
type node struct {
	hash  board.ZobristHash
	score eval.Score // 8
	move  board.Move // 8
	ply   uint16     // 2
	depth uint8      //  1
	bound Bound      // 1
}

type TranspositionTable struct {
	table []node
	mask  uint64
}

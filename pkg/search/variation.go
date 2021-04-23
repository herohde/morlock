package search

import (
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"time"
)

// PV represents the principal variation for some search depth.
type PV struct {
	Depth int           // depth of search
	Moves []board.Move  // principal variation
	Score eval.Score    // evaluation at depth
	Nodes uint64        // interior/leaf nodes searched
	Time  time.Duration // time taken by search
	Hash  float64       // hash table used [0;1]
}

func (p PV) String() string {
	pv := board.PrintMoves(p.Moves)
	return fmt.Sprintf("depth=%v score=%v nodes=%v time=%v hash=%v%% pv=%v", p.Depth, p.Score, p.Nodes, p.Time, int(100*p.Hash), pv)
}

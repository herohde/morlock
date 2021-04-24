package search

import (
	"context"
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/seekerror/logw"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

// TODO(herohde) 4/17/2021: consider shared linked list for principal variation.

// Bound represents the bound of a -- possibly inexact -- search score.
type Bound uint8

const (
	ExactBound Bound = iota
	LowerBound
)

func (b Bound) String() string {
	switch b {
	case ExactBound:
		return "Exact"
	case LowerBound:
		return "Lower"
	default:
		return "?"
	}
}

// TranspositionTable represents a transposition table to speed up search performance.
// Caveat: evaluation heuristics that depend on the game history (notably, hasCastled or
// last move) may be unsuitable for position-keyed caching. If the recent history is short,
// then the table may only be used for depth greater than some limit. Must be thread-safe.
type TranspositionTable interface {
	// Read returns the bound, depth, score and best move for the given position hash, if present.
	Read(hash board.ZobristHash) (Bound, int, eval.Score, board.Move, bool)
	// Write stores the entry into the table, depending on table semantics and replacement policy.
	Write(hash board.ZobristHash, bound Bound, ply, depth int, score eval.Score, move board.Move) bool

	// Size returns the size of the table in bytes.
	Size() uint64
	// Used returns the utilization as a fraction [0;1].
	Used() float64
}

type TranspositionTableFactory func(ctx context.Context, size uint64) TranspositionTable

// metadata captures node metadata, notably precision and best move. 64bits.
type metadata struct {
	bound      Bound        // 1
	from, to   board.Square // bestmove -- to, from
	promotion  board.Piece  // bestmove -- promotion
	ply, depth uint16       //  4
}

// node represents a search result. 24bytes.
type node struct {
	hash  board.ZobristHash // full hash
	score eval.Score
	md    metadata
}

// table is a transposition table. It uses 32bytes/entry.
type table struct {
	table []*node
	mask  uint64
	used  uint64
}

func NewTranspositionTable(ctx context.Context, size uint64) TranspositionTable {
	n := uint64(1 << (63 - 5 - bits.LeadingZeros64(size)))

	logw.Infof(ctx, "Allocating %vMB TT with %v entries", size>>20, n)

	return &table{
		table: make([]*node, n),
		mask:  n - 1,
	}
}

func (t *table) Size() uint64 {
	return uint64(len(t.table)) << 5
}

func (t *table) Used() float64 {
	// https://github.com/census-instrumentation/opencensus-go/issues/587
	// used := atomic.LoadUint64(&t.used)
	used := t.used
	return float64(used) / float64(len(t.table))
}

func (t *table) Read(hash board.ZobristHash) (Bound, int, eval.Score, board.Move, bool) {
	key := uint64(hash) & t.mask
	addr := (*unsafe.Pointer)(unsafe.Pointer(&t.table[key]))

	ptr := (*node)(atomic.LoadPointer(addr))
	if ptr != nil && hash == ptr.hash {
		bestmove := board.Move{From: ptr.md.from, To: ptr.md.to, Promotion: ptr.md.promotion}
		return ptr.md.bound, int(ptr.md.depth), ptr.score, bestmove, true
	}
	return 0, 0, eval.Score{}, board.Move{}, false
}

func (t *table) Write(hash board.ZobristHash, bound Bound, ply, depth int, score eval.Score, move board.Move) bool {
	key := uint64(hash) & t.mask
	addr := (*unsafe.Pointer)(unsafe.Pointer(&t.table[key]))

	fresh := &node{
		hash:  hash,
		score: score,
		md: metadata{
			bound:     bound,
			from:      move.From,
			to:        move.To,
			promotion: move.Promotion,
			ply:       uint16(ply),
			depth:     uint16(depth),
		},
	}

	ptr := (*node)(atomic.LoadPointer(addr))
	for {
		if val(ptr) > val(fresh) {
			return false // skip: higher value existing node
		}
		if atomic.CompareAndSwapPointer(addr, unsafe.Pointer(ptr), unsafe.Pointer(fresh)) {
			if ptr == nil {
				// https://github.com/census-instrumentation/opencensus-go/issues/587
				// atomic.AddUint64(&t.used, 1)
				t.used++
			}
			return true // ok: overwrite value
		}
		ptr = (*node)(atomic.LoadPointer(addr))
	}
}

func (t *table) String() string {
	return fmt.Sprintf("TT[%v @ %v%%]", t.Size(), int(100*t.Used()))
}

// val defines node value towards replacement logic.
func val(n *node) uint16 {
	if n == nil {
		return 0
	}
	return n.md.ply + (n.md.depth << 1)
}

// WriteFilter is a predicate on the Write operation.
type WriteFilter func(hash board.ZobristHash, bound Bound, ply, depth int, score eval.Score, move board.Move) bool

// WriteLimited is a TranspositionTable wrapper that ignores certain writes, such as
// less than a given minimum depth. Useful if evaluation uses recent move history.
type WriteLimited struct {
	Filter WriteFilter
	TT     TranspositionTable
}

func (w WriteLimited) Read(hash board.ZobristHash) (Bound, int, eval.Score, board.Move, bool) {
	return w.TT.Read(hash)
}

func (w WriteLimited) Write(hash board.ZobristHash, bound Bound, ply, depth int, score eval.Score, move board.Move) bool {
	if w.Filter(hash, bound, ply, depth, score, move) {
		return false
	}
	return w.TT.Write(hash, bound, ply, depth, score, move)
}

func (w WriteLimited) Size() uint64 {
	return w.TT.Size()
}

func (w WriteLimited) Used() float64 {
	return w.TT.Used()
}

// NewMinDepthTranspositionTable creates depth-limited TranspositionTables.
func NewMinDepthTranspositionTable(min int) TranspositionTableFactory {
	return func(ctx context.Context, size uint64) TranspositionTable {
		return WriteLimited{
			Filter: func(hash board.ZobristHash, bound Bound, ply, depth int, score eval.Score, move board.Move) bool {
				return depth < min
			},
			TT: NewTranspositionTable(ctx, size),
		}
	}
}

// NoTranspositionTable is a Nop implementation.
type NoTranspositionTable struct{}

func (n NoTranspositionTable) Read(hash board.ZobristHash) (Bound, int, eval.Score, board.Move, bool) {
	return 0, 0, eval.Score{}, board.Move{}, false
}

func (n NoTranspositionTable) Write(hash board.ZobristHash, bound Bound, ply, depth int, score eval.Score, move board.Move) bool {
	return false
}

func (n NoTranspositionTable) Size() uint64 {
	return 0
}

func (n NoTranspositionTable) Used() float64 {
	return 0
}

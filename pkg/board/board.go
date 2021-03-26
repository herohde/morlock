// Package board contain chess board representation and utilities.
package board

import "fmt"

const (
	repetition3Limit   = 3
	repetition5Limit   = 5
	noprogressPlyLimit = 100
)

type node struct {
	pos        *Position
	hash       ZobristHash
	noprogress int

	next Move // if not current
	prev *node
}

// Board represents a chess board, metadata and history of positions to correctly handle game
// results, notably various draw conditions. Not thread-safe.
type Board struct {
	zt          *ZobristTable
	repetitions map[ZobristHash]int

	fullmoves int
	turn      Color
	result    Result
	current   *node
}

func NewBoard(zt *ZobristTable, pos *Position, turn Color, noprogress, fullmoves int) *Board {
	current := &node{
		pos:        pos,
		noprogress: noprogress,
		hash:       zt.Hash(pos, turn),
	}

	repetitions := map[ZobristHash]int{
		current.hash: 1,
	}

	return &Board{
		zt:          zt,
		repetitions: repetitions,
		fullmoves:   fullmoves,
		turn:        turn,
		current:     current,
	}
}

// Fork branches off a new board, sharing the node history for past positions. If forked, the shared
// history should not be mutated (via PopMove) as the forward moves in node might then become stale.
func (b *Board) Fork() *Board {
	fork := &Board{
		zt:          b.zt,
		repetitions: map[ZobristHash]int{},
		fullmoves:   b.fullmoves,
		turn:        b.turn,
		result:      b.result,
		current: &node{
			pos:        b.current.pos,
			hash:       b.current.hash,
			noprogress: b.current.noprogress,
			prev:       b.current.prev,
		},
	}
	for k, v := range b.repetitions {
		fork.repetitions[k] = v
	}

	return fork
}

func (b *Board) Position() *Position {
	return b.current.pos
}

func (b *Board) Turn() Color {
	return b.turn
}

func (b *Board) NoProgress() int {
	return b.current.noprogress
}

func (b *Board) FullMoves() int {
	return b.fullmoves
}

func (b *Board) Result() Result {
	return b.result
}

// PushMove attempts to make a pseudo-legal move. Returns true iff legal.
func (b *Board) PushMove(m Move) bool {
	if b.result.Reason == Checkmate || b.result.Reason == Stalemate {
		return false // there are no legal moves
	} // else: ignore draws that are not always called correctly.

	next, ok := b.current.pos.Move(m)
	if !ok {
		return false
	}

	// (1) Move is legal. Create new node.

	n := &node{
		pos:        next,
		hash:       b.zt.Move(b.current.hash, b.current.pos, m),
		noprogress: updateNoProgress(b.current.noprogress, m),
		prev:       b.current,
	}

	b.current.next = m
	b.current = n

	// (2) Update board-level metadata.

	b.turn = b.turn.Opponent()
	b.repetitions[b.current.hash]++
	if b.turn == White {
		b.fullmoves++
	}

	// (3) Determine if draw condition applies.

	if b.repetitions[b.current.hash] >= repetition3Limit {
		actual := b.identicalPositionCount(b.current, b.turn, b.current.noprogress)
		switch {
		case actual >= repetition5Limit:
			b.result.Outcome = Draw
			b.result.Reason = Repetition5
		case actual >= repetition3Limit:
			b.result.Outcome = Draw
			b.result.Reason = Repetition3
		default:
			// zobrist collision: not an actual repetition
		}
	}

	if b.current.noprogress >= noprogressPlyLimit {
		b.result.Outcome = Draw
		b.result.Reason = NoProgress
	}

	if m.Type == Capture || ((m.Type == CapturePromotion || m.Type == Promotion) && (m.Promotion == Bishop || m.Promotion == Knight)) {
		if b.current.pos.HasInsufficientMaterial() {
			b.result.Outcome = Draw
			b.result.Reason = InsufficientMaterial
		}
	}

	return true
}

func (b *Board) PopMove() (Move, bool) {
	if b.current.prev == nil {
		return Move{}, false
	}

	// (1) Update board-level metadata.

	b.turn = b.turn.Opponent()
	b.repetitions[b.current.hash]--
	b.result = Result{Outcome: Undecided} // a legal move was made, so not terminal
	if b.turn == Black {
		b.fullmoves--
	}

	// (2) Pop current node.

	b.current = b.current.prev
	m := b.current.next
	b.current.next = Move{}
	return m, true
}

// AdjudicateNoLegalMoves adjudicates the position assuming no legal moves exist.
// The result is then either Mate or Stalemate.
func (b *Board) AdjudicateNoLegalMoves() Result {
	result := Result{Outcome: Draw, Reason: Stalemate}
	if b.Position().IsChecked(b.Turn()) {
		result = Result{Outcome: Loss(b.Turn()), Reason: Checkmate}
	}
	b.Adjudicate(result)
	return result
}

// Adjudicate the position as given.
func (b *Board) Adjudicate(result Result) {
	b.result = result
}

func (b *Board) identicalPositionCount(n *node, turn Color, limit int) int {
	ret := 1
	tmp := n.prev
	t := b.turn.Opponent()

	for i := 1; i < limit && tmp != nil; i++ {
		if tmp.hash == n.hash && turn == t && *tmp.pos == *n.pos {
			ret++
		}
		tmp = tmp.prev
		t = t.Opponent()
	}
	return ret
}

// LastMove returns the last move, if any.
func (b *Board) LastMove() (Move, bool) {
	if b.current.prev != nil {
		return b.current.prev.next, true
	}
	return Move{}, false
}

// HasCasted returns true iff the color has castled.
func (b *Board) HasCastled(c Color) bool {
	t := b.turn.Opponent()
	cur := b.current.prev

	for cur != nil {
		if t == c && cur.next.Type == QueenSideCastle || cur.next.Type == KingSideCastle {
			return true
		}
		t = t.Opponent()
		cur = cur.prev
	}
	return false
}

func (b *Board) String() string {
	return fmt.Sprintf("board{pos=%v, turn=%v, hash=%x (%v) noprogress=%v, fullmoves=%v, result=%v}", b.current.pos, b.turn, b.current.hash, b.repetitions[b.current.hash], b.current.noprogress, b.fullmoves, b.result)
}

func updateNoProgress(old int, m Move) int {
	if m.Type != Normal {
		return 0
	}
	return old + 1
}

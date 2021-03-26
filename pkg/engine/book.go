package engine

import (
	"context"
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"sort"
	"strings"
)

// Book represents an opening book.
type Book interface {
	// Find returns a list -- potentially empty -- of moves given a position. Once an empty
	// list is returned, the book should not be consulted again for the game.
	Find(ctx context.Context, fen string) ([]board.Move, error)
}

// Line represents an opening line: e2e4 d7d5.
type Line []string

func (l Line) String() string {
	return strings.Join(l, " ")
}

// NoBook is an empty opening book.
var NoBook = &book{moves: map[string][]board.Move{}}

// NewBook creates an opening book from a set of opening lines.
func NewBook(lines []Line) (Book, error) {
	m := map[string]map[board.Move]bool{}
	for _, line := range lines {
		key := fen.Initial
		for _, str := range line {
			next, err := board.ParseMove(str)
			if err != nil {
				return nil, fmt.Errorf("invalid line '%v': %v", line, err)
			}

			found := false
			pos, turn, _, _, _ := fen.Decode(key)
			for _, candidate := range pos.PseudoLegalMoves(turn) {
				if !candidate.Equals(next) {
					continue
				}

				found = true
				p, ok := pos.Move(candidate)
				if !ok {
					return nil, fmt.Errorf("invalid line '%v': move %v not legal", line, next)
				}

				if m[fenKey(key)] == nil {
					m[fenKey(key)] = map[board.Move]bool{}
				}
				m[fenKey(key)][candidate] = true

				key = fen.Encode(p, turn.Opponent(), 0, 1)
				break
			}

			if !found {
				return nil, fmt.Errorf("invalid line '%v': move %v not found", line, next)
			}
		}
	}

	dedup := map[string][]board.Move{}
	for k, v := range m {
		var list []board.Move
		for move, _ := range v {
			list = append(list, move)
		}
		sort.Sort(board.ByMVVLVA(list))
		dedup[k] = list
	}
	return &book{moves: dedup}, nil
}

type book struct {
	moves map[string][]board.Move // cropped fen -> []move
}

func (b *book) Find(ctx context.Context, fen string) ([]board.Move, error) {
	return b.moves[fenKey(fen)], nil
}

func fenKey(pos string) string {
	parts := strings.Split(pos, " ")
	return strings.Join(parts[:4], " ")
}

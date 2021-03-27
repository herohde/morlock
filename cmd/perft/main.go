// perft is a movegen debugging tools. See: https://www.chessprogramming.org/Perft_Results.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/seekerror/logw"
	"time"
)

var (
	depth    = flag.Int("depth", 4, "Search depth")
	position = flag.String("fen", "", "Start position (default to standard)")
	divide   = flag.Bool("divide", false, "Divide counts by initial move")
)

func main() {
	ctx := context.Background()
	flag.Parse()

	if *position == "" {
		*position = fen.Initial
	}

	pos, turn, _, _, err := fen.Decode(*position)
	if err != nil {
		logw.Exitf(ctx, "Invalid fen '%v': %v", *position, err)
	}

	for i := 1; i <= *depth; i++ {
		start := time.Now()
		nodes := search(pos, turn, i, *divide && i == *depth)
		duration := time.Since(start)

		println(fmt.Sprintf("perft,%v,%v,%v,%v", *position, i, nodes, duration.Microseconds()))
	}
}

func search(pos *board.Position, turn board.Color, depth int, d bool) int64 {
	if depth == 0 {
		return 1
	}

	var nodes int64
	for _, m := range pos.PseudoLegalMoves(turn) {
		if next, ok := pos.Move(m); ok {
			count := search(next, turn.Opponent(), depth-1, false)
			if d {
				println(fmt.Sprintf("%v: %v", m, count))
			}
			nodes += count
		}
	}
	return nodes
}

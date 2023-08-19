package console

import (
	"context"
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/engine"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
	"github.com/seekerror/logw"
	"go.uber.org/atomic"
	"sort"
	"strconv"
	"strings"
)

const ProtocolName = "console"

// Option is a driver option.
type Option func(*options)

type options struct {
	useHash bool
	hash    int
}

// UseHash instructs the driver to expose and use a Hash setting with the given default
// size in MB. If not exposed, the engine will not use a transposition table.
func UseHash(size int) Option {
	return func(opt *options) {
		opt.useHash = true
		opt.hash = size
	}
}

// Driver implements a console driver for debugging.
type Driver struct {
	e     *engine.Engine
	opt   options
	depth int

	out chan<- string

	root   search.Search
	active atomic.Bool // user is waiting for engine to move

	quit   chan struct{}
	closed atomic.Bool
}

func NewDriver(ctx context.Context, e *engine.Engine, root search.Search, in <-chan string, opts ...Option) (*Driver, <-chan string) {
	var opt options
	for _, fn := range opts {
		fn(&opt)
	}

	out := make(chan string, 100)
	d := &Driver{
		e:    e,
		opt:  opt,
		root: root,
		out:  out,
		quit: make(chan struct{}),
	}
	go d.process(ctx, in)

	return d, out
}

func (d *Driver) Close() {
	if d.closed.CAS(false, true) {
		close(d.quit)
	}
}

func (d *Driver) Closed() <-chan struct{} {
	return d.quit
}

func (d *Driver) process(ctx context.Context, in <-chan string) {
	defer d.Close()
	defer close(d.out)

	logw.Infof(ctx, "Console protocol initialized")

	d.out <- fmt.Sprintf("engine %v (%v)", d.e.Name(), d.e.Author())
	d.printBoard(ctx)

	for {
		select {
		case line, ok := <-in:
			if !ok {
				logw.Infof(ctx, "Input stream broken. Exiting")
				return
			}

			parts := strings.Split(strings.TrimSpace(line), " ")
			if len(parts) == 0 {
				break
			}

			cmd := parts[0]
			args := parts[1:]

			switch strings.ToLower(cmd) {
			case "reset", "r":
				// reset [<fenstring>] moves ...

				d.ensureInactive(ctx)

				pos := fen.Initial
				if len(args) > 0 && args[0] != "moves" {
					pos = strings.Join(args[0:6], " ")
				}
				if err := d.e.Reset(ctx, pos, uint64(d.opt.hash)<<20); err != nil {
					logw.Errorf(ctx, "Invalid position: %v", line)
					return
				}
				move := false
				for _, arg := range args {
					if arg == "moves" {
						move = true
						continue
					}
					if !move {
						continue
					}

					if err := d.e.Move(ctx, arg); err != nil {
						logw.Errorf(ctx, "Invalid position move '%v': %v: %v", arg, line, err)
						return
					}
				}
				d.printBoard(ctx)

			case "undo", "u":
				d.ensureInactive(ctx)

				_ = d.e.TakeBack(ctx)
				d.printBoard(ctx)

			case "print", "p":
				d.printBoard(ctx)

			case "analyze", "a":
				d.ensureInactive(ctx)

				var opt search.Options
				if d.depth > 0 {
					opt.DepthLimit = &d.depth
				}
				if len(args) > 0 {
					depth, _ := strconv.Atoi(args[0])
					opt.DepthLimit = &depth
				}

				out, err := d.e.Analyze(ctx, opt)
				if err != nil {
					logw.Errorf(ctx, "Analyze failed: %v", err)
					return
				}
				d.active.Store(true)

				go func() {
					var last search.PV
					for pv := range out {
						last = pv
						d.out <- pv.String()
					}
					d.searchCompleted(ctx, last)
				}()

			case "depth", "d":
				if len(args) > 0 {
					d.depth, _ = strconv.Atoi(args[0])
				}

			case "hash":
				if len(args) > 0 {
					d.opt.hash, _ = strconv.Atoi(args[0])
				}

			case "nohash":
				d.opt.hash = 0

			case "halt", "stop":
				pv, err := d.e.Halt(ctx)
				if err != nil {
					d.searchCompleted(ctx, pv)
				}

			case "quit", "exit", "q":
				d.ensureInactive(ctx)
				return

			case "":
				// ignore empty command

			default:
				// Assume move if not a recognized command.

				d.ensureInactive(ctx)
				if err := d.e.Move(ctx, cmd); err != nil {
					d.out <- fmt.Sprintf("invalid move: '%v'", cmd)
				} else {
					d.printBoard(ctx)
				}
			}

		case <-d.quit:
			d.ensureInactive(ctx)

			logw.Infof(ctx, "Driver closed")
			return
		}
	}
}

func (d *Driver) ensureInactive(ctx context.Context) {
	d.active.Store(false)
	_, _ = d.e.Halt(ctx)
}

func (d *Driver) searchCompleted(ctx context.Context, pv search.PV) {
	if d.active.CAS(true, false) {
		// Search complete

		if len(pv.Moves) > 0 {
			d.out <- fmt.Sprintf("bestmove %v", pv.Moves[0])
		}

		// Ponder each move for score breakdown.

		b := d.e.Board()

		var sub []result
		for _, move := range b.Position().LegalMoves(b.Turn()) {
			nodes, score, moves, _ := d.root.Search(ctx, &search.Context{TT: search.NoTranspositionTable{}, Ponder: []board.Move{move}}, b, pv.Depth, make(chan struct{}))
			sub = append(sub, result{m: move, s: score, n: nodes - 1, pv: moves[1:]})
		}
		sort.Sort(byScore(sub))

		d.out <- fmt.Sprintf("Search, depth=%v", pv.Depth)
		for i := 0; i < len(sub); i++ {
			d.out <- fmt.Sprintf(" %2d. %v\t%v\t\t(%v nodes\tpv %v)", i+1, sub[i].m, sub[i].s, sub[i].n, board.PrintMoves(sub[i].pv))
		}
	} // else: stale or duplicate result
}

const (
	files      = "    a   b   c   d   e   f   g   h"
	horizontal = "  ---------------------------------"
	vertical   = " | "
)

func (d *Driver) printBoard(ctx context.Context) {
	b := d.e.Board()
	p := b.Position()

	d.out <- ""
	d.out <- files
	d.out <- horizontal
	var sb strings.Builder
	sb.WriteString("8" + vertical)
	for i := board.ZeroSquare; i < board.NumSquares; i++ {
		if i != 0 && i%8 == 0 {
			d.out <- sb.String()
			d.out <- horizontal

			sb.Reset()
			sb.WriteString((7 - i.Rank()).String())
			sb.WriteString(vertical)
		}

		if color, piece, ok := p.Square(board.NumSquares - i - 1); ok {
			sb.WriteString(printPiece(color, piece))
		} else {
			sb.WriteString(" ")
		}
		sb.WriteString(vertical)
	}
	d.out <- sb.String()
	d.out <- horizontal
	d.out <- files
	d.out <- ""
	d.out <- fmt.Sprintf("fen:    %v", d.e.Position())
	d.out <- fmt.Sprintf("result: %v, ply: %v, hash: 0x%x", b.Result(), b.Ply(), b.Hash())
	d.out <- ""
}

func printPiece(c board.Color, p board.Piece) string {
	if c == board.White {
		return strings.ToUpper(p.String())
	}
	return strings.ToLower(p.String())
}

type result struct {
	m  board.Move
	s  eval.Score
	n  uint64
	pv []board.Move
}

// byScore is a sort order by score.
type byScore []result

func (b byScore) Len() int {
	return len(b)
}

func (b byScore) Less(i, j int) bool {
	return b[j].s.Less(b[i].s)
}

func (b byScore) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

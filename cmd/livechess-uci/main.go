// livechess-uci is an adaptor for using a DGT EBoard via LiveChess as a UCI engine. The adaptor
// allows use of DGT EBoards in chess programs, such as CuteChess, by pretending to be an engine.
package main

import (
	"context"
	"flag"
	"github.com/herohde/livechess-go/pkg/livechess"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/engine"
	"github.com/herohde/morlock/pkg/engine/uci"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
	"github.com/seekerror/logw"
	"github.com/seekerror/stdlib/pkg/util/iox"
	"strings"
	"sync/atomic"
)

// TODO(herohde) 12/16/2023: change engine to interface. Protocol seems brittle with setup otherwise.

var (
	serial = flag.String("serial", "auto", "Board selection by serial number (default: auto)")
	flip   = flag.Bool("flip", false, "Flip board")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	id := livechess.EBoardSerial(*serial)
	if id == "auto" {
		auto, err := livechess.AutoDetect(ctx, livechess.DefaultClient)
		if err != nil {
			logw.Exitf(ctx, "Watch failed to autodetect board: %v", err)
		}
		id = auto
	}

	client, events, err := livechess.NewFeed(ctx, id)
	if err != nil {
		logw.Exitf(ctx, "Feed for %v failed: %v", id, err)
	}
	if *flip {
		if err := client.Flip(ctx, true); err != nil {
			logw.Exitf(ctx, "Flip board %v failed: %v", id, err)
		}
	}
	if err := client.Setup(ctx, fen.Initial); err != nil {
		logw.Exitf(ctx, "Setup board %v failed: %v", id, err)
	}

	s := newAdaptor(ctx, client, events)

	e := engine.New(ctx, "livechess-uci", "herohde", s,
		engine.WithOptions(engine.Options{Depth: 1}))

	in := engine.ReadStdinLines(ctx)
	switch <-in {
	case uci.ProtocolName:
		driver, out := uci.NewDriver(ctx, e, in)
		go engine.WriteStdoutLines(ctx, out)

		<-driver.Closed()

	default:
		flag.Usage()
		logw.Exitf(ctx, "Protocol not supported")
	}
}

type adaptor struct {
	client livechess.FeedClient

	last  atomic.Pointer[livechess.EBoardEventResponse] // last with start and move list
	pulse *iox.Pulse
}

func newAdaptor(ctx context.Context, client livechess.FeedClient, events <-chan livechess.EBoardEventResponse) *adaptor {
	ret := &adaptor{
		client: client,
		pulse:  iox.NewPulse(),
	}
	go ret.process(ctx, events)
	return ret
}

func (a *adaptor) Search(ctx context.Context, sctx *search.Context, b *board.Board, depth int) (uint64, eval.Score, []board.Move, error) {
	// start := fen.Encode(b.Position(), b.Turn(), b.NoProgress(), b.FullMoves())

	// (1) Generate possible next legal options

	candidates := map[string]board.Move{}
	for _, m := range b.Position().LegalMoves(b.Turn()) {
		b.PushMove(m)
		next := strings.Split(fen.Encode(b.Position(), b.Turn(), 0, 0), " ")[0]
		candidates[next] = m
		b.PopMove()
	}

	if len(candidates) == 0 {
		if result := b.AdjudicateNoLegalMoves(); result.Reason == board.Checkmate {
			return 1, eval.NegInfScore, nil, nil
		}
		return 1, eval.ZeroScore, nil, nil
	}

	// (2) Wait for a board match one of them

	for {
		if last := a.last.Load(); last != nil {
			if m, ok := candidates[last.Board]; ok {
				return 1, eval.ZeroScore, []board.Move{m}, nil
			}
		}

		select {
		case <-a.pulse.Chan():
			// ok: try again
		case <-ctx.Done():
			return 0, eval.InvalidScore, nil, search.ErrHalted
		}
	}
}

func (a *adaptor) process(ctx context.Context, events <-chan livechess.EBoardEventResponse) {
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}

			if len(event.San) > 0 {
				a.last.Store(&event)
				a.pulse.Emit()
			}

		case <-ctx.Done():
			return
		}
	}
}

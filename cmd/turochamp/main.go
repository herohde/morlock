// turochamp is an implementation of Turing and Champernowne's 1948 TUROCHAMP chess engine.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herohde/morlock/cmd/turochamp/turochamp"
	"github.com/herohde/morlock/pkg/engine"
	"github.com/herohde/morlock/pkg/engine/console"
	"github.com/herohde/morlock/pkg/engine/uci"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
	"github.com/seekerror/logw"
	"os"
	"time"
)

var (
	ply = flag.Int("ply", 2, "Search depth limit (zero if no limit)")
)

func init() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: turochamp [options]

TUROCHAMP is a re-implementation of Alan Turing and David Champernowne's 1948
chess engine, described in "Digital computers applied to games" (1953). The
re-implementation uses the UCI protocol for use in modern chess programs.
Options:
`)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	ctx := context.Background()

	logw.Infof(ctx, "TUROCHAMP 1948 chess engine (%v ply)", *ply)

	s := search.AlphaBeta{
		Eval: search.Quiescence{
			Pick: turochamp.IsConsiderableMove,
			Eval: eval.Randomize(turochamp.Eval{}, 10, time.Now().UnixNano()),
		},
	}

	e := engine.New(ctx, "TUROCHAMP (1948)", "Alan Turing and David Champernowne", s, engine.WithDepthLimit(*ply), engine.WithTable(search.NewMinDepthTranspositionTable(1)))

	in := engine.ReadStdinLines(ctx)
	switch <-in {
	case uci.ProtocolName:
		// Use UCI protocol.

		driver, out := uci.NewDriver(ctx, e, in, uci.UseHash(128))
		go engine.WriteStdoutLines(ctx, out)

		<-driver.Closed()

	case console.ProtocolName:
		driver, out := console.NewDriver(ctx, e, s, in, console.UseHash(128))
		go engine.WriteStdoutLines(ctx, out)

		<-driver.Closed()

	default:
		flag.Usage()
		logw.Exitf(ctx, "Protocol not supported")
	}
}

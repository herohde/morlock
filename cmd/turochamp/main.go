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
	"github.com/herohde/morlock/pkg/search"
	"github.com/seekerror/logw"
	"os"
)

var (
	ply   = flag.Uint("ply", 2, "Search depth limit (zero if no limit)")
	noise = flag.Uint("noise", 10, "Evaluation noise in \"millipawns\" (zero if deterministic)")
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
			Explore: turochamp.ConsiderableMovesOnly,
			Eval:    search.Leaf{Eval: turochamp.Eval{}},
		},
	}

	e := engine.New(ctx, "TUROCHAMP (1948)", "Alan Turing and David Champernowne", s,
		engine.WithOptions(engine.Options{Depth: *ply, Noise: *noise}),
	)

	in := engine.ReadStdinLines(ctx)
	switch <-in {
	case uci.ProtocolName:
		driver, out := uci.NewDriver(ctx, e, in)
		go engine.WriteStdoutLines(ctx, out)

		<-driver.Closed()

	case console.ProtocolName:
		driver, out := console.NewDriver(ctx, e, s, in)
		go engine.WriteStdoutLines(ctx, out)

		<-driver.Closed()

	default:
		flag.Usage()
		logw.Exitf(ctx, "Protocol not supported")
	}
}

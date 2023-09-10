// sargon is an implementation of Dan and Kathe Spracklen's 1978 SARGON chess engine.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herohde/morlock/cmd/sargon/sargon"
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
	ply   = flag.Uint("ply", 1, "Search depth limit (zero if no limit)")
	noise = flag.Int("noise", 10, "Evaluation noise in \"millipawns\" (zero if deterministic)")
)

func init() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: sargon [options]

SARGON is a re-implementation of Dan and Kathe Spracklen's 1978 SARGON
chess engine, described in the book "Sargon - a computer chess program".
The re-implementation uses the UCI protocol for use in modern chess
programs.
Options:
`)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	ctx := context.Background()

	logw.Infof(ctx, "SARGON 1978 chess engine (%v ply)", *ply)

	points := &sargon.Points{}
	s := sargon.Hook{
		Eval: search.AlphaBeta{
			Pick: search.IsNotUnderPromotion,
			Eval: sargon.OnePlyIfChecked{
				Eval: eval.Randomize(points, *noise, time.Now().UnixNano()),
			},
		},
		Hook: points,
	}

	e := engine.New(ctx, "SARGON (1978)", "Dan and Kathe Spracklen", s,
		engine.WithOptions(engine.Options{Depth: *ply, Hash: 64}),
		engine.WithTable(search.NewMinDepthTranspositionTable(1)),
	)

	in := engine.ReadStdinLines(ctx)
	switch <-in {
	case uci.ProtocolName:
		// Use UCI protocol.

		driver, out := uci.NewDriver(ctx, e, in, uci.UseBook(sargon.NewBook(), time.Now().UnixNano()))
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

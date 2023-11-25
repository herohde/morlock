package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herohde/morlock/pkg/engine"
	"github.com/herohde/morlock/pkg/engine/console"
	"github.com/herohde/morlock/pkg/engine/uci"
	"github.com/herohde/morlock/pkg/eval"
	"github.com/herohde/morlock/pkg/search"
	"github.com/seekerror/logw"
	"os"
)

func init() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: morlock [options]

MORLOCK is a simple UCI chess engine.
Options:
`)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	ctx := context.Background()

	s := search.AlphaBeta{
		Eval: search.Leaf{Eval: eval.Material{}},
	}
	e := engine.New(ctx, "morlock", "herohde", s,
		engine.WithOptions(engine.Options{Hash: 64}),
		engine.WithTable(search.NewMinDepthTranspositionTable(1)))

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

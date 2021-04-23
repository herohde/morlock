package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herohde/morlock/pkg/engine"
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

	in := engine.ReadStdinLines(ctx)
	switch <-in {
	case uci.ProtocolName:
		// Use UCI protocol.

		s := search.AlphaBeta{
			Eval: search.Quiescence{
				Pick: search.IsQuickGain,
				Eval: eval.Material{},
			},
		}
		e := engine.New(ctx, "morlock", "herohde", s)

		driver, out := uci.NewDriver(ctx, e, in, uci.UseHash(128))
		go engine.WriteStdoutLines(ctx, out)

		<-driver.Closed()

	default:
		flag.Usage()
		logw.Exitf(ctx, "Protocol not supported")
	}
}

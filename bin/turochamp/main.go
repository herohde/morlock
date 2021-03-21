// turochamp is an implementation of Turing and Champernowne's 1948 TUROCHAMP chess engine.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herohde/morlock/pkg/engine"
	"github.com/herohde/morlock/pkg/engine/uci"
	"github.com/herohde/morlock/pkg/search"
	"github.com/herohde/morlock/pkg/turochamp"
	"github.com/seekerror/logw"
	"os"
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

	in := engine.ReadStdinLines(ctx)
	switch <-in {
	case uci.ProtocolName:
		// Use UCI protocol.

		s := search.NewIterative(search.AlphaBeta{
			Eval: turochamp.Quiescence{
				Eval: turochamp.Eval{},
			},
		}, *ply)

		e := engine.New(ctx, "TUROCHAMP", "Alan Turing and David Champernowne", s)

		driver, out := uci.NewDriver(ctx, e, in)
		go engine.WriteStdoutLines(ctx, out)

		<-driver.Closed()

	default:
		flag.Usage()
		logw.Exitf(ctx, "Protocol not supported")
	}
}

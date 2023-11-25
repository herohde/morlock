// bernstein is an implementation of Alex Bernstein's 1957 IBM 704 chess program.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herohde/morlock/cmd/bernstein/bernstein"
	"github.com/herohde/morlock/pkg/engine"
	"github.com/herohde/morlock/pkg/engine/console"
	"github.com/herohde/morlock/pkg/engine/uci"
	"github.com/herohde/morlock/pkg/search"
	"github.com/seekerror/logw"
	"os"
)

var (
	ply      = flag.Uint("ply", 4, "Search depth limit (zero if no limit)")
	branch   = flag.Int("branch", 7, "Search branch factor limit (zero if no limit)")
	material = flag.Int("material", 8, "Material evaluation multiplier")
	noise    = flag.Uint("noise", 0, "Evaluation noise in \"millipawns\" (zero if deterministic)")
)

func init() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: bernstein [options]

BERNSTEIN is a re-implementation of Alex Bernstein, Michael de V. Roberts, Timothy Arbuckle
and Martin Belsky's 1957 chess program one IBM 704, described in "Computer v. Chess-Player"
(1958) and other articles. The re-implementation uses the UCI protocol for use in modern
chess programs.
Options:
`)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	ctx := context.Background()

	logw.Infof(ctx, "BERNSTEIN 1957 chess engine (%v ply, %v-branch limit)", *ply, *branch)

	s := search.AlphaBeta{
		Explore: bernstein.PlausibleMoveTable{Limit: *branch}.Explore,
		Eval: search.Leaf{
			Eval: bernstein.Eval{Factor: *material},
		},
	}

	e := engine.New(ctx, "BERNSTEIN (1957)", "Alex Bernstein, Michael de V. Roberts, Timothy Arbuckle and Martin Belsky", s,
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

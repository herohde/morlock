package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/herohde/morlock/pkg/turochamp"
	"os"

	"github.com/herohde/morlock/pkg/engine"
	"github.com/herohde/morlock/pkg/engine/uci"
	"github.com/herohde/morlock/pkg/search"
	"github.com/seekerror/logw"
)

func main() {
	flag.Parse()
	ctx := context.Background()

	in := readStdinLines(ctx)
	switch <-in {
	case uci.ProtocolName:
		// Use UCI protocol.

		s := search.PVS{
			Eval: turochamp.Quiescence{
				Eval: turochamp.Eval{},
			},
		}
		e := engine.New(ctx, search.NewIterative(s, 2))

		driver, out := uci.NewDriver(ctx, e, in)
		go writeStdoutLines(ctx, out)

		<-driver.Closed()
	}

	logw.Exitf(ctx, "Morlock exited")
}

func readStdinLines(ctx context.Context) <-chan string {
	ret := make(chan string, 1)
	go func() {
		defer close(ret)

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			logw.Debugf(ctx, "<< %v", scanner.Text())
			ret <- scanner.Text()
		}
	}()
	return ret
}

func writeStdoutLines(ctx context.Context, out <-chan string) {
	for line := range out {
		logw.Debugf(ctx, ">> %v", line)
		_, _ = fmt.Fprintln(os.Stdout, line)
	}
}

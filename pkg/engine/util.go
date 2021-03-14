package engine

import (
	"bufio"
	"context"
	"fmt"
	"github.com/seekerror/logw"
	"os"
)

// ReadStdinLines reads stdin lines into a chan. Async.
func ReadStdinLines(ctx context.Context) <-chan string {
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

// WriteStdoutLines writes lines from the given chan to stdout.
func WriteStdoutLines(ctx context.Context, out <-chan string) {
	for line := range out {
		logw.Debugf(ctx, ">> %v", line)
		_, _ = fmt.Fprintln(os.Stdout, line)
	}
}

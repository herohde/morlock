package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herohde/morlock/pkg/board"

	"github.com/seekerror/build"
	"github.com/seekerror/logw"
)

var version = build.NewVersion(0, 83, 0)

func main() {
	flag.Parse()
	ctx := context.Background()

	logw.Exitf(ctx, "Morlock %v exited", version)
}

func Foo(m board.Move) string {
	if m.Promotion.IsValid() {
		return fmt.Sprintf("%v%v%v", m.From, m.To, m.Promotion)
	}
	return fmt.Sprintf("%v%v", m.From, m.To)
}

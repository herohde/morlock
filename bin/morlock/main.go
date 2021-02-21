package main

import (
	"context"
	"flag"

	"github.com/seekerror/build"
	"github.com/seekerror/logw"
)

var version = build.NewVersion(0, 83, 0)

func main() {
	flag.Parse()
	ctx := context.Background()

	logw.Exitf(ctx, "Morlock %v exited", version)
}

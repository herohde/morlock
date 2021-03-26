package sargon

import (
	"context"
	"github.com/herohde/morlock/pkg/engine"
	"github.com/seekerror/logw"
)

// Book contains the sargon opening book of playing either e2e4 or d2d4.
var Book engine.Book

func init() {
	var err error
	Book, err = engine.NewBook([]engine.Line{
		{"e2e4", "e7e5"},
		{"d2d4", "d7d5"},
	})
	if err != nil {
		logw.Exitf(context.Background(), "Invalid book: %v", err)
	}
}

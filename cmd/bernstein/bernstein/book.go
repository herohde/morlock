package bernstein

import "github.com/herohde/morlock/pkg/engine"

var opening = engine.Line([]string{"e2e4"})

func NewBook() engine.Book {
	ret, _ := engine.NewBook([]engine.Line{opening})
	return ret
}

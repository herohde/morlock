package board

import "fmt"

// Result represents the result of a game, if any, with reason.
type Result struct {
	Outcome Outcome
	Reason  Reason
}

func (r Result) IsKnown() bool {
	return r.Outcome > Unknown
}

func (r Result) IsTerminal() bool {
	return r.Outcome > Undecided
}

func (r Result) String() string {
	switch {
	case r.IsTerminal():
		return fmt.Sprintf("%v { %v }", r.Outcome, r.Reason)
	case r.IsKnown():
		return r.Outcome.String()
	default:
		return "?"
	}
}

// Outcome represents the result of a game, if any. Result include a special Unknown
// option to better support the lazy movegen approach, where we only discover the
// true result when no legal move exists. 3 bits.
type Outcome uint8

const (
	Unknown Outcome = iota // = is one of the below options, but not known which yet
	Undecided
	WhiteWins
	BlackWins
	Draw
)

// Win returns a win outcome for the color.
func Win(c Color) Outcome {
	if c == White {
		return WhiteWins
	}
	return BlackWins
}

// Loss returns a loss outcome for the color.
func Loss(c Color) Outcome {
	if c == White {
		return BlackWins
	}
	return WhiteWins
}

func (o Outcome) String() string {
	switch o {
	case WhiteWins:
		return "1-0"
	case BlackWins:
		return "0-1"
	case Draw:
		return "1/2-1/2"
	case Undecided:
		return "tbd"
	default:
		return "?"
	}
}

// Reason is the reason for a terminal result.
type Reason string

const (
	Checkmate Reason = "Checkmate"
	Resigned  Reason = "Opponent Resigned"
	TimedOut  Reason = "Opponent lost on time"

	Stalemate            Reason = "Stalemate"
	Repetition3          Reason = "3-Fold Repetition" // can be claimed, but does not have to be
	Repetition5          Reason = "5-Fold Repetition"
	NoProgress           Reason = "No progress"
	InsufficientMaterial Reason = "Insufficient Material"
	Agreement            Reason = "Agreement"
)

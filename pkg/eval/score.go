package eval

import (
	"fmt"
	"github.com/herohde/morlock/pkg/board"
)

// Score is signed move or position score in pawns. Positive favors white. If all pawns become
// queens and the opponent has only the king left, the standard material advantage score
// is: 9*8 (p) + 9 (q) + 2*5 (r) + 2*3 (k) + 2*3 (b) = 103. Score must be +/- 1,000,000, although
// a human interpretation in centi-pawns is desirable.
type Score float32

const (
	NegInf         = MinScore - 1
	MinScore Score = -1000000
	MaxScore Score = 1000000
	Inf            = MaxScore + 1
)

func (s Score) String() string {
	return fmt.Sprintf("%.2f", s)
}

// Unit returns the signed unit for the color: 1 for White and -1 for Black.
func Unit(c board.Color) Score {
	if c == board.White {
		return 1
	} else {
		return -1
	}
}

// Crop crops a Score into [MinScore;MaxScore].
func Crop(s Score) Score {
	switch {
	case s > MaxScore:
		return MaxScore
	case s < MinScore:
		return MinScore
	default:
		return s
	}
}

// Max returns the largest of the given scores.
func Max(a, b Score) Score {
	if a < b {
		return b
	}
	return a
}

// Min returns the smallest of the given scores.
func Min(a, b Score) Score {
	if a < b {
		return a
	}
	return b
}

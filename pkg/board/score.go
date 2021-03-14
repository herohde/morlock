package board

import "fmt"

// Score is signed move or position score in centi-pawns. Positive favors white. If all pawns
// become queens and the opponent has only the king left, the standard material advantage score
// is: 9*8 (p) + 9 (q) + 2*5 (r) + 2*3 (k) + 2*3 (b) = 103. Score must be within +/- 300.00. 16 bits.
type Score int16

const (
	MinScore Score = -30000
	MaxScore Score = 30000
)

func (s Score) String() string {
	return fmt.Sprintf("%.2f", float64(s)/100)
}

// CropScore crops a Score into [MinScore;MaxScore].
func CropScore(s Score) Score {
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

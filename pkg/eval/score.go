package eval

import (
	"fmt"
)

// ScoreType represents the type of score.
type ScoreType int8

const (
	Invalid ScoreType = iota
	Heuristic
	MateInX
	Inf    // Won position (= opponent checkmate)
	NegInf // Lost position (= in checkmate)
)

// Pawns presents a fractional number of pawns.
type Pawns float32

func (p Pawns) String() string {
	return fmt.Sprintf("%.2f", p)
}

// Limit crops the value to +/- limit.
func Limit(pawns, limit Pawns) Pawns {
	switch {
	case pawns < -limit:
		return -limit
	case limit < pawns:
		return limit
	default:
		return pawns
	}
}

// Score is signed position score in "pawns", unless decided or mate-in-X. Positive favors
// the side to move. If all pawns become queens and the opponent has only the king left,
// the standard material advantage score is: 9*8 (p) + 9 (q) + 2*5 (r) + 2*3 (k) + 2*3 (b)
// = 103. The score can be arbitrary, but is reported as centi-pawns to humans.
type Score struct {
	Type  ScoreType
	Mate  int8 // Non-zero ply to forced mate. Negative if being mated.
	Pawns Pawns
}

var (
	InvalidScore = Score{Type: Invalid}
	ZeroScore    = Score{Type: Heuristic}
	InfScore     = Score{Type: Inf}
	NegInfScore  = Score{Type: NegInf}
)

// HeuristicScore returns a Heuristic score with the given evaluation.
func HeuristicScore(pawns Pawns) Score {
	return Score{Type: Heuristic, Pawns: pawns}
}

// MateInXScore returns a MateInX score with the given evaluation.
func MateInXScore(mate int8) Score {
	return Score{Type: MateInX, Mate: mate}
}

func (s Score) IsInvalid() bool {
	return s.Type == Invalid
}

func (s Score) IsHeuristic() bool {
	return s.Type == Heuristic
}

func (s Score) IsMateInX() bool {
	return s.Type == MateInX
}

func (s Score) IsInf() bool {
	return s.Type == Inf
}

func (s Score) IsNegInf() bool {
	return s.Type == NegInf
}

// MateDistance returns the ply distance to CheckMate, if known. If a forced mate is possible,
// there is generally no reason to search deeper.
func (s Score) MateDistance() (int8, bool) {
	switch s.Type {
	case MateInX:
		if s.Mate < 0 {
			return -s.Mate, true
		}
		return s.Mate, true
	case Inf, NegInf:
		return 0, true
	default:
		return 0, false
	}
}

// Negates returns the negative score, as viewed from the opponent.
func (s Score) Negate() Score {
	switch s.Type {
	case Heuristic:
		return HeuristicScore(-s.Pawns)
	case MateInX:
		return MateInXScore(-s.Mate)
	case Inf:
		return NegInfScore
	case NegInf:
		return InfScore
	default:
		return InvalidScore
	}
}

// Less implements < Score ordering. The group ordering is: -inf < negative mate <
// heuristic < positive mate < inf. Mates are ordered by closeness to checkmate
// within each mate group, e.g., M2 < M1 and M-1 < M-2.
func (s Score) Less(o Score) bool {
	if s == o || s.Type == Inf || o.Type == NegInf {
		return false
	}
	if s.Type == NegInf || o.Type == Inf {
		return true
	}

	switch s.Type {
	case Heuristic:
		switch o.Type {
		case Heuristic:
			return s.Pawns < o.Pawns
		case MateInX:
			return o.Mate > 0
		}

	case MateInX:
		switch o.Type {
		case Heuristic:
			return s.Mate < 0
		case MateInX:
			if s.Mate < 0 || o.Mate < 0 {
				return s.Mate < o.Mate
			}
			return s.Mate > o.Mate
		}
	}

	return false
}

func (s Score) String() string {
	switch s.Type {
	case Heuristic:
		return s.Pawns.String()
	case MateInX:
		return fmt.Sprintf("M%v", s.Mate)
	case Inf:
		return "+inf"
	case NegInf:
		return "-inf"
	default:
		return "?"
	}
}

// IncrementMateDistance adds 1 ply to a MateInX or Inf/NegInf. Heuristic scores are unchanged.
func IncrementMateDistance(s Score) Score {
	switch s.Type {
	case Inf:
		return MateInXScore(1)
	case NegInf:
		return MateInXScore(-1)
	case MateInX:
		if s.Mate < 0 {
			return MateInXScore(s.Mate - 1)
		}
		return MateInXScore(s.Mate + 1)
	default:
		return s
	}
}

// Max returns the largest of the given scores.
func Max(a, b Score) Score {
	if a.Less(b) {
		return b
	}
	return a
}

// Min returns the smallest of the given scores.
func Min(a, b Score) Score {
	if a.Less(b) {
		return a
	}
	return b
}

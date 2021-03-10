package board

// Color represents the playing side/color: white or black. 1 bit.
type Color uint8

const (
	White Color = iota
	Black
)

const (
	ZeroColor Color = 0
	NumColors Color = 2
)

func (c Color) Opponent() Color {
	if c == White {
		return Black
	}
	return White
}

// Unit returns the signed unit for the color: 1 for White and -1 for Black.
func (c Color) Unit() Score {
	if c == White {
		return 1
	} else {
		return -1
	}
}

func (c Color) String() string {
	switch c {
	case White:
		return "w"
	case Black:
		return "b"
	default:
		return "?"
	}
}

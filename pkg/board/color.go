package board

// Color represents the playing side/color: white or black. 1 bit.
type Color uint8

const (
	White Color = iota
	Black
)

func ParseColor(r rune) (Color, bool) {
	switch r {
	case 'w', 'W':
		return White, true
	case 'b', 'B':
		return Black, true
	default:
		return 0, false
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

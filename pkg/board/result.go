package board

// Result represents the result of a game, if any. 2 bits.
type Result uint8

const (
	Undecided Result = iota
	WhiteWins
	BlackWins
	Draw
)

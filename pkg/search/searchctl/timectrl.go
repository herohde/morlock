package searchctl

import (
	"context"
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"github.com/seekerror/logw"
	"github.com/seekerror/stdlib/pkg/lang"
	"time"
)

// TimeControl represents time control information.
type TimeControl struct {
	White, Black time.Duration
	Moves        int // 0 == rest of game
}

// Limits returns a soft and hard limit for making move with the given color. The
// interpretation is that after the soft limit, no new search should be conducted.
func (t TimeControl) Limits(c board.Color) (time.Duration, time.Duration) {
	remainder := t.White
	if c == board.Black {
		remainder = t.Black
	}

	// We assume 40 moves to end the game, if nothing else is known.
	// Let B = T/80 be the soft timeout and the hard timeout be 3B.

	moves := time.Duration(40)
	if t.Moves > 0 {
		moves = time.Duration(t.Moves) + 1
	}

	soft := remainder / (2 * moves)
	hard := 3 * soft
	return soft, hard
}

func (t TimeControl) String() string {
	if t.Moves == 0 {
		return fmt.Sprintf("%.1f<>%.1f", t.White.Seconds(), t.Black.Seconds())
	}
	return fmt.Sprintf("%.1f<>%.1f[moves=%v]", t.White.Seconds(), t.Black.Seconds(), t.Moves)
}

// EnforceTimeControl enforces the time control limits, if any. Returns soft limit.
func EnforceTimeControl(ctx context.Context, h Handle, tc lang.Optional[TimeControl], turn board.Color) (time.Duration, bool) {
	c, ok := tc.V()
	if !ok {
		return 0, false
	}

	soft, hard := c.Limits(turn)
	time.AfterFunc(hard, func() {
		h.Halt()
	})

	logw.Debugf(ctx, "Time control limits for %v: [%v; %v]", c, soft, hard)
	return soft, true
}

# morlock

Morlock is a hobbyist chess engine in Go. It supports a few standard techniques and protocols and
is currently mainly used to implement a few historical engines, TUROCHAMP (1948) and SARGON (1978).
These historical engines can be played on [lichess.org](lichess.org). They are available
24/7 at various search depths.

TUROCHAMP (1948) by Alan Turing and David Champernowne. It uses a full search with an unbounded
quiescence search of "considerable moves" using material ratio and position play heuristics:

*  [turochamp-1ply](https://lichess.org/@/turochamp-1ply). Rating ~1400 (blitz/rapid).
*  [turochamp-2ply](https://lichess.org/@/turochamp-2ply). Rating ~1500 (blitz/rapid).

SARGON (1978) by Dan and Kathe Spracklen. It uses a full search with material exchange, king/queen pins
and board control heuristics:

*  [sargon-1ply](https://lichess.org/@/sargon-1ply). Rating ~1300 (blitz/rapid).
*  [sargon-2ply](https://lichess.org/@/sargon-2ply). Rating ~1400 (blitz/rapid).
*  [sargon-3ply](https://lichess.org/@/sargon-3ply). Rating ~1500 (blitz/rapid).
*  [sargon-4ply](https://lichess.org/@/sargon-4ply). Rating ~1600 (blitz/rapid).

These engines each have quirks, blind spots and limitations, which is part of their charm. They play
at low search depths to entertain rather than win.

Enjoy!

_June 2021_

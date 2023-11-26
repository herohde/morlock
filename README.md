# morlock

Morlock is a hobbyist chess engine in Go. It supports a few standard techniques and protocols and
is currently mainly used to re-implement the following historical chess engines:

### TUROCHAMP (1948) by Alan Turing and David Champernowne

Turochamp uses a full search with an unbounded
quiescence search of "considerable moves" using material ratio and position play heuristics:

*  [turochamp-1ply](https://lichess.org/@/turochamp-1ply). Rating ~1300 (blitz/rapid).
*  [turochamp-2ply](https://lichess.org/@/turochamp-2ply). Rating ~1600 (blitz/rapid).

### BERNSTEIN (1957) by Alex Bernstein, Michael de V. Roberts, Timothy Arbuckle and Martin Belsky 

Bernstein uses a selective search of "plausible moves":

*  [bernstein-4ply](https://lichess.org/@/bernstein-4ply)

### SARGON (1978) by Dan and Kathe Spracklen

Sargon uses a full search with material exchange, king/queen pins and board control heuristics:

*  [sargon-1ply](https://lichess.org/@/sargon-1ply). Rating ~1300 (blitz/rapid).
*  [sargon-2ply](https://lichess.org/@/sargon-2ply). Rating ~1400 (blitz/rapid).
*  [sargon-3ply](https://lichess.org/@/sargon-3ply). Rating ~1700 (blitz/rapid).
*  [sargon-4ply](https://lichess.org/@/sargon-4ply). Rating ~1800 (blitz/rapid).

Each engine can be played 24/7 for free on [lichess.org](https://lichess.org). They have quirks, blind spots and limitations,
which is part of their charm -- and play at low search depths to entertain rather than win.

_November 2023_

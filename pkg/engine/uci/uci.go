// Package uci contains a driver for using the engine under the UCI protocol.
//
// See: http://wbec-ridderkerk.nl/html/UCIProtocol.html
// See: https://en.wikipedia.org/wiki/Universal_Chess_Interface
package uci

import (
	"context"
	"fmt"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/board/fen"
	"github.com/herohde/morlock/pkg/engine"
	"github.com/herohde/morlock/pkg/search"
	"github.com/seekerror/logw"
	"go.uber.org/atomic"
	"strings"
	"time"
)

const ProtocolName = "uci"

// Driver implements a UCI driver for an engine. It is activated if sent "uci".
type Driver struct {
	e   *engine.Engine
	out chan<- string

	active atomic.Bool    // user is waiting for engine to move
	ponder chan search.PV // chan for intermediate search information

	quit   chan struct{}
	closed atomic.Bool
}

func NewDriver(ctx context.Context, e *engine.Engine, in <-chan string) (*Driver, <-chan string) {
	out := make(chan string, 100)
	d := &Driver{
		e:      e,
		out:    out,
		ponder: make(chan search.PV, 400),
		quit:   make(chan struct{}),
	}
	go d.process(ctx, in)

	return d, out
}

func (d *Driver) Close() {
	if d.closed.CAS(false, true) {
		close(d.quit)
	}
}

func (d *Driver) Closed() <-chan struct{} {
	return d.quit
}

func (d *Driver) process(ctx context.Context, in <-chan string) {
	defer d.Close()
	defer close(d.out)

	// * uci
	//
	//	tell engine to use the uci (universal chess interface),
	//	this will be send once as a first command after program boot
	//	to tell the engine to switch to uci mode.
	//	After receiving the uci command the engine must identify itself with the "id" command
	//	and sent the "option" commands to tell the GUI which engine settings the engine supports if any.
	//	After that the engine should sent "uciok" to acknowledge the uci mode.
	//	If no uciok is sent within a certain time period, the engine task will be killed by the GUI.

	logw.Infof(ctx, "UCI protocol initialized")

	// * id
	//	* name <x>
	//		this must be sent after receiving the "uci" command to identify the engine,
	//		e.g. "id name Shredder X.Y\n"
	//	* author <x>
	//		this must be sent after receiving the "uci" command to identify the engine,
	//		e.g. "id author Stefan MK\n"

	d.out <- fmt.Sprintf("id name %v", d.e.Name())
	d.out <- fmt.Sprintf("id author herohde")

	// * uciok
	//
	//	Must be sent after the id and optional options to tell the GUI that the engine
	//	has sent all infos and is ready in uci mode.

	d.out <- fmt.Sprintf("uciok")

	for {
		select {
		case line, ok := <-in:
			if !ok {
				logw.Infof(ctx, "Input stream broken. Exiting")
				return
			}

			parts := strings.Split(strings.TrimSpace(line), " ")
			if len(parts) == 0 {
				break
			}

			cmd := parts[0]
			args := parts[1:]

			switch strings.ToLower(cmd) {
			case "isready":
				// * isready
				//
				//  this is used to synchronize the engine with the GUI. When the GUI has sent a command or
				//	multiple commands that can take some time to complete,
				//	this command can be used to wait for the engine to be ready again or
				//	to ping the engine to find out if it is still alive.
				//	E.g. this should be sent after setting the path to the tablebases as this can take some time.
				//	This command is also required once before the engine is asked to do any search
				//	to wait for the engine to finish initializing.
				//	This command must always be answered with "readyok" and can be sent also when the engine is calculating
				//	in which case the engine should also immediately answer with "readyok" without stopping the search.

				// * readyok
				//
				//	This must be sent when the engine has received an "isready" command and has
				//	processed all input and is ready to accept new commands now.
				//	It is usually sent after a command that can take some time to be able to wait for the engine,
				//	but it can be used anytime, even when the engine is searching,
				//	and must always be answered with "isready".

				d.out <- "readyok"

			case "debug":
				// * debug [ on | off ]
				//
				//	switch the debug mode of the engine on and off.
				//	In debug mode the engine should sent additional infos to the GUI, e.g. with the "info string" command,
				//	to help debugging, e.g. the commands that the engine has received etc.
				//	This mode should be switched off by default and this command can be sent
				//	any time, also when the engine is thinking.

			case "setoption":
				// * setoption name <id> [value <x>]
				//
				//	this is sent to the engine when the user wants to change the internal parameters
				//	of the engine. For the "button" type no value is needed.
				//	One string will be sent for each parameter and this will only be sent when the engine is waiting.
				//	The name and value of the option in <id> should not be case sensitive and can inlude spaces.
				//	The substrings "value" and "name" should be avoided in <id> and <x> to allow unambiguous parsing,
				//	for example do not use <name> = "draw value".
				//	Here are some strings for the example below:
				//	   "setoption name Nullmove value true\n"
				//      "setoption name Selectivity value 3\n"
				//	   "setoption name Style value Risky\n"
				//	   "setoption name Clear Hash\n"
				//	   "setoption name NalimovPath value c:\chess\tb\4;c:\chess\tb\5\n"

			case "register":
				// * register
				//
				//	this is the command to try to register an engine or to tell the engine that registration
				//	will be done later. This command should always be sent if the engine	has sent "registration error"
				//	at program startup.
				//	The following tokens are allowed:
				//	* later
				//	   the user doesn't want to register the engine now.
				//	* name <x>
				//	   the engine should be registered with the name <x>
				//	* code <y>
				//	   the engine should be registered with the code <y>
				//	Example:
				//	   "register later"
				//	   "register name Stefan MK code 4359874324"

			case "ucinewgame":
				// * ucinewgame
				//
				//   this is sent to the engine when the next search (started with "position" and "go") will be from
				//   a different game. This can be a new game the engine should play or a new game it should analyse but
				//   also the next position from a testsuite with positions only.
				//   If the GUI hasn't sent a "ucinewgame" before the first "position" command, the engine shouldn't
				//   expect any further ucinewgame commands as the GUI is probably not supporting the ucinewgame command.
				//   So the engine should not rely on this command even though all new GUIs should support it.
				//   As the engine's reaction to "ucinewgame" can take some time the GUI should always send "isready"
				//   after "ucinewgame" to wait for the engine to finish its operation.

				d.ensureInactive(ctx)

			case "position":
				// * position [fen <fenstring> | startpos ]  moves <move1> .... <movei>
				//
				//	set up the position described in fenstring on the internal board and
				//	play the moves on the internal chess board.
				//	if the game was played  from the start position the string "startpos" will be sent
				//	Note: no "new" command is needed. However, if this position is from a different game than
				//	the last position sent to the engine, the GUI should have sent a "ucinewgame" inbetween.

				d.ensureInactive(ctx)

				position := fen.Initial
				if len(args) >= 7 && args[0] == "fen" {
					position = strings.Join(args[1:7], " ")
				}

				// TODO(herohde) 3/9/2021: check if just continuation of game.

				if err := d.e.Reset(ctx, position); err != nil {
					logw.Errorf(ctx, "Invalid position: %v", line)
					return
				}

				move := false
				for _, arg := range args {
					if arg == "moves" {
						move = true
						continue
					}
					if !move {
						continue
					}

					m, err := board.ParseMove(arg)
					if err != nil {
						logw.Errorf(ctx, "Invalid position move '%v': %v: %v", arg, line, err)
						return
					}
					if err := d.e.Move(ctx, m); err != nil {
						logw.Errorf(ctx, "Invalid position move '%v': %v: %v", arg, line, err)
						return
					}
				}

			case "go":
				// * go
				//
				//	start calculating on the current position set up with the "position" command.
				//	There are a number of commands that can follow this command, all will be sent in the same string.
				//	If one command is not sent its value should be interpreted as it would not influence the search.
				//	* searchmoves <move1> .... <movei>
				//		restrict search to this moves only
				//		Example: After "position startpos" and "go infinite searchmoves e2e4 d2d4"
				//		the engine should only search the two moves e2e4 and d2d4 in the initial position.
				//	* ponder
				//		start searching in pondering mode.
				//		Do not exit the search in ponder mode, even if it's mate!
				//		This means that the last move sent in in the position string is the ponder move.
				//		The engine can do what it wants to do, but after a "ponderhit" command
				//		it should execute the suggested move to ponder on. This means that the ponder move sent by
				//		the GUI can be interpreted as a recommendation about which move to ponder. However, if the
				//		engine decides to ponder on a different move, it should not display any mainlines as they are
				//		likely to be misinterpreted by the GUI because the GUI expects the engine to ponder
				//	   on the suggested move.
				//	* wtime <x>
				//		white has x msec left on the clock
				//	* btime <x>
				//		black has x msec left on the clock
				//	* winc <x>
				//		white increment per move in mseconds if x > 0
				//	* binc <x>
				//		black increment per move in mseconds if x > 0
				//	* movestogo <x>
				//      there are x moves to the next time control,
				//		this will only be sent if x > 0,
				//		if you don't get this and get the wtime and btime it's sudden death
				//	* depth <x>
				//		search x plies only.
				//	* nodes <x>
				//	   search x nodes only,
				//	* mate <x>
				//		search for a mate in x moves
				//	* movetime <x>
				//		search exactly x mseconds
				//	* infinite
				//		search until the "stop" command. Do not exit the search without being told so in this mode!

				d.ensureInactive(ctx)

				var opt search.Options

				// TODO(herohde) 3/8/2021: parse options.

				out, err := d.e.Analyze(ctx, opt)
				if err != nil {
					logw.Errorf(ctx, "Analyze failed: %v", err)
					return
				}
				d.active.Store(true)
				go func() {
					var last search.PV
					for pv := range out {
						last = pv
						d.ponder <- pv
					}
					d.searchCompleted(ctx, last)
				}()

				/*
					time.AfterFunc(10*time.Second, func() {
						_, _ = d.e.Halt(ctx)
					})
				*/

			case "stop":
				// * stop
				//
				//	stop calculating as soon as possible,
				//	don't forget the "bestmove" and possibly the "ponder" token when finishing the search

				pv, err := d.e.Halt(ctx)
				if err != nil {
					d.searchCompleted(ctx, pv)
				}

			case "ponderhit":
				// * ponderhit
				//
				//	the user has played the expected move. This will be sent if the engine was told to ponder on the same move
				//	the user has played. The engine should continue searching but switch from pondering to normal search.

			case "quit":
				// * quit
				//
				//	quit the program as soon as possible
				return

			default:
				logw.Warningf(ctx, "Unknown command '%v': %v", cmd, args)
			}

		case pv := <-d.ponder:
			// * info
			//	the engine wants to send infos to the GUI. This should be done whenever one of the info has changed.
			//	The engine can send only selected infos and multiple infos can be send with one info command,
			//	e.g. "info currmove e2e4 currmovenumber 1" or
			//	     "info depth 12 nodes 123456 nps 100000".
			//	Also all infos belonging to the pv should be sent together
			//	e.g. "info depth 2 score cp 214 time 1242 nodes 2124 nps 34928 pv e2e4 e7e5 g1f3"
			//	I suggest to start sending "currmove", "currmovenumber", "currline" and "refutation" only after one second
			//	to avoid too much traffic.
			//	Additional info:
			//	* depth
			//		search depth in plies
			//	* seldepth
			//		selective search depth in plies,
			//		if the engine sends seldepth there must also a "depth" be present in the same string.
			//	* time
			//		the time searched in ms, this should be sent together with the pv.
			//	* nodes
			//		x nodes searched, the engine should send this info regularly
			//	* pv  ...
			//		the best line found
			//	* multipv
			//		this for the multi pv mode.
			//		for the best move/pv add "multipv 1" in the string when you send the pv.
			//		in k-best mode always send all k variants in k strings together.
			//	* score
			//		* cp
			//			the score from the engine's point of view in centipawns.
			//		* mate
			//			mate in y moves, not plies.
			//			If the engine is getting mated use negativ values for y.
			//		* lowerbound
			//	      the score is just a lower bound.
			//		* upperbound
			//		   the score is just an upper bound.
			//	* currmove
			//		currently searching this move
			//	* currmovenumber
			//		currently searching move number x, for the first move x should be 1 not 0.
			//	* hashfull
			//		the hash is x permill full, the engine should send this info regularly
			//	* nps
			//		x nodes per second searched, the engine should send this info regularly
			//	* tbhits
			//		x positions where found in the endgame table bases
			//	* cpuload
			//		the cpu usage of the engine is x permill.
			//	* string
			//		any string str which will be displayed be the engine,
			//		if there is a string command the rest of the line will be interpreted as .
			//	* refutation   ...
			//	   move  is refuted by the line  ... , i can be any number >= 1.
			//	   Example: after move d1h5 is searched, the engine can send
			//	   "info refutation d1h5 g6h5"
			//	   if g6h5 is the best answer after d1h5 or if g6h5 refutes the move d1h5.
			//	   if there is norefutation for d1h5 found, the engine should just send
			//	   "info refutation d1h5"
			//		The engine should only send this if the option "UCI_ShowRefutations" is set to true.
			//	* currline   ...
			//	   this is the current line the engine is calculating.  is the number of the cpu if
			//	   the engine is running on more than one cpu.  = 1,2,3....
			//	   if the engine is just using one cpu,  can be omitted.
			//	   If  is greater than 1, always send all k lines in k strings together.
			//		The engine should only send this if the option "UCI_ShowCurrLine" is set to true.

			if d.active.Load() {
				d.out <- printPV(pv)
			}

		case <-d.quit:
			d.ensureInactive(ctx)

			logw.Infof(ctx, "Driver closed")
			return
		}
	}
}

func (d *Driver) ensureInactive(ctx context.Context) {
	d.active.Store(false)
	_, _ = d.e.Halt(ctx)
}

func (d *Driver) searchCompleted(ctx context.Context, pv search.PV) {
	if d.active.CAS(true, false) {
		if len(pv.Moves) > 0 {
			// * bestmove <move1> [ ponder <move2> ]
			//
			//	the engine has stopped searching and found the move <move> best in this position.
			//	the engine can send the move it likes to ponder on. The engine must not start pondering automatically.
			//	this command must always be sent if the engine stops searching, also in pondering mode if there is a
			//	"stop" command, so for every "go" command a "bestmove" command is needed!
			//	Directly before that the engine should send a final "info" command with the final search information,
			//	the the GUI has the complete statistics about the last search.

			d.out <- printPV(pv)
			d.out <- fmt.Sprintf("bestmove %v", printMove(pv.Moves[0]))
		} else {
			// No PV. Position is checkmate or stalemate. Send NullMove.

			d.out <- fmt.Sprintf("bestmove 0000")
		}
	} // else: stale or duplicate result
}

func printPV(pv search.PV) string {
	// "info depth 2 score cp 214 time 1242 nodes 2124 nps 34928 pv e2e4 e7e5 g1f3"

	parts := []string{"info"}
	if len(pv.Moves) > 0 {
		parts = append(parts, fmt.Sprintf("depth %v", len(pv.Moves)))
	}
	parts = append(parts, fmt.Sprintf("score cp %v", int16(pv.Score)))
	if pv.Nodes > 0 {
		parts = append(parts, fmt.Sprintf("nodes %v", pv.Nodes))
	}
	if pv.Time > 0 {
		parts = append(parts, fmt.Sprintf("time %v", pv.Time.Milliseconds()))
	}
	if pv.Nodes > 0 && pv.Time > 0 {
		parts = append(parts, fmt.Sprintf("nps %v", uint64(time.Second)*pv.Nodes/uint64(pv.Time)))
	}
	if len(pv.Moves) > 0 {
		parts = append(parts, "pv")
		parts = append(parts, board.FormatMoves(pv.Moves, printMove))
	}

	return strings.Join(parts, " ")
}

func printMove(m board.Move) string {
	return fmt.Sprintf("%v%v%v", m.From, m.To, printPromoPiece(m.Promotion))
}

func printPromoPiece(p board.Piece) string {
	switch p {
	case board.Queen:
		return "q"
	case board.Rook:
		return "r"
	case board.Knight:
		return "n"
	case board.Bishop:
		return "b"
	default:
		return ""
	}

}

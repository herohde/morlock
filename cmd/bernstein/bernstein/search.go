package bernstein

import (
	"context"
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/search"
)

type PlausibleMoveTable struct {
	Limit int
}

func (p PlausibleMoveTable) Explore(ctx context.Context, b *board.Board) (board.MovePriorityFn, board.MovePredicateFn) {
	pmt := FindPlausibleMoves(b)
	return search.Selection(truncate(pmt, p.Limit))
}

func truncate[T any](list []T, limit int) []T {
	if len(list) > limit {
		return list[:limit]
	}
	return list
}

// FindPlausibleMoves returns the best plausible legal moves in the given position, up to a limit,
// using the Plausible Move questions in order:
//
//	(1) Is the King in check?
//	(2) a) Can material be gained?
//	    b) Can material be lost?
//	    c) Can material be exchanged?
//	(3) Is castling possible?
//	(4) Can minor pieces be developed?
//	(5) Can key squares be occupied? Key squares are those squares which are controlled
//	    by diagonally connected pawns.)
//	(6) Can open files be occupied or invaded?
//	(7) Can pawns be moved?
//	(8) Can pieces be moved?
//
// If 3 is true, no further moves are considered.
//
// "The reason for stopping the decision routine, if castling is possible, is that
// the castling move does very little to enhance the score of a position, but is
// nevertheless a very important element in bringing the king to safety. Therefore,
// whenever castling is possible no other alternatives except for material exchanges
// are given, and eventually, when there are no exchanges or pieces to be gotten
// out of attack, the program is forced to castle."
//
// As a special case, the initial position generates center pawn moves even
// tough all pawn moves are otherwise considered equally.
func FindPlausibleMoves(b *board.Board) []board.Move {
	pos := b.Position()
	side := b.Turn()

	// NOTE(herohde) 11/24/2023: not mentioned, but probably not under-promoting. Based on the
	// Table 1 presentation, it is not clear whether the program can handle 2 queens at all.
	// Also no explicit allowance for mating moves.

	moves := board.FindMoves(pos.LegalMoves(side), board.Move.IsNotUnderPromotion)
	board.SortByPriority(moves, TA1(side)) // square order
	board.SortByPriority(moves, Table1)    // center pawn preference

	//	(1) Is the King in check?

	if pos.IsChecked(side) {
		// All legal moves necessarily gets out of check. Preference: capture, block then flee.

		fn := func(move board.Move) board.MovePriority {
			switch {
			case move.IsCaptureOrEnPassant():
				return 2
			case move.Piece == board.King:
				return 0
			default:
				return 1
			}
		}
		board.SortByPriority(moves, fn)
		return moves
	}

	//	(2) Can material be gained, lost or exchanged?
	//	(3) Is castling possible? Is so, stop.

	// TODO(herohde) 11/24/2023: unclear to what extent static exchange evaluation is performed.
	// For now, keep it simple and explore how it predicts the published games.

	gain := func(move board.Move) bool {
		switch move.Type {
		case board.CapturePromotion, board.Promotion:
			return true // not explicitly mentioned, but we consider promotions even if attacked
		case board.Capture:
			return MaterialValue(move.Capture) > MaterialValue(move.Piece) || !pos.IsAttacked(side, move.To)
		case board.EnPassant:
			return !pos.IsAttacked(side, move.To)
		default:
			return false
		}
	}

	loss := func(move board.Move) bool {
		if pos.IsAttacked(side, move.From) {
			return !pos.IsAttacked(side, move.To)
		}
		return false
	}

	exchange := func(move board.Move) bool {
		if move.Type == board.Capture {
			return MaterialValue(move.Capture) == MaterialValue(move.Piece) && pos.IsAttacked(side, move.To)
		}
		return false
	}

	rank := map[board.Move]board.MovePriority{}
	castle := false

	for _, move := range moves {
		switch {
		case gain(move):
			rank[move] = 23
		case loss(move):
			rank[move] = 22
		case exchange(move):
			rank[move] = 21
		case move.IsCastle():
			rank[move] = 20
			castle = true
		default:
			// skip
		}
	}

	if castle {
		moves = board.FindMoves(moves, func(move board.Move) bool {
			return rank[move] > 0 // limit moves to rule 2 and 3
		})

		board.SortByPriority(moves, func(move board.Move) board.MovePriority {
			return rank[move]
		})
		return moves
	}

	//	(4) Can minor pieces be developed?
	//	(5) Can squares defended by the pawns of a pawn chain be occupied?
	//	(6) Can open files be occupied or invaded?
	//	(7) Can pawns be moved?
	//	(8) Can pieces be moved?

	// TODO(herohde) 11/25/2023: randomize "any move"? If not, Table1 ordering severely hampers
	// exploration -- notably if ahead: a "free" King can take up 8 moves. Then there is no driver
	// for progress.

	develop := func(move board.Move) bool {
		return move.Piece.IsBishopOrKnight() && move.From.Rank() == board.PromotionRank(side.Opponent())
	}

	for _, move := range moves {
		if _, ok := rank[move]; ok {
			continue // skip: covered by rule 2-3
		}

		switch {
		case develop(move):
			rank[move] = 11
		case move.Piece == board.Pawn:
			rank[move] = 10
		default:
			// any move
		}
	}

	board.SortByPriority(moves, func(move board.Move) board.MovePriority {
		return rank[move]
	})
	return moves
}

// TA1 captures the board representation table (TA1) bias towards the opponent end (from H7
// to A1 if white). It appears the move generation favors moves by destination square:
//
//	(1) "At the start of a game the machine's pieces always reside on the squares 00
//	     to 07, 10 to 17; the opponent's on the squares 60 ot 67, 70 ot 77."
//	(2) "The first word of TA1 refers to square 77, the last word to square 00."
func TA1(side board.Color) board.MovePriorityFn {
	return func(move board.Move) board.MovePriority {
		if side == board.White {
			return board.MovePriority(move.To.Rank().V()*8 + move.To.File().V())
		} else {
			return board.MovePriority((8-move.To.Rank().V())*8 + (8 - move.To.File().V()))
		}
	}
}

// Table1 lists the pieces as considered: KQBKRP. The pawn center priority may stem from the
// order of the 8 pawns in that table. This baseline order provides a cutoff, notably rule 7 and 8.
func Table1(move board.Move) board.MovePriority {
	switch move.Piece {
	/*
		// NOTE(herohde) 11/25/2023: ignore pieces for now.

		case board.King:
			return 14
		case board.Queen:
			return 13
		case board.Bishop:
			return 12
		case board.Knight:
			return 11
		case board.Rook:
			return 10
	*/
	case board.Pawn:
		switch move.From.File() {
		case board.FileA:
			return 1
		case board.FileB:
			return 3
		case board.FileC:
			return 5
		case board.FileD:
			return 7
		case board.FileE:
			return 8
		case board.FileF:
			return 6
		case board.FileG:
			return 4
		case board.FileH:
			return 2
		default:
			return 0
		}
	default:
		return 0
	}
}

// NOTES for Scientific American game:
//
// Moves:
//  1. Book preference.
//
//  2. How to justify bishop move Bf1c4? 4 knight moves + 5 destinations for the bishop. Is there
//     a development preference? Longest move? Not 2nd rank? Luck? We do prune 2 options.
//     Is there a preference among minor piece development (bishop > knight). Table 1 lists
//     pieces in KQBKRP order.
//
//  3. Move 3: why pick d2d3? Are center pawn moves preferred throughout or just in the opening move?
//     Or does _forming_ pawn chains of length 3+, say, have priority?
//
//  4. Same question for Bishop move Bc1g5 as 2.
//
//  5. Bishop x knight. Exchange.
//
//  6. Kg1f3. Piece development. 4 options. Q: Does moving the bishop to a _different_ pawn chain square
//     count as rule 5? It would eat up options fast.
//
//  7. Castling.
//
//  8. Pawn exchange.
//
//  9. Bc4b5+. Avoid material loss. 2 options. 2 for knight development. Then pawn moves? How does the
//     "capture" of the pawn on e5 rank? Pruned as a move because it's defended?
//
// 10. c2c4. Why not exchange? It does form a pawn chain, so it suggests that is a preference.
//
// 11. Bishop exchange Bb4xc5. Why not free pawn capture? Both are options, though.
//
// 12. Pawn capture d3b4.
//
// 13. Knight moves away from material loss. 6 destinations.
//
// 14. Knight moves away from material loss. 3 destinations, but only 1 is not attacked.
//
// 15. Pawn move f2f3. Why? We have 1 "free" capture and 2 knight development. If a pawn move,
//     There are lots of options. Center preference? Or "not attacked" preference. Why not form
//     a pawn chain with b2b3?
//
// 16. Rf1e1. Why? It seems "invading" open files mean that opponent pawn is still there.
//     I.e., open files refer to just our own pawns. Maybe a preference there.
//
// 17. Nb1c3. Minor piece development.
//
// 18. Kh3f2. Check. Move or block. 3 legal moves, so it comes down to search.
//
// 19. g2g3. Why? Not in check and Queen has 7 destinations, but only 1 or 2 "safe" ones. Hence,
//     move feasibility apparently discards moving into (apparent) material loss upfront. Which
//     makes sense given the TA tables with attack/defence information. OTOH the g2 pawn is
//     at risk of loss as well, so it could be chance in material loss avoidance.
//
// 20. Nc3xd1. Material gain. Not rook? Lowest attacker/not further loss found in search or
//     preference?
//
// 21. b2b3. Pawn chain or center preference? Unless 2b means pieces en prise?
//
// 22. h2h4. Pawn move.

// NOTES for incomplete Chess Review game:
//
// Moves:
//  1. Book preference.
//
//  2. How to justify bishop move Bf1b5? It is different than the other game, so the engine
//     is either _not_ deterministic or has changed between these games. It's still 4 knight
//     moves + 5 destinations for the bishop, so signs point to bishop preference.
//
//  3. d2d4. The lack of protecting the en prise pawn suggests that defending is not an option
//     and the pawn cannot move. Why not d2d3? No pawn chain preference, it seems. But just
//     middle-out pawn ordering. Random pawn selection is also an option, but it seems unlikely.
//
//  4. Material gain of en prise knight.
//
//  5. Bb5c4. Material loss avoidance. 5 options + pawn move.
//
//  6. Bc1c2. Blocking check. 4 options. Bishop or search (mobility) preference?
//
//  7. Bc4d3. Material loss. Also 4 options. Why not c4b3?
//
//  8. Kb1c3. Material loss of b2 generates b2b3, b2b4 but they search poorly due to rook capture.
//     Development move that lucks into preventing material loss?
//
//  9. f2f4? Why not 0-0-0? That is wrong, unless there is a condition that prevents castling
//     if it would leave a piece en prise. Seems like the engine leans into the static exchange info.
//
// 10. Exchange.
//
// 11. Check capture.
//
// 12. Kg1f3. Minor development. Open rank-rook move an option as well.
//
// 13. Exchange.
//
// 14. e4e5. Why not rook to half-open file? Does our own piece block that or search better?
//
// 15. h2h4. Why this pawn move. Area control?
//
// 16. h4h5. Understands rook still covering -- otherwise, it looks like an en prise move.
//     Alternatively, we don't filter them out and there just happen to be 7 pawn moves.

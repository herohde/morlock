package search_test

import (
	"github.com/herohde/morlock/pkg/board"
	"github.com/herohde/morlock/pkg/search"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMVVLVA(t *testing.T) {
	nb := board.Move{Type: board.Normal, Piece: board.Bishop}
	nq := board.Move{Type: board.Normal, Piece: board.Queen}
	cqb := board.Move{Type: board.Capture, Piece: board.Queen, Capture: board.Bishop}
	crb := board.Move{Type: board.Capture, Piece: board.Rook, Capture: board.Bishop}
	ckb := board.Move{Type: board.Capture, Piece: board.Knight, Capture: board.Bishop}
	cqp := board.Move{Type: board.Capture, Piece: board.Queen, Capture: board.Pawn}
	crp := board.Move{Type: board.Capture, Piece: board.Rook, Capture: board.Pawn}
	pb := board.Move{Type: board.Promotion, Piece: board.Pawn, Promotion: board.Bishop}
	pr := board.Move{Type: board.Promotion, Piece: board.Pawn, Promotion: board.Rook}
	pq := board.Move{Type: board.Promotion, Piece: board.Pawn, Promotion: board.Queen}
	cpqr := board.Move{Type: board.CapturePromotion, Piece: board.Pawn, Promotion: board.Queen, Capture: board.Rook}
	cprb := board.Move{Type: board.CapturePromotion, Piece: board.Pawn, Promotion: board.Rook, Capture: board.Bishop}
	cpqb := board.Move{Type: board.CapturePromotion, Piece: board.Pawn, Promotion: board.Queen, Capture: board.Bishop}
	ep := board.Move{Type: board.EnPassant, Piece: board.Pawn}

	tests := []struct {
		in, out []board.Move
	}{
		{[]board.Move{nb, nq, ep}, []board.Move{ep, nb, nq}},
		{[]board.Move{cqb, crb, ckb, cqp, crp}, []board.Move{ckb, crb, cqb, crp, cqp}},
		{[]board.Move{pb, pr, pq}, []board.Move{pq, pr, pb}},
		{[]board.Move{cpqr, cprb, cpqb}, []board.Move{cpqr, cpqb, cprb}},

		{
			[]board.Move{nb, nq, cqb, crb, ckb, cqp, crp, pb, pr, pq, cpqr, cprb, cpqb, ep},
			[]board.Move{cpqr, cpqb, pq, cprb, pr, ckb, crb, cqb, pb, ep, crp, cqp, nb, nq},
		},
	}

	for _, tt := range tests {
		var moves []board.Move
		list := board.NewMoveList(tt.in, search.MVVLVA)
		for {
			move, ok := list.Next()
			if !ok {
				break
			}
			moves = append(moves, move)
		}
		assert.Equal(t, board.PrintMoves(moves), board.PrintMoves(tt.out))
	}
}

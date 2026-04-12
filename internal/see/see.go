package see

import (
	. "github.com/ChizhovVadim/chesstools/pkg/chess"
)

var pieceValuesSEE = [...]int{Empty: 0, Pawn: 1, Knight: 4, Bishop: 4, Rook: 6, Queen: 12, King: 120}

// Static Exchange Evaluation
// https://www.chessprogramming.org/Static_Exchange_Evaluation
func See(pos *Position, mv Move) int {
	var from = mv.From()
	var to = mv.To()
	var pc = mv.MovingPiece()
	var sd = pos.WhiteMove
	var sc = 0
	if mv.CapturedPiece() != Empty {
		sc += pieceValuesSEE[mv.CapturedPiece()]
	}
	if mv.Promotion() != Empty {
		pc = mv.Promotion()
		sc += pieceValuesSEE[pc] - pieceValuesSEE[Pawn]
	}
	var pieces = (pos.White | pos.Black) &^ SquareMask[from]
	sc -= seeRec(pos, !sd, to, pieces, pc)
	return sc
}

func seeRec(pos *Position, sd bool, to int, pieces uint64, cp int) int {
	var bs = 0
	var pc, from = getLeastValuableAttacker(pos, to, sd, pieces)
	if from != SquareNone {
		var sc = pieceValuesSEE[cp]
		if cp != King {
			sc -= seeRec(pos, !sd, to, pieces&^SquareMask[from], pc)
		}
		if sc > bs {
			bs = sc
		}
	}
	return bs
}

func getAttacks(p *Position, to int, side bool, occ uint64) uint64 {
	var att = (PawnAttacks(to, !side) & p.Pawns) |
		(KnightAttacks[to] & p.Knights) |
		(KingAttacks[to] & p.Kings) |
		(BishopAttacks(to, occ) & (p.Bishops | p.Queens)) |
		(RookAttacks(to, occ) & (p.Rooks | p.Queens))
	return p.PiecesByColor(side) & att
}

func getLeastValuableAttacker(p *Position, to int, side bool, occ uint64) (attacker, from int) {
	attacker = Empty
	from = SquareNone
	var att = getAttacks(p, to, side, occ) & occ
	if att == 0 {
		return
	}
	var newTarget = pieceValuesSEE[King] + 1
	for ; att != 0; att &= att - 1 {
		var f = FirstOne(att)
		var piece = p.WhatPiece(f)
		if pieceValuesSEE[piece] < newTarget {
			attacker = piece
			from = f
			newTarget = pieceValuesSEE[piece]
		}
	}
	return
}

package searchrand

import "github.com/ChizhovVadim/chesstools/pkg/chess"

func evalMaterial(p *chess.Position) int {
	var val = 1*(chess.PopCount(p.Pawns&p.White)-chess.PopCount(p.Pawns&p.Black)) +
		4*(chess.PopCount(p.Knights&p.White)-chess.PopCount(p.Knights&p.Black)) +
		4*(chess.PopCount(p.Bishops&p.White)-chess.PopCount(p.Bishops&p.Black)) +
		6*(chess.PopCount(p.Rooks&p.White)-chess.PopCount(p.Rooks&p.Black)) +
		12*(chess.PopCount(p.Queens&p.White)-chess.PopCount(p.Queens&p.Black))
	if !p.WhiteMove {
		val = -val
	}
	return val
}

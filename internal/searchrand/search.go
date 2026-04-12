package searchrand

import (
	"math/rand/v2"

	"github.com/ChizhovVadim/chesstools/pkg/chess"
	"github.com/ChizhovVadim/chesstools/pkg/game"
)

const (
	stackSize     = 8
	maxHeight     = stackSize - 1
	valueInfinity = 30_000
)

type Searcher struct {
	margin int
	stack  [stackSize]struct {
		position chess.Position
		moveList [chess.MaxMoves]chess.OrderedMove
	}
}

func New(margin int) *Searcher {
	return &Searcher{margin: margin}
}

func (t *Searcher) PlayRandomOpening(
	g *game.Game,
	openingSize int,
) bool {
	for range openingSize {
		var move = t.findMove(&g.Position)
		if move == chess.MoveEmpty ||
			!g.MakeMove(game.MoveItem{
				Move:      move,
				IsOpening: true,
			}) {
			return false
		}
	}
	return true
}

func (t *Searcher) findMove(position *chess.Position) chess.Move {
	const height = 0
	const beta = valueInfinity

	var ml = position.GenerateMoves(t.stack[height].moveList[:])
	var child = &t.stack[height+1].position
	var movesSearched int
	var best int

	for i := range ml {
		var move = ml[i].Move
		if !position.MakeMove(move, child) {
			continue
		}
		movesSearched += 1
		if movesSearched == 1 {
			var score = -t.quiescence(-beta, valueInfinity, height+1)
			ml[i].Key = int32(score)
			best = score
		} else {
			var score = -t.quiescence(-beta, -(best - t.margin - 1), height+1)
			ml[i].Key = int32(score)
			if score > best {
				best = score
			}
		}
	}
	var candidates []chess.Move
	for i := range ml {
		var move = ml[i].Move
		if !position.MakeMove(move, child) {
			continue
		}
		if int(ml[i].Key) >= best-t.margin {
			candidates = append(candidates, move)
		}
	}
	if len(candidates) == 0 {
		return chess.MoveEmpty
	}
	return candidates[rand.IntN(len(candidates))]
}

func (t *Searcher) quiescence(alpha, beta, height int) int {
	var position = &t.stack[height].position
	var eval = evalMaterial(position)
	if eval > alpha {
		alpha = eval
		if alpha >= beta {
			return alpha
		}
	}
	if height >= maxHeight {
		return eval
	}
	var ml = position.GenerateCaptures(t.stack[height].moveList[:])
	var child = &t.stack[height+1].position
	for i := range ml {
		var move = ml[i].Move
		if seeValue := see(position, move); seeValue < 0 {
			continue
		}
		if !position.MakeMove(move, child) {
			continue
		}
		var score = -t.quiescence(-beta, -alpha, height+1)
		if score > alpha {
			alpha = score
			if alpha >= beta {
				break
			}
		}
	}
	return alpha
}

package game

import (
	"maps"
	"slices"
	"time"

	"github.com/ChizhovVadim/chesstools/pkg/chess"
)

const (
	GameResultNone = iota
	GameResultDraw
	GameResultWhiteWins
	GameResultBlackWins
)

type Game struct {
	Date          time.Time
	Round         int
	White         string
	Black         string
	StartFen      string
	Moves         []MoveItem
	Position      chess.Position
	Repeats       map[uint64]int // включая текущую позицию
	ResultComment string
	Result        int
}

type MoveItem struct {
	Move      chess.Move
	Score     Score
	Depth     int
	IsOpening bool
}

type Score struct {
	Centipawns int
	Mate       int
}

func NewGame(startFen string) (Game, error) {
	if startFen == "" {
		startFen = chess.InitialPositionFen
	}
	var startPos, err = chess.NewPositionFromFEN(startFen)
	if err != nil {
		return Game{}, err
	}
	return Game{
		StartFen: startFen,
		Position: startPos,
		Repeats:  map[uint64]int{startPos.Key: 1},
		Result:   GameResultNone,
	}, nil
}

func (g *Game) Clone() Game {
	return Game{
		Date:          g.Date,
		Round:         g.Round,
		White:         g.White,
		Black:         g.Black,
		StartFen:      g.StartFen,
		Moves:         slices.Clone(g.Moves),
		Position:      g.Position,
		Repeats:       maps.Clone(g.Repeats),
		ResultComment: g.ResultComment,
		Result:        g.Result,
	}
}

func (g *Game) WhiteTurn() bool {
	return g.Position.WhiteMove
}

func (g *Game) MakeMove(item MoveItem) bool {
	var ml = g.Position.GenerateLegalMoves()
	for _, mv := range ml {
		if mv == item.Move {
			var child chess.Position
			g.Position.MakeMove(mv, &child)
			g.Moves = append(g.Moves, item)
			g.Position = child
			if mv.MovingPiece() == chess.Pawn || mv.CapturedPiece() != chess.Empty {
				clear(g.Repeats)
			}
			g.Repeats[g.Position.Key] += 1
			g.updateResult()
			return true
		}
	}
	return false
}

func (g *Game) updateResult() {
	if !hasLegalMove(&g.Position) {
		if g.Position.IsCheck() {
			if g.Position.WhiteMove {
				g.Result = GameResultBlackWins
				g.ResultComment = "checkmate"
				return
			} else {
				g.Result = GameResultWhiteWins
				g.ResultComment = "checkmate"
				return
			}
		} else {
			g.Result = GameResultDraw
			g.ResultComment = "stalemate"
			return
		}
	}
	if g.Position.Rule50 >= 100 {
		g.Result = GameResultDraw
		g.ResultComment = "50 moves"
		return
	}
	if isLowMaterial(&g.Position) {
		g.Result = GameResultDraw
		g.ResultComment = "low material"
		return
	}
	if g.Repeats[g.Position.Key] >= 3 {
		g.Result = GameResultDraw
		g.ResultComment = "3 fold repetition"
		return
	}
}

func isLowMaterial(p *chess.Position) bool {
	if (p.Pawns|p.Rooks|p.Queens) == 0 &&
		!chess.MoreThanOne(p.Knights|p.Bishops) {
		return true
	}

	return false
}

func hasLegalMove(p *chess.Position) bool {
	var buf [chess.MaxMoves]chess.OrderedMove
	var child chess.Position
	var ml = p.GenerateMoves(buf[:])
	for i := range ml {
		if p.MakeMove(ml[i].Move, &child) {
			return true
		}
	}
	return false
}

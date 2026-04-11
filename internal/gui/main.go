package gui

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ChizhovVadim/chesstools/pkg/chess"
	"github.com/ChizhovVadim/chesstools/pkg/game"
	"github.com/ChizhovVadim/chesstools/pkg/uci"
)

func Run(eng *uci.Service) error {
	var g, err = game.NewGame("")
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(os.Stdin)
	for g.Result == game.GameResultNone {
		PrintPosition(&g.Position)
		if g.WhiteTurn() {
			if !scanner.Scan() {
				return nil
			}
			line := scanner.Text()
			if line == "quit" {
				return nil
			}
			var mv = g.Position.ParseMoveLAN(line)
			if mv == chess.MoveEmpty {
				continue
			}
			g.MakeMove(game.MoveItem{Move: mv})
		} else {
			eng.Position(&g)
			var sr, err = eng.Go(&g, uci.LimitsType{MoveTime: 3_000}, nil)
			if err != nil {
				return err
			}
			fmt.Println(sr)
			if !g.MakeMove(game.MoveItem{Move: sr.BestMove}) {
				return fmt.Errorf("bad search result %v", sr)
			}
		}
	}
	return nil
}

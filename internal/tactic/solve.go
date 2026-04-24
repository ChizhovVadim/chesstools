package tactic

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/ChizhovVadim/chesstools/pkg/game"
	"github.com/ChizhovVadim/chesstools/pkg/uci"
)

func Solve(tests []EpdItem, eng *uci.Service, moveTime time.Duration) error {
	log.Println("Solve tactic started")
	start := time.Now()
	defer func() {
		log.Println("Solve tactic finished",
			time.Since(start))
	}()

	var total, solved int
	for _, test := range tests {
		var searchResult = executeTest(test, eng, moveTime)
		if slices.Contains(test.BestMoves, searchResult.BestMove) {
			solved += 1
		}
		total += 1
		fmt.Println(test.Content)
		fmt.Printf("%+v\n", searchResult)
		fmt.Printf("Solved: %v, Total: %v\n", solved, total)
		fmt.Println()
	}

	return nil
}

func executeTest(test EpdItem, uciEngine *uci.Service, moveTime time.Duration) uci.SearchResult {
	var game, err = game.NewGame(test.Position.String())
	if err != nil {
		log.Fatal(err)
	}
	uciEngine.Position(&game)
	var limit = uci.LimitsType{
		MoveTime: int(moveTime.Milliseconds()),
	}
	res, err := uciEngine.Go(&game, limit, nil)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

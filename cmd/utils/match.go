package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/ChizhovVadim/chesstools/internal/arena"
	"github.com/ChizhovVadim/chesstools/internal/cli"
	"github.com/ChizhovVadim/chesstools/pkg/uci"
)

// играет матч между двумя движками
func matchHandler(args []string) error {
	var date = time.Now()
	var (
		openingsPath   = ""
		outputGamePath = cli.MapPath(fmt.Sprintf("~/chess/games/match-%v.pgn", date.Format("2006-01-02_15_04")))
		concurrency    = runtime.GOMAXPROCS(0)
		timeLimit      = uci.LimitsType{Nodes: 100_000}
		playerA        = ""
		playerB        = ""
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&playerA, "playerA", playerA, "")
	flagset.StringVar(&playerB, "playerB", playerB, "")
	flagset.Parse(args)

	var di = &diContainer{}
	return arena.PlayMatch(context.Background(), concurrency, openingsPath, outputGamePath,
		func() *uci.Process {
			return di.BuildEngine(playerA)
		},
		func() *uci.Process {
			return di.BuildEngine(playerB)
		},
		timeLimit,
	)
}

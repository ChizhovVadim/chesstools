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
		nodes          = 100_000
		playerA        = ""
		playerB        = ""
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.IntVar(&nodes, "nodes", nodes, "")
	flagset.StringVar(&playerA, "playerA", playerA, "")
	flagset.StringVar(&playerB, "playerB", playerB, "")
	flagset.Parse(args)

	var tl = uci.LimitsType{Nodes: nodes}

	var di = &diContainer{}
	return arena.PlayMatch(context.Background(), concurrency, openingsPath, outputGamePath,
		newPlayer("", tl, di.MustEngineConfig(playerA)),
		newPlayer("", tl, di.MustEngineConfig(playerB)),
	)
}

func newPlayer(
	name string,
	timeLimit uci.LimitsType,
	engineConfig EngineConfig,
) arena.PlayerConfig {
	if name == "" {
		name = engineConfig.Name
	}
	return arena.PlayerConfig{
		Name:      name,
		TimeLimit: timeLimit,
		Command:   engineConfig.Command,
		Arg:       engineConfig.Arg,
		Options:   engineConfig.Options,
	}
}

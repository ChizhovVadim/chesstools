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

// генерирует датасет для обучения нейронной сети
func datasetHandler(args []string) error {
	var date = time.Now()
	var (
		concurrency    = runtime.GOMAXPROCS(0)
		gamesCount     = 100
		openingSize    = 9
		timeLimit      = uci.LimitsType{Depth: 8}
		player         = ""
		outputGamePath = cli.MapPath(fmt.Sprintf("~/chess/dataset/arena-%v.pgn", date.Format("2006-01-02_15_04")))
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&player, "player", player, "")
	flagset.IntVar(&gamesCount, "games_count", gamesCount, "")
	flagset.IntVar(&openingSize, "opening_size", openingSize, "")
	flagset.Parse(args)

	var di = &diContainer{}
	return arena.GenerateDataset(context.Background(), concurrency, gamesCount, openingSize,
		func() *uci.Process { return di.BuildEngine(player) },
		timeLimit,
		outputGamePath,
	)
}

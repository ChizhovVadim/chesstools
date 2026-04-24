package dataset

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ChizhovVadim/chesstools/pkg/game"
	"github.com/ChizhovVadim/chesstools/pkg/pgn"
	"github.com/ChizhovVadim/chesstools/pkg/uci"
	"golang.org/x/sync/errgroup"
)

type PlayerConfig struct {
	Name      string
	TimeLimit uci.LimitsType
	Command   string
	Arg       string
	Options   []uci.Option
}

func GenerateDataset(
	ctx context.Context,
	concurrency int,
	gamesCount int,
	openingSize int,
	playerConfig PlayerConfig,
	outputGamePath string,
) error {
	log.Println("GenerateDataset started")
	start := time.Now()
	defer func() {
		log.Println("GenerateDataset finished",
			time.Since(start))
	}()

	g, ctx := errgroup.WithContext(ctx)
	var games = make(chan game.Game)
	g.Go(func() error {
		return saveGames(ctx, outputGamePath, games)
	})

	var gameIndex = new(int32)
	var wg = &sync.WaitGroup{}
	for range concurrency {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			return playDatasetGames(ctx, gameIndex, int32(gamesCount), openingSize, playerConfig, games)
		})
	}
	g.Go(func() error {
		wg.Wait()
		close(games)
		return nil
	})

	return g.Wait()
}

func saveGames(
	_ context.Context,
	outputGamePath string,
	games <-chan game.Game,
) error {
	f, err := os.OpenFile(outputGamePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for g := range games {
		log.Printf("Finished game %v\n",
			g.Round)

		var err = pgn.Write(&g, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func playDatasetGames(
	ctx context.Context,
	gameIndex *int32,
	gamesCount int32,
	openingSize int,
	playerConfig PlayerConfig,
	games chan<- game.Game,
) error {
	var uciProcess, err = uci.Start(playerConfig.Command, playerConfig.Arg)
	if err != nil {
		return err
	}
	defer uciProcess.Close()
	var uciService = uci.NewService(uciProcess.Reader(), uciProcess.Writer())
	if err := uciService.Init(playerConfig.Options); err != nil {
		return err
	}

	var searcher = NewSearcher(1)
	for {
		var g, _ = game.NewGame("")
		g.Date = time.Now()
		g.White = playerConfig.Name
		g.Black = playerConfig.Name

		if ok := searcher.PlayRandomOpening(&g, openingSize); !ok {
			continue
		}

		var err = playGame(&g, uciService, playerConfig.TimeLimit)
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case games <- g:
			if atomic.AddInt32(gameIndex, 1) >= gamesCount {
				return nil
			}
		}
	}
}

func playGame(g *game.Game, player *uci.Service, tl uci.LimitsType) error {
	player.UciNewgame()

	for g.Result == game.GameResultNone {
		player.Position(g)
		var sr, err = player.Go(g, tl, nil)
		if err != nil {
			return err
		}
		if !g.MakeMove(game.MoveItem{
			Move: sr.BestMove,
			Score: game.Score{
				Mate:       sr.Score.Mate,
				Centipawns: sr.Score.Centipawns,
			},
			Depth: sr.Depth,
		}) {
			return fmt.Errorf("illegal engine move %v", sr)
		}
	}

	return nil
}

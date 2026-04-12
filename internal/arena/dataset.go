package arena

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ChizhovVadim/chesstools/internal/searchrand"
	"github.com/ChizhovVadim/chesstools/pkg/game"
	"github.com/ChizhovVadim/chesstools/pkg/uci"
	"golang.org/x/sync/errgroup"
)

func GenerateDataset(
	ctx context.Context,
	concurrency int,
	gamesCount int,
	openingSize int,
	playerBuilder func() *uci.Process,
	timeLimit uci.LimitsType,
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
			return playDatasetGames(ctx, gameIndex, int32(gamesCount), openingSize, playerBuilder, timeLimit, games)
		})
	}
	g.Go(func() error {
		wg.Wait()
		close(games)
		return nil
	})

	return g.Wait()
}

func playDatasetGames(
	ctx context.Context,
	gameIndex *int32,
	gamesCount int32,
	openingSize int,
	playerBuilder func() *uci.Process,
	timeLimit uci.LimitsType,
	games chan<- game.Game,
) error {
	var player = playerBuilder()
	defer player.Close()
	if err := player.Init(); err != nil {
		return err
	}

	var searcher = searchrand.New(1)

	for {
		var g, _ = game.NewGame("")
		g.Date = time.Now()
		g.White = player.Name()
		g.Black = g.White

		if ok := searcher.PlayRandomOpening(&g, openingSize); !ok {
			continue
		}

		var err = playGame(&g, player.Service, player.Service, timeLimit)
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

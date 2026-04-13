package arena

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ChizhovVadim/chesstools/pkg/game"
	"github.com/ChizhovVadim/chesstools/pkg/pgn"
	"github.com/ChizhovVadim/chesstools/pkg/uci"

	"golang.org/x/sync/errgroup"
)

func PlayMatch(
	ctx context.Context,
	concurrency int,
	openingsPath string,
	outputGamePath string,
	playerABuilder, playerBBuilder func() *uci.Process,
	timeLimit uci.LimitsType,
) error {
	log.Println("PlayMatch started")
	start := time.Now()
	defer func() {
		log.Println("PlayMatch finished",
			time.Since(start))
	}()

	var openings = make(chan game.Game)
	var games = make(chan game.Game)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(openings)
		return prepareOpenings(ctx, openingsPath, openings)
	})
	g.Go(func() error {
		return saveGames(ctx, outputGamePath, games)
	})

	var wg = &sync.WaitGroup{}
	for range concurrency {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			return playGames(ctx, playerABuilder, playerBBuilder, timeLimit, openings, games)
		})
	}
	g.Go(func() error {
		wg.Wait()
		close(games)
		return nil
	})

	return g.Wait()
}

func prepareOpenings(
	ctx context.Context,
	openingsPath string,
	openings chan<- game.Game,
) error {
	var r io.Reader
	if openingsPath == "" {
		r = strings.NewReader(DefaultOpenings)
	} else {
		var file, err = os.Open(openingsPath)
		if err != nil {
			return err
		}
		defer file.Close()
		r = file
	}
	data, err := pgn.LoadOpenings(r)
	if err != nil {
		return err
	}
	log.Println("Openings loaded", "size", len(data))

	for openingIndex, opening := range data {
		// каждый дебют за оба цвета.
		for i := 0; i < 2; i += 1 {
			var g = opening.Clone()
			g.Round = 1 + 2*openingIndex + i
			select {
			case <-ctx.Done():
				return ctx.Err()
			case openings <- g:
			}
		}
	}

	return nil
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

	var totalGames = 0
	var wins, losses, draws int

	for g := range games {
		log.Printf("Finished game %v: %v {%v}\n",
			g.Round,
			pgn.GameResultName(g.Result),
			g.ResultComment)

		totalGames += 1
		if g.Result == game.GameResultDraw {
			draws += 1
		} else if g.Result == game.GameResultWhiteWins && g.Round%2 == 1 ||
			g.Result == game.GameResultBlackWins && g.Round%2 != 1 {
			wins += 1
		} else {
			losses += 1
		}

		var stat = computeStat(wins, losses, draws)
		log.Printf("Score: %v - %v - %v  [%.3f] %v\n",
			wins, losses, draws, stat.winningFraction, totalGames)
		log.Printf("Elo difference: %.1f, LOS: %.1f %%\n",
			stat.eloDifference, stat.los*100)

		var err = pgn.Write(&g, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func playGames(
	ctx context.Context,
	playerABuilder, playerBBuilder func() *uci.Process,
	timeLimit uci.LimitsType,
	openings <-chan game.Game,
	games chan<- game.Game,
) error {
	var playerA = playerABuilder()
	defer playerA.Close()
	if err := playerA.Init(); err != nil {
		return err
	}

	var playerB = playerBBuilder()
	defer playerB.Close()
	if err := playerB.Init(); err != nil {
		return err
	}

	for g := range openings {
		var white, black *uci.Process
		if g.Round%2 == 1 {
			white = playerA
			black = playerB
		} else {
			white = playerB
			black = playerA
		}
		g.Date = time.Now()
		g.White = white.Name()
		g.Black = black.Name()
		var err = playGame(&g, white.Service, black.Service, timeLimit)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case games <- g:
		}
	}

	return nil
}

func playGame(g *game.Game, white, black *uci.Service, timeLimit uci.LimitsType) error {
	white.UciNewgame()
	black.UciNewgame()

	for g.Result == game.GameResultNone {
		var activePlayer *uci.Service
		if g.WhiteTurn() {
			activePlayer = white
		} else {
			activePlayer = black
		}
		activePlayer.Position(g)
		var sr, err = activePlayer.Go(g, timeLimit, nil)
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

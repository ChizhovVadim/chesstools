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

type PlayerConfig struct {
	Name      string
	TimeLimit uci.LimitsType
	Command   string
	Arg       string
	Options   []uci.Option
}

type Player struct {
	Name      string
	TimeLimit uci.LimitsType
	*uci.Service
}

func PlayMatch(
	ctx context.Context,
	concurrency int,
	openingsPath string,
	outputGamePath string,
	playerA, playerB PlayerConfig,
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
			return playGames(ctx, playerA, playerB, openings, games)
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
	playerAConfig, playerBConfig PlayerConfig,
	openings <-chan game.Game,
	games chan<- game.Game,
) error {
	var configs = [2]PlayerConfig{playerAConfig, playerBConfig}
	var players [2]Player

	for i, config := range configs {
		var eng, err = uci.Start(config.Command, config.Arg)
		if err != nil {
			return err
		}
		defer eng.Close()

		var player = Player{
			Name:      config.Name,
			TimeLimit: config.TimeLimit,
			Service:   uci.NewService(eng.Reader(), eng.Writer()),
		}
		if err := player.Service.Init(config.Options); err != nil {
			return err
		}
		players[i] = player
	}

	for g := range openings {
		var white, black Player
		if g.Round%2 == 1 {
			white = players[0]
			black = players[1]
		} else {
			white = players[1]
			black = players[0]
		}
		var err = playGame(&g, white, black)
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

func playGame(g *game.Game, white, black Player) error {
	g.Date = time.Now()
	g.White = white.Name
	g.Black = black.Name

	white.UciNewgame()
	black.UciNewgame()

	for g.Result == game.GameResultNone {
		var activePlayer Player
		if g.WhiteTurn() {
			activePlayer = white
		} else {
			activePlayer = black
		}
		activePlayer.Position(g)
		var sr, err = activePlayer.Go(g, activePlayer.TimeLimit, nil)
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

package main

import (
	"flag"
	"time"

	"github.com/ChizhovVadim/chesstools/internal/cli"
	"github.com/ChizhovVadim/chesstools/internal/tactic"
	"github.com/ChizhovVadim/chesstools/pkg/uci"
)

// решает тактические тесты
func tacticHandler(args []string) error {
	var (
		tacticTestsPath = cli.MapPath("~/chess/tests/tests.epd")
		moveTime        = 3 * time.Second
		engineKey       = ""
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&tacticTestsPath, "path", tacticTestsPath, "")
	flagset.DurationVar(&moveTime, "time", moveTime, "")
	flagset.StringVar(&engineKey, "engine", engineKey, "")
	flagset.Parse(args)

	var tests, err = tactic.Load(tacticTestsPath)
	if err != nil {
		return err
	}

	var di = &diContainer{}

	var engConfig = di.MustEngineConfig(engineKey)
	process, err := uci.Start(engConfig.Command, engConfig.Arg)
	if err != nil {
		return err
	}
	defer process.Close()

	var service = uci.NewService(process.Reader(), process.Writer())
	if err := service.Init(engConfig.Options); err != nil {
		return err
	}

	return tactic.Solve(tests, service, moveTime)
}

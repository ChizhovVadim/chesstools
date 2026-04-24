package main

import (
	"flag"

	"github.com/ChizhovVadim/chesstools/internal/gui"
	"github.com/ChizhovVadim/chesstools/pkg/uci"
)

// игра с uci движком в консоли
func guiHandler(args []string) error {
	var engineKey = ""

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&engineKey, "engine", engineKey, "")
	flagset.Parse(args)

	var di = &diContainer{}

	var engConfig = di.MustEngineConfig(engineKey)
	var process, err = uci.Start(engConfig.Command, engConfig.Arg)
	if err != nil {
		return err
	}
	defer process.Close()
	var service = uci.NewService(process.Reader(), process.Writer())
	if err := service.Init(engConfig.Options); err != nil {
		return err
	}
	return gui.Run(service)
}

package main

import (
	"flag"

	"github.com/ChizhovVadim/chesstools/internal/gui"
)

func guiHandler(args []string) error {
	var engineKey = ""

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&engineKey, "engine", engineKey, "")
	flagset.Parse(args)

	var di = &diContainer{}

	var eng = di.BuildEngine(engineKey)
	defer eng.Close()
	if err := eng.Init(); err != nil {
		return err
	}
	return gui.Run(eng.Service)
}

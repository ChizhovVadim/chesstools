package main

import (
	"log"

	"github.com/ChizhovVadim/chesstools/internal/cli"
)

func main() {
	var app = &cli.App{}
	app.AddCommand("gui", guiHandler)
	app.AddCommand("arena", matchHandler)
	app.AddCommand("dataset", datasetHandler)
	var err = app.Run()
	if err != nil {
		log.Println("run failed",
			"error", err)
		return
	}
}

package main

import (
	"github.com/ChizhovVadim/chesstools/internal/cli"
	"github.com/ChizhovVadim/chesstools/internal/quality"
	"github.com/ChizhovVadim/chesstools/pkg/uci"
)

// вычисляет ошибку оценочной функции на валидационном датасете
func qualityHandler(args []string) error {
	var (
		path         = cli.MapPath("~/Projects/counterdev/counter")
		sigmoidScale = 3.5 / 512
		datasetPath  = cli.MapPath("~/chess/tuner/quiet-labeled.epd")
	)

	var uciProcess, err = uci.Start(path, "")
	if err != nil {
		return err
	}
	defer uciProcess.Close()

	var data = quality.LoadValidationDataset(datasetPath)
	var service = quality.NewService(uciProcess.Reader(), uciProcess.Writer(), sigmoidScale)
	return quality.CheckEvalQuality(service, data)
}

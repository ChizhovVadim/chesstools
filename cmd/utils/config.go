package main

import (
	"encoding/json"
	"os"

	"github.com/ChizhovVadim/chesstools/pkg/uci"
)

type EngineConfig struct {
	Name    string
	Command string
	Arg     string
	Options []uci.Option
}

func loadEngineConfigs(path string) ([]EngineConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []EngineConfig
	err = json.NewDecoder(file).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

package main

import (
	"log"
	"slices"
	"strings"

	"github.com/ChizhovVadim/chesstools/internal/cli"
	"github.com/ChizhovVadim/chesstools/pkg/uci"
)

type diContainer struct {
	engineConfigs []EngineConfig
}

func (di *diContainer) EngineConfigs() []EngineConfig {
	if di.engineConfigs == nil {
		var cfg, err = loadEngineConfigs("engines.json")
		if err != nil {
			log.Fatalf("load engine configs failed %v", err)
		}
		di.engineConfigs = cfg
	}
	return di.engineConfigs
}

func (di *diContainer) BuildEngine(key string) *uci.Process {
	if key == "" {
		key = "counter55"
	}
	var configs = di.EngineConfigs()
	var index = slices.IndexFunc(configs, func(e EngineConfig) bool {
		return e.Name == key
	})
	if index == -1 {
		log.Fatalf("engine %v config not found", key)
	}
	var info = configs[index]
	var args []string
	if info.Arg != "" {
		args = strings.Fields(info.Arg)
	}
	return uci.NewProcess(key, cli.MapPath(info.Command), args, info.Options)
}

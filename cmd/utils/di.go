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
		var engineConfigs, err = loadEngineConfigs("engines.json")
		if err != nil {
			log.Fatalf("load engine configs failed %v", err)
		}
		for i := range engineConfigs {
			engineConfigs[i].Command = cli.MapPath(engineConfigs[i].Command)
		}
		di.engineConfigs = engineConfigs
	}
	return di.engineConfigs
}

func (di *diContainer) EngineBuilder(key string) func() *uci.Process {
	var configs = di.EngineConfigs()
	if len(configs) == 0 {
		log.Fatalf("empty engine configs")
	}
	var config EngineConfig
	if key == "" {
		config = configs[0]
	} else {
		var index = slices.IndexFunc(configs, func(e EngineConfig) bool {
			return e.Name == key
		})
		if index == -1 {
			log.Fatalf("engine %v config not found", key)
		}
		config = configs[index]
	}
	return func() *uci.Process {
		var args []string
		if config.Arg != "" {
			args = strings.Fields(config.Arg)
		}
		return uci.NewProcess(config.Name, config.Command, args, config.Options)
	}
}

func (di *diContainer) BuildEngine(key string) *uci.Process {
	return di.EngineBuilder(key)()
}

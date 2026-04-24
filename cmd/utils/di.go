package main

import (
	"log"
	"slices"

	"github.com/ChizhovVadim/chesstools/internal/cli"
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

func (di *diContainer) EngineConfig(key string) (EngineConfig, bool) {
	var configs = di.EngineConfigs()
	if len(configs) == 0 {
		return EngineConfig{}, false
	}
	if key == "" {
		return configs[0], true
	}
	var index = slices.IndexFunc(configs, func(e EngineConfig) bool {
		return e.Name == key
	})
	if index == -1 {
		return EngineConfig{}, false
		//log.Fatalf("engine %v config not found", key)
	}
	return configs[index], true
}

func (di *diContainer) MustEngineConfig(key string) EngineConfig {
	var config, ok = di.EngineConfig(key)
	if !ok {
		log.Fatalf("engine %v config not found", key)
	}
	return config
}

package main

import (
	"fmt"

	xd_rsync "github.com/fabiofcferreira/xd-rsync"
	"github.com/fabiofcferreira/xd-rsync/logger"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		if _, isNotFoundError := err.(viper.ConfigFileNotFoundError); isNotFoundError {
			panic(fmt.Errorf("config file (config.json) was not found"))
		}

		panic(fmt.Errorf("unknown error: %w", err))
	}

	cfg, err := xd_rsync.GetConfig()
	if err != nil {
		panic(err)
	}

	logger, err := logger.CreateLogger(cfg.IsProductionMode, map[string]interface{}{
		"app_name": "xd_rsync",
	})
	if err != nil {
		panic(fmt.Errorf("logger error: %w", err))
	}

	logger.Info("startup_complete", "XD Rsync startup completed", nil)
}

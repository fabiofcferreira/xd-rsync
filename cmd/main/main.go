package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/fabiofcferreira/xd-rsync/database"
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

	cfg, err := GetConfig()
	if err != nil {
		panic(err)
	}

	logger, err := logger.CreateLogger(cfg.IsProductionMode, map[string]interface{}{
		"app_name": "xd_rsync",
	})
	if err != nil {
		panic(fmt.Errorf("logger error: %w", err))
	}

	dbService := &database.Service{}
	err = dbService.Init(&database.ServiceInitialisationInput{
		DSN:    cfg.DSN,
		Logger: logger,
	})
	if err != nil {
		os.Exit(1)
	}

	logger.Info("startup_complete", "XD Rsync startup completed", nil)

	// Wait for key press if close on finish is disabled
	if !cfg.CloseOnFinish {
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

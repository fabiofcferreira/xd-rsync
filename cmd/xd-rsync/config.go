package main

import (
	"fmt"
	"time"

	xd_rsync "github.com/fabiofcferreira/xd-rsync"
	"github.com/spf13/viper"
)

var ENVIRONMENTS = []string{"development", "staging", "production"}

func GetConfig() (*xd_rsync.Config, error) {
	cfg := &xd_rsync.Config{
		Queues: &xd_rsync.QueuesConfig{},
	}

	environment := viper.GetString("environment")
	if len(environment) == 0 {
		return nil, fmt.Errorf("environment not specified")
	}

	if environment != "development" && environment != "staging" && environment != "production" {
		return nil, fmt.Errorf("environment '%s' not supported", environment)
	}

	cfg.Environment = environment
	cfg.IsProductionMode = environment == "staging" || environment == "production"

	awsRegion := viper.GetString("awsRegion")
	if len(awsRegion) == 0 {
		fmt.Println("🫣 AWS region not specified. Defaulting to 'eu-west-2'")
	}
	cfg.AwsRegion = awsRegion

	dsn := viper.GetString("dsn")
	if len(dsn) == 0 {
		return nil, fmt.Errorf("database URI not specified")
	}
	cfg.DSN = dsn

	productUpdatesSnsArn := viper.GetString("queues.productUpdatesSnsQueueArn")
	if len(productUpdatesSnsArn) == 0 {
		fmt.Println("🫣 Product updates SNS queue ARN not specified.")
	}
	cfg.Queues.ProductUpdatesSnsQueueArn = productUpdatesSnsArn

	parsedSyncFrequency, err := time.ParseDuration(viper.GetString("syncFrequency"))
	if err == nil {
		cfg.SyncFrequency = parsedSyncFrequency
		fmt.Println("🫣 Sync frequency is invalid. Defaulting to 5 minutes")
	} else {
		cfg.SyncFrequency = 5 * time.Minute
	}

	fmt.Println("✅ Configuration validated!")
	fmt.Printf("⏳ Synchronisation configured to run every %s\n", cfg.SyncFrequency.String())

	return cfg, nil
}

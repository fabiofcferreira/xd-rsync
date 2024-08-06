package main

import (
	"fmt"
	"time"

	xd_rsync "github.com/fabiofcferreira/xd-rsync"
	"github.com/spf13/viper"
)

var ENVIRONMENTS = []string{"development", "staging", "production"}

func loadConfig() error {
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

	return nil
}

func GetConfig() (*xd_rsync.Config, error) {
	err := loadConfig()
	if err != nil {
		return nil, err
	}

	cfg := &xd_rsync.Config{
		Queues:        &xd_rsync.QueuesConfig{},
		DatadogConfig: &xd_rsync.DatadogConfig{},
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
	} else {
		cfg.SyncFrequency = 5 * time.Minute
		fmt.Println("🫣 Sync frequency is invalid. Defaulting to 5 minutes")
	}

	ingestHost := viper.GetString("datadog.ingestHost")
	if len(ingestHost) > 0 {
		cfg.DatadogConfig.IngestHost = &ingestHost
	}

	datadogApiKey := viper.GetString("datadog.apiKey")
	if len(datadogApiKey) > 0 {
		cfg.DatadogConfig.ApiKey = &datadogApiKey
	}

	parsedEventBaseFields := viper.GetStringMapString("datadog.eventBaseFields")
	cfg.DatadogConfig.EventBaseFields = &map[string]interface{}{
		"app_name":    "xd_rsync",
		"environment": cfg.Environment,
	}

	for key, value := range parsedEventBaseFields {
		(*cfg.DatadogConfig.EventBaseFields)[key] = value
	}

	fmt.Println("✅ Configuration validated!")
	fmt.Printf("⏳ Synchronisation configured to run every %s\n", cfg.SyncFrequency.String())

	return cfg, nil
}

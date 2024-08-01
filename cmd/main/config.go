package main

import (
	"fmt"

	"github.com/spf13/viper"
)

var ENVIRONMENTS = []string{"development", "staging", "production"}

type QueuesConfig struct {
	ProductUpdatesSnsQueueArn string `json:"productUpdatesSnsQueueArn,omitempty"`
}

type Config struct {
	Environment      string        `json:"environment"`
	IsProductionMode bool          `json:"isProductionMode"`
	AwsRegion        string        `json:"awsRegion"`
	DSN              string        `json:"dsn"`
	Queues           *QueuesConfig `json:"queues"`
	CloseOnFinish    bool          `json:"closeOnFinish"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{
		Queues: &QueuesConfig{},
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
		viper.Set("awsRegion", "eu-west-2")
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

	cfg.CloseOnFinish = viper.GetBool("closeOnFinish")

	fmt.Println("✅ Configuration validated!")
	return cfg, nil
}

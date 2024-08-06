package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func CreateProductionLogger(initialFields map[string]interface{}) (*Logger, error) {
	config := zap.NewProductionConfig()

	config.EncoderConfig.TimeKey = "timestamp"

	config.Development = false

	config.InitialFields = initialFields
	config.InitialFields["loggerMode"] = "production"

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	defer logger.Sync()

	return &Logger{
		instance: logger,
	}, nil
}

func CreateDevelopmentLogger(initialFields map[string]interface{}) (*Logger, error) {
	config := zap.NewDevelopmentConfig()

	config.EncoderConfig.TimeKey = "timestamp"

	config.Development = true

	config.InitialFields = initialFields
	config.InitialFields["loggerMode"] = "development"

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		instance: logger,
	}, nil
}

type LoggerOptions struct {
	IsProduction      bool
	InitialFields     map[string]interface{}
	DatadogApiKey     *string
	DatadogIngestHost *string
}

func CreateLogger(opts *LoggerOptions) (*Logger, error) {
	var logger *Logger
	var err error

	if opts.IsProduction {
		logger, err = CreateProductionLogger(opts.InitialFields)
	} else {
		logger, err = CreateDevelopmentLogger(opts.InitialFields)
	}

	if err != nil {
		return nil, fmt.Errorf("could not create logger: %w", err)
	}

	logger.eventBaseFields = opts.InitialFields

	isDatadogIngestHostValid := opts.DatadogIngestHost != nil && len(*opts.DatadogIngestHost) > 0
	isDatadogApiKeyValid := opts.DatadogApiKey != nil && len(*opts.DatadogApiKey) > 0

	if !isDatadogIngestHostValid && isDatadogApiKeyValid {
		fmt.Println("⚠️ Datadog ingest host was not provided. Datadog ingestion is disabled!")
	}

	if isDatadogIngestHostValid && !isDatadogApiKeyValid {
		fmt.Println("⚠️ Datadog API key was not provided. Datadog ingestion is disabled!")
	}

	if isDatadogIngestHostValid && isDatadogApiKeyValid {
		logger.datadogClient, err = createDatadogIngestClient(*opts.DatadogIngestHost, *opts.DatadogApiKey)
		if err != nil {
			return nil, err
		}
	}

	return logger, nil
}

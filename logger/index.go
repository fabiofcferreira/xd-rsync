package logger

import (
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
	IsProduction  bool
	InitialFields map[string]interface{}
	DatadogApiKey *string
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
		return nil, err
	}

	logger.eventBaseFields = opts.InitialFields

	if opts.DatadogApiKey != nil && len(*opts.DatadogApiKey) > 0 {
		logger.datadogClient, err = createDatadogIngestClient(*opts.DatadogApiKey)
		if err != nil {
			return nil, err
		}
	}

	return logger, nil
}

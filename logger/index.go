package logger

import (
	"go.uber.org/zap"
)

type Logger struct {
	instance *zap.Logger
}

func (txLogger *Logger) Debug(event_type string, message string, extraFields *map[string]interface{}) {
	fields := mapExtraFieldsToZap(map[string]interface{}{})
	if extraFields != nil {
		fields = mapExtraFieldsToZap(*extraFields)
	}

	txLogger.instance.Debug(message, *fields...)
}

func (txLogger *Logger) Info(event_type string, message string, extraFields *map[string]interface{}) {
	fields := mapExtraFieldsToZap(map[string]interface{}{})
	if extraFields != nil {
		fields = mapExtraFieldsToZap(*extraFields)
	}

	txLogger.instance.Info(message, *fields...)
}

func (txLogger *Logger) Warn(event_type string, message string, extraFields *map[string]interface{}) {
	fields := mapExtraFieldsToZap(map[string]interface{}{})
	if extraFields != nil {
		fields = mapExtraFieldsToZap(*extraFields)
	}

	txLogger.instance.Warn(message, *fields...)
}

func (txLogger *Logger) Error(event_type string, message string, extraFields *map[string]interface{}) {
	fields := mapExtraFieldsToZap(map[string]interface{}{})
	if extraFields != nil {
		fields = mapExtraFieldsToZap(*extraFields)
	}

	txLogger.instance.Error(message, *fields...)
}

func (txLogger *Logger) Fatal(event_type string, message string, extraFields *map[string]interface{}) {
	fields := mapExtraFieldsToZap(map[string]interface{}{})
	if extraFields != nil {
		fields = mapExtraFieldsToZap(*extraFields)
	}

	txLogger.instance.Fatal(message, *fields...)
}

func CreateProductionLogger(initialFields map[string]interface{}) (*Logger, error) {
	config := zap.NewProductionConfig()

	config.EncoderConfig.TimeKey = "timestamp"

	config.Development = false

	config.InitialFields = initialFields
	config.InitialFields["environment"] = "production"

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
	config.InitialFields["environment"] = "development"

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		instance: logger,
	}, nil
}

func CreateLogger(isProduction bool, initialFields map[string]interface{}) (*Logger, error) {
	if isProduction {
		return CreateProductionLogger(initialFields)
	}

	return CreateDevelopmentLogger(initialFields)
}

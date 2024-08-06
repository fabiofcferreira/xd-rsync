package logger

import (
	"maps"

	"go.uber.org/zap"
)

type Logger struct {
	instance        *zap.Logger
	datadogClient   *DatadogIngestClient
	eventBaseFields map[string]interface{}
}

func (txLogger *Logger) Debug(event_type string, message string, extraFields *map[string]interface{}) {
	eventPayload := map[string]interface{}{
		"event_type": event_type,
		"level":      "debug",
	}

	if extraFields != nil {
		maps.Copy(eventPayload, *extraFields)
	}

	maps.Copy(eventPayload, txLogger.eventBaseFields)

	txLogger.instance.Debug(message, *mapExtraFieldsToZap(eventPayload)...)

	if txLogger.datadogClient != nil {
		txLogger.datadogClient.SendEvent(eventPayload)
	}

}

func (txLogger *Logger) Info(event_type string, message string, extraFields *map[string]interface{}) {
	eventPayload := map[string]interface{}{
		"event_type": event_type,
		"level":      "info",
	}

	if extraFields != nil {
		maps.Copy(eventPayload, *extraFields)
	}

	maps.Copy(eventPayload, txLogger.eventBaseFields)

	txLogger.instance.Info(message, *mapExtraFieldsToZap(eventPayload)...)

	if txLogger.datadogClient != nil {
		txLogger.datadogClient.SendEvent(eventPayload)
	}
}

func (txLogger *Logger) Warn(event_type string, message string, extraFields *map[string]interface{}) {
	eventPayload := map[string]interface{}{
		"event_type": event_type,
		"level":      "warn",
	}

	if extraFields != nil {
		maps.Copy(eventPayload, *extraFields)
	}

	maps.Copy(eventPayload, txLogger.eventBaseFields)

	txLogger.instance.Warn(message, *mapExtraFieldsToZap(eventPayload)...)

	if txLogger.datadogClient != nil {
		txLogger.datadogClient.SendEvent(eventPayload)
	}
}

func (txLogger *Logger) Error(event_type string, message string, extraFields *map[string]interface{}) {
	eventPayload := map[string]interface{}{
		"event_type": event_type,
		"level":      "error",
	}

	if extraFields != nil {
		maps.Copy(eventPayload, *extraFields)
	}

	maps.Copy(eventPayload, txLogger.eventBaseFields)

	txLogger.instance.Error(message, *mapExtraFieldsToZap(eventPayload)...)

	if txLogger.datadogClient != nil {
		txLogger.datadogClient.SendEvent(eventPayload)
	}
}

func (txLogger *Logger) Fatal(event_type string, message string, extraFields *map[string]interface{}) {
	eventPayload := map[string]interface{}{
		"event_type": event_type,
		"level":      "fatal",
	}

	if extraFields != nil {
		maps.Copy(eventPayload, *extraFields)
	}

	maps.Copy(eventPayload, txLogger.eventBaseFields)

	txLogger.instance.Fatal(message, *mapExtraFieldsToZap(eventPayload)...)

	if txLogger.datadogClient != nil {
		txLogger.datadogClient.SendEvent(eventPayload)
	}
}

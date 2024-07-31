package logger

import "go.uber.org/zap"

func mapExtraFieldsToZap(extraFields map[string]interface{}) *[]zap.Field {
	fields := &[]zap.Field{}
	for key, value := range extraFields {
		*fields = append(*fields,
			zap.Any(key, value))
	}

	return fields
}

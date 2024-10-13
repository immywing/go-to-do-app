package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/uuid"
)

type contextKey string

const traceIDKey = contextKey("traceID")

func AddTraceID(ctx context.Context) context.Context {
	traceID := uuid.New().String()
	return context.WithValue(ctx, traceIDKey, traceID)
}

func GetTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(traceIDKey).(string)
	if !ok {
		return "unknown"
	}
	return traceID
}

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func LogWithTrace(ctx context.Context, logData map[string]interface{}, message string) {
	traceID := GetTraceID(ctx)
	logData["traceID"] = traceID
	var keyValues []interface{}
	for key, value := range logData {
		keyValues = append(keyValues, key, value)
	}
	logger.Info(message, keyValues...)
}

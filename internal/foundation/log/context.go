package log

import (
	"context"
	"os"

	"go.uber.org/zap"
)

type (
	ctxMarker struct{}

	ctxLogger struct {
		logger *Logger
	}
)

// nolint:gochecknoglobals
var ctxMarkerKey = &ctxMarker{}

// Extract takes the call-scoped Logger.
func Extract(ctx context.Context) *Logger {
	l, ok := ctx.Value(ctxMarkerKey).(*ctxLogger)
	if !ok || l == nil {
		// If there's no logger already initialized, then create a new default one
		return New(os.Stdout)
	}

	return l.logger
}

// ToContext adds the log.Logger to the context for extraction later.
// Returning the new context that has been created.
func ToContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, ctxMarkerKey, &ctxLogger{
		logger: logger,
	})
}

// AddFields adds fields to the context logger.
func AddFields(ctx context.Context, fields ...zap.Field) {
	Extract(ctx).WithFields(fields...)
}

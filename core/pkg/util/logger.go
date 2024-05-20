package util

import (
	"context"

	"github.com/sirupsen/logrus"
)

type loggerContextKey string

const loggerContextKeyVal = loggerContextKey("logrus.FieldLogger")

// Logger returns the appropriate logger for current context
func Logger(ctx context.Context) logrus.FieldLogger {
	if logger, ok := ctx.Value(loggerContextKeyVal).(logrus.FieldLogger); ok {
		return logger
	}
	return logrus.StandardLogger()
}

// WithLogger adds a value to the context for the logger
func WithLogger(ctx context.Context, logger logrus.FieldLogger) context.Context {
	return context.WithValue(ctx, loggerContextKeyVal, logger)
}

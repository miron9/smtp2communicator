package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxLogger struct{}

// ContextWithLogger puts logger into context
//
// This is convenience function inserting logger into context.
//
// Parameters:
//
// - ctx (context.Context): context
// - log (*zap.SugaredLogger): logger
//
// Returns:
//
// - newContext (context.Context): context with logger
func ContextWithLogger(ctx context.Context, log *zap.SugaredLogger) (newContext context.Context) {
	newContext = context.WithValue(ctx, ctxLogger{}, log)
	return
}

// LoggerFromContext gets logger from context
//
// This is a convenience function retrieved logger from context.
//
// Parameters:
//
// - ctx (context.Context): context
//
// Returns:
//
// - log (*zap.SugaredLogger): logger
func LoggerFromContext(ctx context.Context) (logger *zap.SugaredLogger) {
	var ok bool
	if logger, ok = ctx.Value(ctxLogger{}).(*zap.SugaredLogger); ok {
		return
	}
	return &zap.SugaredLogger{}
}

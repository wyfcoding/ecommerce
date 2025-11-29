package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"gorm.io/gorm/logger"

	"go.opentelemetry.io/otel/trace"
)

var (
	defaultLogger *Logger
	once          sync.Once
)

type Logger struct {
	*slog.Logger
	Service string
	Module  string
}

// TraceHandler adds trace_id and span_id from context
type TraceHandler struct {
	slog.Handler
}

func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

// NewLogger creates a new slog Logger
func NewLogger(service, module string) *Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
			}
			return a
		},
	}

	// Use JSON handler for structured logging
	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)
	traceHandler := &TraceHandler{Handler: jsonHandler}

	logger := slog.New(traceHandler).With(
		slog.String("service", service),
		slog.String("module", module),
	)

	return &Logger{
		Logger:  logger,
		Service: service,
		Module:  module,
	}
}

// InitLogger initializes the default global logger
func InitLogger(service, module string) {
	once.Do(func() {
		defaultLogger = NewLogger(service, module)
		slog.SetDefault(defaultLogger.Logger)
	})
}

// GetLogger returns the default logger
func GetLogger() *Logger {
	if defaultLogger == nil {
		return NewLogger("unknown", "unknown")
	}
	return defaultLogger
}

// GormLogger is a custom GORM logger that uses slog
type GormLogger struct {
	logger        *slog.Logger
	SlowThreshold time.Duration
}

// NewGormLogger creates a new GormLogger
func NewGormLogger(logger *Logger, slowThreshold time.Duration) *GormLogger {
	return &GormLogger{
		logger:        logger.Logger,
		SlowThreshold: slowThreshold,
	}
}

// LogMode implements gorm logger.Interface
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	// Slog doesn't have a direct equivalent to changing log level on the fly for a specific logger instance
	// without creating a new one, but for GORM we can just return the same logger
	// or wrap it if we want to filter. For now, we return the same logger.
	return l
}

// Info implements gorm logger.Interface
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.InfoContext(ctx, fmt.Sprintf(msg, data...))
}

// Warn implements gorm logger.Interface
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.WarnContext(ctx, fmt.Sprintf(msg, data...))
}

// Error implements gorm logger.Interface
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.ErrorContext(ctx, fmt.Sprintf(msg, data...))
}

// Trace implements gorm logger.Interface
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []any{
		slog.String("sql", sql),
		slog.Duration("elapsed", elapsed),
	}
	if rows != -1 {
		fields = append(fields, slog.Int64("rows", rows))
	}

	if err != nil && err != logger.ErrRecordNotFound {
		fields = append(fields, slog.Any("error", err))
		l.logger.ErrorContext(ctx, "gorm trace error", fields...)
	} else if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		fields = append(fields, slog.String("type", "slow_query"))
		l.logger.WarnContext(ctx, "gorm trace slow query", fields...)
	} else {
		l.logger.DebugContext(ctx, "gorm trace", fields...)
	}
}

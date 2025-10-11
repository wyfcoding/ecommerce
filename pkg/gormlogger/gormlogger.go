package gormlogger

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// GormLogger is a custom GORM logger that integrates with Zap.
type GormLogger struct {
	zapLogger *zap.Logger
	LogLevel  logger.LogLevel
	SlowThreshold time.Duration
}

// New creates a new GormLogger instance.
func New(zapLogger *zap.Logger, logLevel logger.LogLevel, slowThreshold time.Duration) *GormLogger {
	return &GormLogger{
		zapLogger: zapLogger,
		LogLevel:  logLevel,
		SlowThreshold: slowThreshold,
	}
}

// LogMode log mode
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info prints info log
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.zapLogger.Sugar().Infof(msg, data...)
	}
}

// Warn prints warn log
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.zapLogger.Sugar().Warnf(msg, data...)
	}
}

// Error prints error log
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.zapLogger.Sugar().Errorf(msg, data...)
	}
}

// Trace prints trace log
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Duration("elapsed", elapsed),
	}
	if rows == -1 {
		fields = append(fields, zap.String("rows", "-"))
	} else {
		fields = append(fields, zap.Int64("rows", rows))
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.zapLogger.Error("gorm trace error", fields...)
	} else if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		fields = append(fields, zap.String("type", "slow_query"))
		l.zapLogger.Warn("gorm trace slow query", fields...)
	} else {
		l.zapLogger.Debug("gorm trace", fields...)
	}
}

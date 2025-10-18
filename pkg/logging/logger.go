package logging

import (
	"context"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
)

// NewLogger 根据提供的配置创建一个新的 zap Logger 实例。
// level: 日志级别 (例如, "debug", "info", "warn", "error")
// format: 日志格式 ("json" 或 "console")
// output: 日志输出路径 (例如, "stdout", "stderr", 或一个文件路径)
func NewLogger(level, format, output string) *zap.Logger {
	// 解析日志级别字符串
	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		logLevel = zapcore.InfoLevel // 如果解析失败，默认为 Info 级别
	}

	// 配置日志编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // 时间格式: 2006-01-02T15:04:05.000Z0700
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 在终端中为日志级别添加颜色

	// 根据 format 参数选择编码器 (json 或 console)
	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 设置日志输出位置
	writeSyncer, _, err := zap.Open(output)
	if err != nil {
		// 如果打开指定路径失败，则默认输出到标准错误
		writeSyncer = zapcore.Lock(os.Stderr)
	}

	// 创建核心 logger
	core := zapcore.NewCore(encoder, writeSyncer, logLevel)

	// 创建最终的 logger，并添加调用者信息
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return logger
}

// GinLogger 返回一个 Gin 框架的中间件，用于记录每个HTTP请求的详细信息。
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 请求处理完成，记录日志
		cost := time.Since(start)
		logger.Info(path, // 使用请求路径作为日志消息
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// For 返回一个包含了 OpenTelemetry Trace ID 和 Span ID 的子 logger。
// 这对于在微服务架构中追踪完整的请求链路至关重要。
func For(ctx context.Context, logger *zap.Logger) *zap.Logger {
	// 检查 context 中是否存在有效的 span
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		// 如果有，则在日志中自动附加 trace_id 和 span_id 字段
		return logger.With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	// 如果没有 trace 信息，返回原始 logger
	return logger
}

// GormLogger is a custom GORM logger that integrates with Zap.
type GormLogger struct {
	zapLogger     *zap.Logger
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

// New creates a new GormLogger instance.
func NewGormLogger(zapLogger *zap.Logger, logLevel logger.LogLevel, slowThreshold time.Duration) *GormLogger {
	return &GormLogger{
		zapLogger:     zapLogger,
		LogLevel:      logLevel,
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

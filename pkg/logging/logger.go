package logging

import (
	"context"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger 根据提供的级别、格式和输出路径创建一个新的 zap Logger。
func NewLogger(level, format, output string) *zap.Logger {
	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		logLevel = zapcore.InfoLevel // 默认 Info 级别
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 彩色日志

	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 支持多输出，例如 stdout 和文件
	var cores []zapcore.Core

	// Console/File output
	writeSyncer, _, err := zap.Open(output)
	if err != nil {
		writeSyncer = zapcore.Lock(os.Stderr)
	}
	cores = append(cores, zapcore.NewCore(encoder, writeSyncer, logLevel))

	// Optional: Remote logging (simulating Logstash/Elasticsearch)
	// In a real system, this would be more sophisticated (e.g., using a dedicated client, buffering, retries).
	// For demonstration, we'll just print to stderr if remote logging is enabled.
	// A real remote logger would use a network connection (TCP/HTTP).
	if os.Getenv("REMOTE_LOGGING_ENABLED") == "true" {
		remoteEncoderConfig := zap.NewProductionEncoderConfig()
		remoteEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		remoteEncoderConfig.EncodeLevel = zapcore.CapitalEncoder // No color for remote
		remoteEncoder := zapcore.NewJSONEncoder(remoteEncoderConfig) // JSON for remote

		// Simulate sending to a remote endpoint (e.g., Logstash TCP input)
		// In a real scenario, you'd replace this with actual network client.
		remoteWriter := zapcore.AddSync(os.Stderr) // For demonstration, still stderr
		cores = append(cores, zapcore.NewCore(remoteEncoder, remoteWriter, logLevel))
		zap.S().Info("Remote logging enabled (simulated to stderr).")
	}

	logger := zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1))
	return logger
}

// GinLogger 返回一个 Gin 中间件，用于记录请求日志。
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		cost := time.Since(start)
		logger.Info(path,
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

// For 返回一个带有 trace_id 和 span_id 的 logger。
func For(ctx context.Context, logger *zap.Logger) *zap.Logger {
	// 检查 context 中是否存在有效的 span
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		// 如果有，则在日志中自动附加 trace_id 和 span_id
		return logger.With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return logger
}

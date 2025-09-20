package logging

import (
	"context"

	"go.opentelemetry.io/otel/trace" // 引入 trace 包
	"go.uber.org/zap"
)

var globalLogger *zap.Logger

// InitLogger 初始化全局的 zap Logger
func InitLogger() {
	// 使用 zap 推荐的生产环境配置
	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize zap logger: " + err.Error())
	}
	// defer logger.Sync() // 在 main 函数退出前调用
	globalLogger = logger
}

// For a simple global logger approach. In more complex systems,
// you might pass the logger instance via dependency injection.
func For(ctx context.Context) *zap.Logger {
	// 检查 context 中是否存在有效的 span
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		// 如果有，则在日志中自动附加 trace_id 和 span_id
		return globalLogger.With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return globalLogger
}

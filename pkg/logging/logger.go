// Package logging 提供了统一的结构化日志（slog）封装，支持OpenTelemetry追踪上下文注入和GORM日志集成。
package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"gorm.io/gorm/logger" // GORM的日志接口

	"go.opentelemetry.io/otel/trace" // OpenTelemetry追踪
)

var (
	// defaultLogger 是全局默认的Logger实例，采用单例模式。
	defaultLogger *Logger
	// once 用于确保InitLogger函数只被执行一次，保证defaultLogger的单例性。
	once sync.Once
)

// Logger 结构体封装了原生的 `*slog.Logger`，并添加了服务名和模块名，方便在日志中区分来源。
type Logger struct {
	*slog.Logger
	Service string // 服务名称
	Module  string // 模块名称
}

// TraceHandler 是一个自定义的 `slog.Handler` 装饰器，用于从 `context.Context` 中提取并注入 `trace_id` 和 `span_id` 到日志记录中。
type TraceHandler struct {
	slog.Handler
}

// Handle 方法实现了 `slog.Handler` 接口，在处理日志记录之前，
// 会尝试从上下文获取OpenTelemetry的SpanContext，如果有效，则将trace_id和span_id添加到日志属性中。
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() { // 检查SpanContext是否有效，即是否存在正在进行的追踪
		r.AddAttrs(
			slog.String("trace_id", spanCtx.TraceID().String()), // 注入追踪ID
			slog.String("span_id", spanCtx.SpanID().String()),   // 注入Span ID
		)
	}
	return h.Handler.Handle(ctx, r) // 调用被装饰的原始Handler继续处理日志
}

// NewLogger 创建一个新的Logger实例。
// service: 日志所属的服务名称。
// module: 日志所属的模块名称。
// 返回一个配置好的Logger实例，其日志输出格式为JSON，并默认包含服务名和模块名。
func NewLogger(service, module string) *Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo, // 默认日志级别为Info
		// ReplaceAttr 允许修改或替换日志属性。这里将默认的 "time" 键改为 "timestamp"。
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
			}
			return a
		},
	}

	// 使用JSON格式处理器将日志输出到标准输出。
	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)
	// 装饰JSON处理器，使其能够注入追踪ID。
	traceHandler := &TraceHandler{Handler: jsonHandler}

	// 创建一个slog.Logger实例，并设置默认属性，如服务名和模块名。
	logger := slog.New(traceHandler).With(
		slog.String("service", service),
		slog.String("module", module),
		// slog.String("level", ""), // 移除默认的level字段以避免重复或自定义显示
	)

	return &Logger{
		Logger:  logger,
		Service: service,
		Module:  module,
	}
}

// InitLogger 初始化全局默认的Logger。
// 此函数应在应用程序启动时调用一次，以配置全局日志行为。
func InitLogger(service, module string) {
	once.Do(func() { // 确保只初始化一次
		defaultLogger = NewLogger(service, module)
		slog.SetDefault(defaultLogger.Logger) // 设置slog的默认Logger
	})
}

// GetLogger 返回全局默认的Logger实例。
// 如果尚未通过InitLogger初始化，它会返回一个带有"unknown"服务和模块的默认Logger。
func GetLogger() *Logger {
	if defaultLogger == nil {
		return NewLogger("unknown", "unknown")
	}
	return defaultLogger
}

// GormLogger 是一个自定义的GORM日志器，它实现了 `gorm.io/gorm/logger.Interface` 接口，
// 从而允许GORM将数据库操作日志输出到统一的slog日志系统中。
type GormLogger struct {
	logger        *slog.Logger  // 用于输出日志的slog实例
	SlowThreshold time.Duration // 慢查询阈值，超过此阈值的SQL查询将被记录为警告
}

// NewGormLogger 创建一个新的GormLogger实例。
// logger: 使用的Logger实例。
// slowThreshold: 慢查询的持续时间阈值。
func NewGormLogger(logger *Logger, slowThreshold time.Duration) *GormLogger {
	return &GormLogger{
		logger:        logger.Logger,
		SlowThreshold: slowThreshold,
	}
}

// LogMode 实现了gorm logger.Interface的LogMode方法。
// 鉴于slog的设计，通常不直接在运行时动态更改现有logger实例的级别，
// 而是通过创建新的Handler或在HandlerOptions中配置。
// 此处直接返回自身，表示沿用当前logger的配置。
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	// GORM的LogLevel与slog的LogLevel映射关系需要自行定义，或者在此处根据level做过滤。
	// 目前简单返回自身，意味着GORM的日志级别控制将主要依赖于NewLogger中配置的slog级别。
	return l
}

// Info 实现了gorm logger.Interface的Info方法，将GORM的Info级别日志输出为slog的Info级别。
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.InfoContext(ctx, fmt.Sprintf(msg, data...))
}

// Warn 实现了gorm logger.Interface的Warn方法，将GORM的Warn级别日志输出为slog的Warn级别。
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.WarnContext(ctx, fmt.Sprintf(msg, data...))
}

// Error 实现了gorm logger.Interface的Error方法，将GORM的Error级别日志输出为slog的Error级别。
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.ErrorContext(ctx, fmt.Sprintf(msg, data...))
}

// Trace 实现了gorm logger.Interface的Trace方法，用于记录SQL查询的详细信息，包括耗时、SQL语句和错误。
// 慢查询会以Warn级别记录，错误查询以Error级别记录，普通查询以Debug级别记录。
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin) // 计算SQL执行耗时
	sql, rows := fc()            // 获取SQL语句和影响的行数

	fields := []any{
		slog.String("sql", sql),           // SQL语句
		slog.Duration("elapsed", elapsed), // 执行耗时
	}
	if rows != -1 { // 检查是否返回了影响行数
		fields = append(fields, slog.Int64("rows", rows)) // 影响行数
	}

	if err != nil && err != logger.ErrRecordNotFound { // 如果存在错误且不是“未找到记录”错误
		fields = append(fields, slog.Any("error", err)) // 错误信息
		l.logger.ErrorContext(ctx, "gorm trace error", fields...)
	} else if l.SlowThreshold != 0 && elapsed > l.SlowThreshold { // 如果是慢查询
		fields = append(fields, slog.String("type", "slow_query"))
		l.logger.WarnContext(ctx, "gorm trace slow query", fields...)
	} else { // 普通查询
		l.logger.DebugContext(ctx, "gorm trace", fields...)
	}
}

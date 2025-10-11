package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
)

// Config 结构体用于 Jaeger Tracer 配置。
type Config struct {
	ServiceName    string `toml:"service_name"`
	JaegerEndpoint string `toml:"jaeger_endpoint"`
}

// InitTracer 初始化并注册 Jaeger Tracer Provider
func InitTracer(conf *Config) (*sdktrace.TracerProvider, func(), error) {
	// 创建 Jaeger exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(conf.JaegerEndpoint)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create jaeger exporter: %w", err)
	}

	// 创建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		// 始终对 span 进行采样
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		// 使用批量处理器，性能更好
		sdktrace.WithBatcher(exporter),
		// 设置服务名等资源属性
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(conf.ServiceName),
		)),
	)

	// 注册为全局 TracerProvider
	otel.SetTracerProvider(tp)

	// 注册 W3C trace context 和 baggage propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			zap.S().Errorf("failed to shutdown TracerProvider: %v", err)
		}
	}

	return tp, cleanup, nil
}

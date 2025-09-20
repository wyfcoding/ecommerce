package tracing

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// InitTracer 初始化并注册 Jaeger Tracer Provider
func InitTracer(serviceName, jaegerEndpoint string) (*sdktrace.TracerProvider, error) {
	// 创建 Jaeger exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		return nil, err
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
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	// 注册为全局 TracerProvider
	otel.SetTracerProvider(tp)

	// 注册 W3C trace context 和 baggage propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

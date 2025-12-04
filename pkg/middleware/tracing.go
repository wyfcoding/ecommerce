package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// TracingMiddleware 创建一个链路追踪中间件。
// 该中间件负责从HTTP请求头中提取追踪上下文，并为当前请求创建一个新的OpenTelemetry Span，
// 从而实现分布式追踪。
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从HTTP请求头中提取OpenTelemetry追踪上下文。
		// propagation.HeaderCarrier 包装了 c.Request.Header，使其符合TextMapCarrier接口。
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// 获取一个OpenTelemetry Tracer实例，用于创建Span。
		tracer := otel.Tracer(serviceName)
		// 为当前请求启动一个新的Span。Span的名称通常是HTTP方法和请求路径。
		ctx, span := tracer.Start(ctx, c.Request.Method+" "+c.FullPath())
		defer span.End() // 确保在请求处理结束时结束Span。

		// 将包含新Span的上下文注入到Gin的请求上下文中。
		// 这样，后续的业务逻辑就可以通过 c.Request.Context() 获取到这个Span，并继续追踪。
		c.Request = c.Request.WithContext(ctx)

		// 调用请求链中的下一个处理程序。
		c.Next()

		// 在这里可以根据需要记录HTTP状态码或其他请求处理结果到Span中。
		// 例如：span.SetStatus(codes.Code(c.Writer.Status()), "")
		// 此处注释，遵循不修改原逻辑的原则。
	}
}

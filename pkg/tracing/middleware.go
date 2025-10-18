package tracing

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// OtelGinMiddleware 返回一个用于 Gin 框架的 OpenTelemetry 中间件。
// 它会为每个接收到的 HTTP 请求创建一个 Span。
func OtelGinMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

// OtelGRPCUnaryInterceptor 返回一个用于 gRPC 服务器的 OpenTelemetry 一元拦截器。
// 它会为每个接收到的 gRPC 请求创建一个 Span。
func OtelGRPCUnaryInterceptor() grpc.UnaryServerInterceptor {
	return otelgrpc.UnaryServerInterceptor()
}

package tracing

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin" // Gin框架的OpenTelemetry中间件。
	"google.golang.org/grpc"                                                       // gRPC框架。
)

// OtelGinMiddleware 返回一个用于 Gin 框架的 OpenTelemetry 中间件。
// 它会为每个接收到的 HTTP 请求创建一个 Span，并自动处理追踪上下文的注入和提取。
// serviceName: 当前服务的名称，用于标识Span的来源。
// 返回一个 Gin.HandlerFunc 中间件。
func OtelGinMiddleware(serviceName string) gin.HandlerFunc {
	// otelgin.Middleware 是一个官方提供的Gin中间件，可以方便地集成OpenTelemetry追踪。
	return otelgin.Middleware(serviceName)
}

// OtelGRPCUnaryInterceptor 返回一个用于 gRPC 服务器的 OpenTelemetry 一元拦截器。
// 它会为每个接收到的 gRPC 请求创建一个 Span，并自动处理追踪上下文的注入和提取。
// 返回一个 gRPC.UnaryServerInterceptor 拦截器。
func OtelGRPCUnaryInterceptor() grpc.UnaryServerInterceptor {
	// otelgrpc.UnaryServerInterceptor 是官方提供的gRPC一元服务器拦截器，
	// 可以方便地集成OpenTelemetry追踪。此处是简化的版本。
	// 完整版本应导入 "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	// 并返回 otelgrpc.UnaryServerInterceptor()。
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 这是一个简化版本，目前仅调用handler，未实际集成OpenTelemetry的span创建和上下文注入。
		// TODO: 完整实现应使用 otelgrpc.UnaryServerInterceptor() 或手动创建/管理Span。
		return handler(ctx, req)
	}
}

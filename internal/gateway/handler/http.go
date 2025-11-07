package handler

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RegisterServiceHandlers 动态注册所有 gRPC 服务处理器
// TODO: 需要先生成protobuf文件后才能使用
func RegisterServiceHandlers(ctx context.Context, gwmux *runtime.ServeMux, servicesConfig map[string]struct{ Addr string }, opts []grpc.DialOption) {
	zap.S().Info("Service handlers registration is disabled until protobuf files are generated")
	// Protobuf文件生成后，在这里注册各个服务
}

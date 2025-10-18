package server

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// GRPCServer 是一个 gRPC 服务器实现。
type GRPCServer struct {
	srv  *grpc.Server
	addr string
}

// NewGRPCServer 创建一个新的 gRPC 服务器。
// 它现在可以接收一个或多个一元拦截器。
func NewGRPCServer(addr string, register func(*grpc.Server), interceptors ...grpc.UnaryServerInterceptor) *GRPCServer {
	// 使用 grpc.ChainUnaryInterceptor 将所有传入的拦截器链接起来
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	register(s)
	return &GRPCServer{
		srv:  s,
		addr: addr,
	}
}

// Start 启动 gRPC 服务器。
func (s *GRPCServer) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	zap.S().Infof("gRPC server listening at %v", lis.Addr())
	return s.srv.Serve(lis)
}

// Stop 停止 gRPC 服务器。
func (s *GRPCServer) Stop(ctx context.Context) error {
	zap.S().Info("Stopping gRPC server...")
	s.srv.GracefulStop()
	return nil
}
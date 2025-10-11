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
func NewGRPCServer(addr string, register func(*grpc.Server)) *GRPCServer {
	s := grpc.NewServer()
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
		return fmt.Errorf("failed to listen: %w", err) // 错误信息保持英文
	}
	zap.S().Infof("gRPC server listening at %v", lis.Addr()) // 日志信息保持英文
	return s.srv.Serve(lis)
}

// Stop 停止 gRPC 服务器。
func (s *GRPCServer) Stop(ctx context.Context) error {
	zap.S().Info("Stopping gRPC server...") // 日志信息保持英文
	s.srv.GracefulStop()
	return nil
}

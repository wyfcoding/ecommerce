package handler

import (
	"fmt"
	"net"

	v1 "ecommerce/api/admin/v1"
	"ecommerce/internal/admin/service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// StartGRPCServer 启动 gRPC 服务器。
// 它监听指定的地址和端口，并注册 AdminServiceServer。
func StartGRPCServer(svc *service.AdminService, addr string, port int) (*grpc.Server, chan error) {
	errChan := make(chan error, 1)
	// 监听 TCP 地址
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %w", err)
		return nil, errChan
	}
	// 创建新的 gRPC 服务器
	s := grpc.NewServer()
	// 注册 AdminServiceServer
	v1.RegisterAdminServiceServer(s, svc)

	zap.S().Infof("gRPC server listening at %v", lis.Addr())
	// 在 goroutine 中启动 gRPC 服务器，以便不阻塞主线程
	go func() {
		if err := s.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(errChan)
	}()
	return s, errChan
}

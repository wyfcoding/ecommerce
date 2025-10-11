package adminhandler

import (
	"fmt"
	"net"

	v1 "ecommerce/api/admin/v1"
	"ecommerce/internal/admin/service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// startGRPCServer 启动 gRPC 服务器
func StartGRPCServer(adminService *service.AdminService, authInterceptor *service.AuthInterceptor, addr string, port int) (*grpc.Server, chan error) {
	errChan := make(chan error, 1)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %w", err)
		return nil, errChan
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Auth),
	)
	v1.RegisterAdminServer(s, adminService)

	zap.S().Infof("gRPC server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(errChan)
	}()
	return s, errChan
}

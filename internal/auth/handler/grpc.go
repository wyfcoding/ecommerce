package authhandler

import (
	"fmt"
	"net"

	v1 "ecommerce/api/auth/v1"
	"ecommerce/internal/auth/service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// StartGRPCServer 启动 gRPC 服务器
func StartGRPCServer(svc *service.AuthService, addr string, port int) (*grpc.Server, chan error) {
	errChan := make(chan error, 1)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %w", err)
		return nil, errChan
	}
	s := grpc.NewServer()
	v1.RegisterAuthServiceServer(s, svc)

	zap.S().Infof("gRPC server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(errChan)
	}()
	return s, errChan
}

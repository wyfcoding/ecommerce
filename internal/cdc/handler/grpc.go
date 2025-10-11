package cdchandler

import (
	"fmt"
	"net"

	v1 "ecommerce/api/cdc/v1"
	"ecommerce/internal/cdc/service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// StartGRPCServer 启动 gRPC 服务器
func StartGRPCServer(svc *service.CdcService, addr string, port int) (*grpc.Server, chan error) {
	errChan := make(chan error, 1)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %w", err)
		return nil, errChan
	}
	s := grpc.NewServer()
	v1.RegisterCdcServiceServer(s, svc)

	zap.S().Infof("gRPC server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(errChan)
	}()
	return s, errChan
}

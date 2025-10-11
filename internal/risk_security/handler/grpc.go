package risksecurityhandler

import (
	"fmt"
	"net"

	v1 "ecommerce/api/risk_security/v1"
	"ecommerce/internal/risk_security/service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// StartGRPCServer starts the gRPC server.
func StartGRPCServer(svc *service.RiskSecurityService, addr string, port int) (*grpc.Server, chan error) {
	errChan := make(chan error, 1)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %w", err)
		return nil, errChan
	}
	s := grpc.NewServer()
	v1.RegisterRiskSecurityServiceServer(s, svc)

	zap.S().Infof("gRPC server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(errChan)
	}()
	return s, errChan
}

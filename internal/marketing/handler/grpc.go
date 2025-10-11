package marketinghandler

import (
	"fmt"
	"net"

	v1 "ecommerce/api/marketing/v1"
	"ecommerce/internal/marketing/service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// StartGRPCServer starts the gRPC server.
func StartGRPCServer(marketingService *service.MarketingService, addr string, port int) (*grpc.Server, chan error) {
	errChan := make(chan error, 1)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %w", err)
		return nil, errChan
	}
	s := grpc.NewServer()
	v1.RegisterMarketingServer(s, marketingService)

	zap.S().Infof("gRPC server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(errChan)
	}()
	return s, errChan
}

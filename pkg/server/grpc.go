package server

import (
	"context"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	server *grpc.Server
	addr   string
	logger *slog.Logger
}

func NewGRPCServer(addr string, logger *slog.Logger, register func(*grpc.Server), interceptors ...grpc.UnaryServerInterceptor) *GRPCServer {
	opts := []grpc.ServerOption{}
	if len(interceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(interceptors...))
	}

	s := grpc.NewServer(opts...)
	register(s)
	reflection.Register(s)

	return &GRPCServer{
		server: s,
		addr:   addr,
		logger: logger,
	}
}

func (s *GRPCServer) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.logger.Info("Starting gRPC server", "addr", s.addr)

	errChan := make(chan error, 1)
	go func() {
		errChan <- s.server.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func (s *GRPCServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping gRPC server")
	s.server.GracefulStop()
	return nil
}

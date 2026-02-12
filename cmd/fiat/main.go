package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/wyfcoding/ecommerce/go-api/fiat/v1"
	"github.com/wyfcoding/ecommerce/internal/fiat/application"
	"github.com/wyfcoding/ecommerce/internal/fiat/domain"
	"github.com/wyfcoding/ecommerce/internal/fiat/interfaces"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. 初始化日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("starting fiat service")

	// 2. 依赖注入
	fiatService := &domain.FiatService{} // 假设 domain/fiat_service.go 已存在
	appService := application.NewFiatApplicationService(fiatService, logger)
	handler := interfaces.NewFiatHandler(appService)

	// 3. 启动 gRPC 服务
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50061"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFiatServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	// 4. 优雅关停
	go func() {
		logger.Info("gRPC server listening on", "port", port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down fiat server...")
	grpcServer.GracefulStop()
	logger.Info("server exited")
}

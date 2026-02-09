package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/wyfcoding/ecommerce/goapi/orchestrator/v1"
	"github.com/wyfcoding/ecommerce/internal/orchestrator/application"
	"github.com/wyfcoding/ecommerce/internal/orchestrator/infrastructure"
	"github.com/wyfcoding/ecommerce/internal/orchestrator/interfaces"
	"github.com/wyfcoding/pkg/saga"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. 初始化日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("starting orchestrator service")

	// 2. 初始化数据库
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:root1234@tcp(127.0.0.1:3306)/ecommerce?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	// 3. 依赖注入
	repo := infrastructure.NewOrchestratorRepository(db)
	engine := saga.NewEngine()
	appService := application.NewOrchestratorApplicationService(repo, engine, logger)
	handler := interfaces.NewOrchestratorHandler(appService)

	// 4. 启动 gRPC 服务
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50062"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrchestratorServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	// 5. 优雅关停
	go func() {
		logger.Info("gRPC server listening on", "port", port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down orchestrator server...")
	grpcServer.GracefulStop()
	logger.Info("server exited")
}

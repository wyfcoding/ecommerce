package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/wyfcoding/ecommerce/goapi/kyc"
	"github.com/wyfcoding/ecommerce/internal/kyc/application"
	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
	"github.com/wyfcoding/ecommerce/internal/kyc/infrastructure"
	"github.com/wyfcoding/ecommerce/internal/kyc/interfaces"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. 初始化日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("starting KYC service")

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

	// 自动迁移
	db.AutoMigrate(&domain.KYCApplication{})

	// 3. 依赖注入
	repo := infrastructure.NewKYCRepository(db)
	appService := application.NewKYCApplicationService(repo, logger)
	handler := interfaces.NewKYCHandler(appService)

	// 4. 启动 gRPC 服务
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50063"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterKYCServiceServer(grpcServer, handler)
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

	logger.Info("shutting down kyc server...")
	grpcServer.GracefulStop()
	logger.Info("server exited")
}

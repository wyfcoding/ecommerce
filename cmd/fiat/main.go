package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	pb "github.com/wyfcoding/ecommerce/go-api/fiat/v1"
	"github.com/wyfcoding/ecommerce/internal/fiat/application"
	"github.com/wyfcoding/ecommerce/internal/fiat/domain"
	"github.com/wyfcoding/ecommerce/internal/fiat/infrastructure"
	"github.com/wyfcoding/ecommerce/internal/fiat/interfaces"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("starting fiat service")

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(127.0.0.1:3306)/ecommerce_fiat?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	txRepo := infrastructure.NewGormFiatTransactionRepository(db)
	rateRepo := infrastructure.NewGormExchangeRateRepository(db)
	bankAccountRepo := infrastructure.NewGormBankAccountRepository(db)
	channelRepo := infrastructure.NewGormFiatChannelRepository(db)

	fiatService := domain.NewFiatService(txRepo, rateRepo, channelRepo, nil)

	appService := application.NewFiatApplicationService(
		txRepo,
		rateRepo,
		bankAccountRepo,
		channelRepo,
		fiatService,
		nil,
		logger,
	)
	handler := interfaces.NewFiatHandler(appService)

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

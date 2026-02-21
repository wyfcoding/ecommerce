// 变更说明：
// 促销计价大引擎启动器。涵盖双库 (Redis 限流锁 + MySQL 落盘)。
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/wyfcoding/ecommerce/internal/promotion/application"
	"github.com/wyfcoding/ecommerce/internal/promotion/infrastructure"
	promogrpc "github.com/wyfcoding/ecommerce/internal/promotion/interfaces/grpc"
	promohttp "github.com/wyfcoding/ecommerce/internal/promotion/interfaces/http"

	"github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/redis"
	"github.com/wyfcoding/pkg/server"
	"github.com/wyfcoding/pkg/tracing"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. 初始化平台配置
	var cfg config.Config
	if err := config.Load("configs/ecommerce/promotion/config.toml", &cfg); err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 2. Logging & Tracing 全埋点集成
	logger := logging.NewLogger(cfg.Server.Name, "main", cfg.Log.Level)
	slog.SetDefault(logger.Logger)
	if traceShutdown, err := tracing.InitTracer(cfg.Tracing); err == nil {
		defer traceShutdown(ctx)
	}

	// 3. 仓储初始化 (MySQL CQRS 两端 + Redis Lua 超发锁)
	db, err := database.NewDB(cfg.Data.Database, cfg.CircuitBreaker, logger, nil)
	if err != nil {
		logger.Error("DB connect failed", "err", err)
		os.Exit(1)
	}

	rdb, _, err := redis.NewClient(&cfg.Data.Redis, logger)
	if err != nil {
		logger.Error("Redis config failed", "err", err)
		os.Exit(1)
	}

	writeRepo, readRepo := infrastructure.NewPromotionRepository(db.RawDB())
	cacheRepo, err := infrastructure.NewPromotionCache(rdb)
	if err != nil {
		logger.Error("Failed to init Lua Cache script", "err", err)
	}

	producer := kafka.NewProducer(&cfg.MessageQueue.Kafka, logger, nil)
	publisher := kafka.NewEventPublisher(producer)
	defer producer.Close()

	// 4. CQRS Application
	cmdService := application.NewPromotionCommandService(writeRepo, publisher, logger.Logger)
	qryService := application.NewPromotionQueryService(readRepo, cacheRepo, logger.Logger)

	// 5. 接口挂载
	ginEngine := server.NewDefaultGinEngine()
	httpHandler := promohttp.NewPromotionHandler(cmdService, qryService, logger.Logger)
	httpHandler.RegisterRoutes(ginEngine)
	httpServer := server.NewGinServer(ginEngine, fmt.Sprintf(":%d", cfg.Server.HTTP.Port), logger.Logger)

	grpcServer := grpc.NewServer()
	// pb.RegisterPromotionServiceServer(grpcServer, promogrpc.NewPromotionGrpcServer(cmdService, logger))
	_ = promogrpc.NewPromotionGrpcServer(cmdService, logger.Logger)

	// 6. 并发控制与优雅停机
	go func() {
		logger.Info("Starting Promotion App HTTP", "port", cfg.Server.HTTP.Port)
		_ = httpServer.Start(ctx)
	}()
	go func() {
		lis, _ := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPC.Port))
		logger.Info("Starting Promotion App gRPC", "port", cfg.Server.GRPC.Port)
		_ = grpcServer.Serve(lis)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Halting Promotion app...")
	grpcServer.GracefulStop()
	_ = httpServer.Stop(ctx)
	logger.Info("Promotion App halted completely.")
}

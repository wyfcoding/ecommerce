package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/user/application"
	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/ecommerce/internal/user/infrastructure/messaging"
	"github.com/wyfcoding/ecommerce/internal/user/infrastructure/persistence"
	grpcInterface "github.com/wyfcoding/ecommerce/internal/user/interfaces/grpc"
	httpInterface "github.com/wyfcoding/ecommerce/internal/user/interfaces/http"
	"github.com/wyfcoding/pkg/algorithm/infra"
	"github.com/wyfcoding/pkg/cache"
	"github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	pb "github.com/wyfcoding/ecommerce/goapi/user/v1"
)

var configFile = flag.String("f", "configs/user/config.toml", "the config file")

func main() {
	flag.Parse()

	// 1. 加载配置
	var cfg config.Config
	if err := config.Load(*configFile, &cfg); err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	config.PrintWithMask(cfg)

	// 2. 初始化 Logger
	logging.InitLogger(cfg.Server.Name, "user-service", cfg.Log.Level)
	logger := logging.Default()
	// slog.SetDefault(logger.Logger) // Already done in InitLogger

	// 3. 初始化 Metrics
	m := metrics.NewMetrics(cfg.Server.Name)

	// 4. 初始化数据库
	db, err := gorm.Open(mysql.Open(cfg.Data.Database.DSN), &gorm.Config{
		Logger: logging.NewGormLogger(logger, 200*time.Millisecond),
	})
	if err != nil {
		logger.ErrorContext(context.Background(), "failed to connect database", "error", err)
		os.Exit(1)
	}
	// 自动迁移 (生产环境建议使用 migrate 工具)
	if cfg.Server.Environment == "dev" {
		db.AutoMigrate(&domain.User{}, &domain.Address{})
	}

	// 5. 初始化 Redis Cache
	// Fix: cfg.Resiliency.CircuitBreaker -> cfg.CircuitBreaker
	redisCache, err := cache.NewRedisCache(&cfg.Data.Redis, cfg.CircuitBreaker, logger, m)
	if err != nil {
		logger.ErrorContext(context.Background(), "failed to init redis cache", "error", err)
		os.Exit(1)
	}
	defer redisCache.Close()

	// 6. 初始化 Kafka Publisher
	var publisher domain.EventPublisher
	if len(cfg.MessageQueue.Kafka.Brokers) > 0 {
		publisher = messaging.NewKafkaPublisher(cfg.MessageQueue.Kafka.Brokers, logger.Logger)
		if closer, ok := publisher.(interface{ Close() error }); ok {
			defer closer.Close()
		}
	} else {
		logger.WarnContext(context.Background(), "Kafka brokers not configured, event publishing disabled")
	}

	// 7. 初始化依赖
	userRepo := persistence.NewUserRepository(db)
	addressRepo := persistence.NewAddressRepository(db)
	antiBot := infra.NewAntiBotDetector()

	// 8. 初始化 Application Service
	cmdService := application.NewUserCommandService(
		userRepo,
		addressRepo,
		publisher,
		redisCache, // Injected
		cfg.MessageQueue.Kafka.Topic,
		cfg.JWT.Secret,
		cfg.JWT.Issuer,
		cfg.JWT.ExpireDuration,
		antiBot,
		logger.Logger, // Pass *slog.Logger
	)
	queryService := application.NewUserQuery(
		userRepo,
		addressRepo,
		redisCache, // Injected
		antiBot,
		logger.Logger, // Pass *slog.Logger
	)

	// 9. 初始化 Handlers
	httpHandler := httpInterface.NewUserHandler(cmdService, queryService)
	grpcHandler := grpcInterface.NewGrpcHandler(cmdService, queryService)

	// 10. 启动 gRPC Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPC.Port))
	if err != nil {
		logger.ErrorContext(context.Background(), "failed to listen tcp", "error", err)
		os.Exit(1)
	}
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, grpcHandler)
	go func() {
		logger.InfoContext(context.Background(), "gRPC server started", "addr", cfg.Server.GRPC.Addr, "port", cfg.Server.GRPC.Port)
		if err := s.Serve(lis); err != nil {
			logger.ErrorContext(context.Background(), "gRPC server failed", "error", err)
		}
	}()

	// 11. 启动 HTTP Server
	r := gin.Default()
	r.GET("/metrics", gin.WrapH(m.Handler()))
	httpHandler.RegisterHandlers(r)
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.HTTP.Port)
		logger.InfoContext(context.Background(), "HTTP server started", "port", cfg.Server.HTTP.Port)
		if err := r.Run(addr); err != nil {
			logger.ErrorContext(context.Background(), "HTTP server failed", "error", err)
		}
	}()

	// 12. 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.InfoContext(context.Background(), "Shutting down server...")
}

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/user/application"
	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/ecommerce/internal/user/infrastructure/messaging"
	"github.com/wyfcoding/ecommerce/internal/user/infrastructure/persistence"
	grpcInterface "github.com/wyfcoding/ecommerce/internal/user/interfaces/grpc"
	httpInterface "github.com/wyfcoding/ecommerce/internal/user/interfaces/http"
	"github.com/wyfcoding/pkg/algorithm/infra"
	"github.com/wyfcoding/pkg/config"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // optimized to use cfg.Log.Level if mapper available
	}))
	slog.SetDefault(logger)

	// 3. 初始化数据库
	db, err := gorm.Open(mysql.Open(cfg.Data.Database.DSN), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}
	// 自动迁移 (生产环境建议使用 migrate 工具)
	if cfg.Server.Environment == "dev" {
		db.AutoMigrate(&domain.User{}, &domain.Address{})
	}

	// 4. 初始化 Kafka Publisher
	// 如果配置中 KafkaEnabled 或 broker 非空 (根据实际 config.toml)
	var publisher domain.EventPublisher
	if len(cfg.MessageQueue.Kafka.Brokers) > 0 {
		publisher = messaging.NewKafkaPublisher(cfg.MessageQueue.Kafka.Brokers, logger)
		// 注意: Kafka Publisher 需要 Close，但 application 层目前没有 Close 接口
		// 可以在 main defer close
		if closer, ok := publisher.(interface{ Close() error }); ok {
			defer closer.Close()
		}
	} else {
		logger.Warn("Kafka brokers not configured, event publishing disabled")
	}

	// 5. 初始化依赖
	userRepo := persistence.NewUserRepository(db)
	addressRepo := persistence.NewAddressRepository(db)
	antiBot := infra.NewAntiBotDetector()

	// 6. 初始化 Application Service
	cmdService := application.NewUserCommandService(
		userRepo,
		addressRepo,
		publisher,
		cfg.MessageQueue.Kafka.Topic,
		cfg.JWT.Secret,
		cfg.JWT.Issuer,
		cfg.JWT.ExpireDuration,
		antiBot,
		logger,
	)
	queryService := application.NewUserQuery(
		userRepo,
		addressRepo,
		antiBot,
		logger,
	)

	// 7. 初始化 Handlers
	httpHandler := httpInterface.NewUserHandler(cmdService, queryService)
	grpcHandler := grpcInterface.NewGrpcHandler(cmdService, queryService)

	// 8. 启动 gRPC Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPC.Port))
	if err != nil {
		logger.Error("failed to listen tcp", "error", err)
		os.Exit(1)
	}
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, grpcHandler)
	go func() {
		logger.Info("gRPC server started", "addr", cfg.Server.GRPC.Addr, "port", cfg.Server.GRPC.Port)
		if err := s.Serve(lis); err != nil {
			logger.Error("gRPC server failed", "error", err)
		}
	}()

	// 9. 启动 HTTP Server
	r := gin.Default()
	httpHandler.RegisterHandlers(r)
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.HTTP.Port)
		logger.Info("HTTP server started", "port", cfg.Server.HTTP.Port)
		if err := r.Run(addr); err != nil {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	// 10. 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")
}

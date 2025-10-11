package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	// Project packages
	"ecommerce/pkg/config"
	"ecommerce/pkg/database/redis"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/snowflake"
	"ecommerce/pkg/tracing"

	// Service-specific imports
	// v1 "ecommerce/api/ai_model/v1"
	"ecommerce/internal/ai_model/biz"
	"ecommerce/internal/ai_model/data"
	aimodelhandler "ecommerce/internal/ai_model/handler"
	"ecommerce/internal/ai_model/service"
)

// Config 结构体用于映射 TOML 配置文件
type Config struct {
	Server struct {
		HTTP struct {
			Addr    string        `toml:"addr"`
			Port    int           `toml:"port"`
			Timeout time.Duration `toml:"timeout"`
		} `toml:"http"`
		GRPC struct {
			Addr    string        `toml:"addr"`
			Port    int           `toml:"port"`
			Timeout time.Duration `toml:"timeout"`
		} `toml:"grpc"`
	}
	Data struct {
		Redis redis.Config `toml:"redis"` // Use pkg/database/redis.Config
		// ... other service client configs
	} `toml:"data"`
	Snowflake snowflake.Config `toml:"snowflake"` // Use pkg/snowflake.Config
	Log       logging.Config   `toml:"log"`       // Use pkg/logging.Config
	Trace     tracing.Config   `toml:"trace"`     // Use pkg/tracing.Config
	Metrics   struct {
		Port string `toml:"port"`
	} `toml:"metrics"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/ai_model.toml", "config file path")
	flag.Parse()

	// 1. 加载配置
	var cfg Config
	if err := config.LoadConfig(configPath, &cfg); err != nil {
		zap.S().Fatalf("failed to load config: %v", err)
	}

	// 2. 初始化日志
	logger := logging.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output)
	zap.ReplaceGlobals(logger)
	defer logger.Sync() // Flush any buffered log entries

	// 3. 初始化追踪
	_, cleanupTracing, err := tracing.InitTracer(&cfg.Trace)
	if err != nil {
		zap.S().Fatalf("failed to init tracing: %v", err)
	}
	defer cleanupTracing()

	// 4. 初始化指标暴露
	cleanupMetrics := metrics.ExposeHttp(cfg.Metrics.Port)
	defer cleanupMetrics()

	// 5. 初始化依赖
	// 初始化 Redis
	redisClient, cleanupRedis, err := redis.NewRedisClient(&cfg.Data.Redis)
	if err != nil {
		zap.S().Fatalf("failed to new redis client: %v", err)
	}
	defer cleanupRedis()

	// 初始化雪花算法
	snowflakeNode, cleanupSnowflake, err := snowflake.NewSnowflakeNode(&cfg.Snowflake)
	if err != nil {
		zap.S().Fatalf("failed to init snowflake: %v", err)
	}
	defer cleanupSnowflake()

	// 6. 依赖注入 (DI)
	dataInstance, cleanupData, err := data.NewData(redisClient) // Assuming data.NewData takes redisClient
	if err != nil {
		zap.S().Fatalf("failed to new data: %v", err)
	}
	defer cleanupData()

	aiModelRepo := data.NewAiModelRepo(dataInstance)
	aiModelUsecase := biz.NewAiModelUsecase(aiModelRepo, snowflakeNode)

	// 初始化服务层
	aiModelService := service.NewAiModelService(aiModelUsecase)

	// 7. 启动 gRPC 服务器
	grpcServer, grpcErrChan := aimodelhandler.StartGRPCServer(aiModelService, cfg.Server.GRPC.Addr, cfg.Server.GRPC.Port)
	if grpcServer == nil {
		zap.S().Fatalf("failed to start gRPC server: %v", <-grpcErrChan)
	}

	// 8. 启动 Gin HTTP 服务器
	ginServer, ginErrChan := aimodelhandler.StartGinServer(aiModelService, cfg.Server.HTTP.Addr, cfg.Server.HTTP.Port)
	if ginServer == nil {
		zap.S().Fatalf("failed to start Gin HTTP server: %v", <-ginErrChan)
	}

	// 9. 等待中断信号或服务器错误以实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down ai_model service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down ai_model service due to gRPC error...")
	case err := <-ginErrChan:
		zap.S().Errorf("Gin HTTP server error: %v", err)
		zap.S().Info("Shutting down ai_model service due to Gin HTTP error...")
	}

	// 优雅地关闭服务器
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ginServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("Gin HTTP server shutdown error: %v", err)
	}
}

// startGRPCServer 启动 gRPC 服务器

// startGinServer 启动 Gin HTTP 服务器

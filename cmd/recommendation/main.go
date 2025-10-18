package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	recommendationhandler "ecommerce/internal/recommendation/handler"
	"ecommerce/internal/recommendation/repository"
	"ecommerce/internal/recommendation/service"
	recommendationclient "ecommerce/internal/recommendation/client"
	configpkg "ecommerce/pkg/config"
	mysqlpkg "ecommerce/pkg/database/mysql"
	redisPkg "ecommerce/pkg/database/redis"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/snowflake"
	"ecommerce/pkg/tracing"
)

// Config 结构体用于映射 TOML 配置文件
type Config struct {
	configpkg.ServerConfig `toml:"server"`
	Data                   struct {
		Database mysqlpkg.Config `toml:"database"`
		Redis    redisPkg.Config `toml:"redis"`
		AIModelService struct {
			Addr string `toml:"addr"`
		} `toml:"ai_model_service"`
	} `toml:"data"`
	Snowflake snowflake.Config `toml:"snowflake"`
	Log       logging.Config   `toml:"log"`
	Trace     tracing.Config   `toml:"trace"`
	Metrics   struct {
		Port string `toml:"port"`
	} `toml:"metrics"`
}

func main() {
	// 1. 加载配置
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/recommendation.toml", "config file path")
	flag.Parse()

	var cfg Config
	if err := configpkg.LoadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger := logging.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output)
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// 3. 初始化追踪
	_, cleanupTracing, err := tracing.InitTracer(&cfg.Trace)
	if err != nil {
		zap.S().Fatalf("failed to init tracing: %v", err)
	}
	defer cleanupTracing()

	// 4. 初始化指标暴露
	cleanupMetrics := metrics.ExposeHttp(cfg.Metrics.Port)
	defer cleanupMetrics()

	// 5. 初始化雪花算法
	snowflakeNode, err := snowflake.NewSnowflakeNode(&cfg.Snowflake)
	if err != nil {
		zap.S().Fatalf("failed to init snowflake: %v", err)
	}

	// 6. 依赖注入 (DI)
	// 数据库连接
	db, cleanupDB, err := mysqlpkg.NewGORMDB(&cfg.Data.Database)
	if err != nil {
		zap.S().Fatalf("failed to connect database: %v", err)
	}
	defer cleanupDB()

	// 初始化 Redis
	redisClient, redisCleanup, err := redisPkg.NewRedisClient(&cfg.Data.Redis)
	if err != nil {
		zap.S().Fatalf("failed to new redis client: %v", err)
	}
	defer redisCleanup()

	// 初始化 AI Model service client
	aiModelConn, err := grpc.Dial(cfg.Data.AIModelService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Fatalf("failed to connect to AI model service: %v", err)
	}
	defer aiModelConn.Close()
	aiModelClient := recommendationclient.NewAIModelClient(aiModelConn)

	recommendationRepo := repository.NewRecommendationRepo(db, redisClient, snowflakeNode)
	recommendationService := service.NewRecommendationService(recommendationRepo, aiModelClient)

	// 7. 启动 gRPC 和 HTTP Gateway
	grpcServer, grpcErrChan := recommendationhandler.StartGRPCServer(recommendationService, cfg.Server.GRPC.Addr, cfg.Server.GRPC.Port)
	if grpcServer == nil {
		zap.S().Fatalf("failed to start gRPC server: %v", <-grpcErrChan)
	}
	httpServer, httpErrChan := recommendationhandler.StartHTTPServer(context.Background(), cfg.Server.GRPC.Addr, cfg.Server.GRPC.Port, cfg.Server.HTTP.Addr, cfg.Server.HTTP.Port)
	if httpServer == nil {
		zap.S().Fatalf("failed to start HTTP server: %v", <-httpErrChan)
	}

	// 8. 等待中断信号或服务器错误以实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down recommendation service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
	case err := <-httpErrChan:
		zap.S().Errorf("HTTP server error: %v", err)
	}

	// 优雅地关闭服务器
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("HTTP server shutdown error: %v", err)
	}
}

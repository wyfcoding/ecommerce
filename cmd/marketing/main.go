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

	v1 "ecommerce/api/marketing/v1"
	marketinghandler "ecommerce/internal/marketing/handler"
	"ecommerce/internal/marketing/model"
	"ecommerce/internal/marketing/repository"
	"ecommerce/internal/marketing/service"
	configpkg "ecommerce/pkg/config"
	mysqlpkg "ecommerce/pkg/database/mysql"
	redisPkg "ecommerce/pkg/database/redis"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"
)

// Config 结构体用于映射 TOML 配置文件
type Config struct {
	configpkg.ServerConfig `toml:"server"`
	Data                   struct {
		Database mysqlpkg.Config `toml:"database"`
		Redis    redisPkg.Config `toml:"redis"`
	} `toml:"data"`
	configpkg.LogConfig `toml:"log"`
	configpkg.TraceConfig `toml:"trace"`
	Metrics                struct {
		Port string `toml:"port"`
	} `toml:"metrics"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/marketing.toml", "config file path")
	flag.Parse()

	// 1. 加载配置
	var cfg Config
	if err := configpkg.LoadConfig(configPath, &cfg); err != nil {
		zap.S().Fatalf("failed to load config: %v", err)
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

	// 5. 初始化依赖
	// 数据库连接
	db, cleanupDB, err := mysqlpkg.NewGORMDB(&cfg.Data.Database)
	if err != nil {
		zap.S().Fatalf("failed to connect database: %v", err)
	}
	defer cleanupDB()

	// 自动迁移数据库表结构
	if err := db.AutoMigrate(&model.CouponTemplate{}, &model.UserCoupon{}, &model.Promotion{}); err != nil {
		zap.S().Fatalf("failed to migrate database: %v", err)
	}

	// 初始化 Redis
	redisClient, redisCleanup, err := redisPkg.NewRedisClient(&cfg.Data.Redis)
	if err != nil {
		zap.S().Fatalf("failed to new redis client: %v", err)
	}
	defer redisCleanup()

	// 6. 依赖注入 (DI)
	couponRepo := repository.NewCouponRepo(db, redisClient)
	promotionRepo := repository.NewPromotionRepo(db, redisClient)
	couponService := service.NewCouponService(couponRepo, zap.S())
	promotionService := service.NewPromotionService(promotionRepo, zap.S())

	// 7. 启动 gRPC 服务器
	grpcServer, grpcErrChan := marketinghandler.StartGRPCServer(couponService, promotionService, cfg.Server.GRPC.Addr, cfg.Server.GRPC.Port)
	if grpcServer == nil {
		zap.S().Fatalf("failed to start gRPC server: %v", <-grpcErrChan)
	}

	// 8. 启动 Gin HTTP 服务器
	ginServer, ginErrChan := marketinghandler.StartHTTPServer(couponService, promotionService, cfg.Server.HTTP.Addr, cfg.Server.HTTP.Port)
	if ginServer == nil {
		zap.S().Fatalf("failed to start Gin HTTP server: %v", <-ginErrChan)
	}

	// 9. 等待中断信号或服务器错误以实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down marketing service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
	case err := <-ginErrChan:
		zap.S().Errorf("Gin HTTP server error: %v", err)
	}

	// 优雅地关闭服务器
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ginServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("Gin HTTP server shutdown error: %v", err)
	}
}

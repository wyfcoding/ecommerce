package main

import (
	"context"
	"ecommerce/api/analytics/v1"
	"ecommerce/internal/analytics/biz"
	"ecommerce/internal/analytics/data"
	"ecommerce/internal/analytics/service"
	"ecommerce/pkg/database/clickhouse"
	"ecommerce/pkg/logging"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Config 结构体用于映射 analytics.toml 配置文件
type Config struct {
	Server struct {
		Grpc struct {
			Addr string `toml:"addr"`
			Port int    `toml:"port"`
		} `toml:"grpc"`
		Http struct {
			Addr string `toml:"addr"`
			Port int    `toml:"port"`
		} `toml:"http"`
	} `toml:"server"`
	Log struct {
		Level  string `toml:"level"`
		Format string `toml:"format"`
		Output string `toml:"output"`
	} `toml:"log"`
	Data struct {
		Clickhouse clickhouse.Config `toml:"clickhouse"`
	} `toml:"data"`
}

func main() {
	// 1. 加载配置
	config, err := loadConfig("./configs/analytics.toml")
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger := logging.NewLogger(config.Log.Level, config.Log.Format, config.Log.Output)
	zap.ReplaceGlobals(logger)

	// 3. 初始化 ClickHouse 连接
	chClient, chCleanup, err := clickhouse.NewClickHouseClient(&config.Data.Clickhouse)
	if err != nil {
		zap.S().Fatalf("failed to connect to clickhouse: %v", err)
	}
	defer chCleanup()

	// 4. 初始化 Data 层
	dataInstance, dataCleanup := data.NewData(nil) // No GORM DB for analytics
	defer dataCleanup()

	// 5. 初始化 Repo 层
	analyticsRepo := data.NewAnalyticsRepo(dataInstance, chClient)

	// 6. 初始化 Usecase 层
	analyticsUsecase := biz.NewAnalyticsUsecase(analyticsRepo)

	// 7. 初始化 gRPC Server
	s := grpc.NewServer()
	analyticsService := service.NewAnalyticsService(analyticsUsecase)
	v1.RegisterAnalyticsServiceServer(s, analyticsService)

	// 8. 启动 gRPC Server
	grpcAddr := fmt.Sprintf("%s:%d", config.Server.Grpc.Addr, config.Server.Grpc.Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		zap.S().Fatalf("failed to listen: %v", err)
	}

	go func() {
		zap.S().Infof("Analytics Service gRPC server listening on %s", grpcAddr)
		if err := s.Serve(lis); err != nil {
			zap.S().Fatalf("failed to serve: %v", err)
		}
	}()

	// 9. 优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.S().Info("Shutting down Analytics Service...")
	s.GracefulStop()
}

// loadConfig 从 TOML 文件加载配置
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

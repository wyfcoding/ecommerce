package main

import (
	"context"
	"ecommerce/api/asset/v1"
	"ecommerce/internal/asset/biz"
	"ecommerce/internal/asset/data"
	"ecommerce/internal/asset/service"
	"ecommerce/pkg/database/mysql"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/minio"
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

// Config 结构体用于映射 asset.toml 配置文件
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
		Database mysql.Config `toml:"database"`
		Minio    minio.Config `toml:"minio"`
	} `toml:"data"`
}

func main() {
	// 1. 加载配置
	config, err := loadConfig("./configs/asset.toml")
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger := logging.NewLogger(config.Log.Level, config.Log.Format, config.Log.Output)
	zap.ReplaceGlobals(logger)

	// 3. 初始化数据库连接
	db, dbCleanup, err := mysql.NewGORMDB(&config.Data.Database)
	if err != nil {
		zap.S().Fatalf("failed to connect to mysql: %v", err)
	}
	defer dbCleanup()

	// 4. 初始化 MinIO 客户端
	minioClient, err := minio.NewMinioClient(&config.Data.Minio)
	if err != nil {
		zap.S().Fatalf("failed to connect to minio: %v", err)
	}

	// 5. 初始化 Data 层
	dataInstance, dataCleanup := data.NewData(db)
	defer dataCleanup()

	// 6. 初始化 Repo 层
	assetRepo := data.NewAssetRepo(dataInstance, minioClient)

	// 7. 初始化 Usecase 层
	assetUsecase := biz.NewAssetUsecase(assetRepo)

	// 8. 初始化 gRPC Server
	s := grpc.NewServer()
	assetService := service.NewAssetService(assetUsecase)
	v1.RegisterAssetServiceServer(s, assetService)

	// 9. 启动 gRPC Server
	grpcAddr := fmt.Sprintf("%s:%d", config.Server.Grpc.Addr, config.Server.Grpc.Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		zap.S().Fatalf("failed to listen: %v", err)
	}

	go func() {
		zap.S().Infof("Asset Service gRPC server listening on %s", grpcAddr)
		if err := s.Serve(lis); err != nil {
			zap.S().Fatalf("failed to serve: %v", err)
		}
	}()

	// 10. 优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.S().Info("Shutting down Asset Service...")
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

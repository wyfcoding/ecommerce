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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	grpcServer := grpc.NewServer()
	assetService := service.NewAssetService(assetUsecase)
	v1.RegisterAssetServiceServer(grpcServer, assetService)

	// 9. 启动 gRPC 和 HTTP Gateway
	grpcErrChan := make(chan error, 1)
	go func() {
		grpcAddr := fmt.Sprintf("%s:%d", config.Server.Grpc.Addr, config.Server.Grpc.Port)
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			grpcErrChan <- fmt.Errorf("failed to listen gRPC: %w", err)
			return
		}
		zap.S().Infof("Asset Service gRPC server listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			grpcErrChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(grpcErrChan)
	}()

	httpServer, httpErrChan := startHTTPServer(context.Background(), config.Server.Grpc.Addr, config.Server.Grpc.Port, config.Server.Http.Addr, config.Server.Http.Port)
	if httpServer == nil {
		zap.S().Fatalf("failed to start HTTP server: %v", <-httpErrChan)
	}

	// 10. 优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down Asset Service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down Asset Service due to gRPC error...")
	case err := <-httpErrChan:
		zap.S().Errorf("HTTP server error: %v", err)
		zap.S().Info("Shutting down Asset Service due to HTTP error...")
	}

	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("HTTP server shutdown error: %v", err)
	}
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

// startHTTPServer 启动 HTTP Gateway
func startHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterAssetServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for AssetService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(gin.Recovery())
	// Add service-specific Gin routes here
	// For example:
	// r.POST("/asset/upload", handler.UploadAsset)

	r.Any("/*any", gin.WrapH(mux))

	httpEndpoint := fmt.Sprintf("%s:%d", httpAddr, httpPort)
	server := &http.Server{
		Addr:    httpEndpoint,
		Handler: r,
	}

	zap.S().Infof("HTTP server listening at %s", httpEndpoint)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to serve HTTP: %w", err)
		}
		close(errChan)
	}()
	return server, errChan
}

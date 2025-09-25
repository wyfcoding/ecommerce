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
<<<<<<< HEAD
	"net/http"
=======
>>>>>>> 04d1270d593e17e866ec0ca4dad1f5d56021f07d
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
<<<<<<< HEAD
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
=======
	"go.uber.org/zap"
	"google.golang.org/grpc"
>>>>>>> 04d1270d593e17e866ec0ca4dad1f5d56021f07d
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
<<<<<<< HEAD
	grpcServer := grpc.NewServer()
	analyticsService := service.NewAnalyticsService(analyticsUsecase)
	v1.RegisterAnalyticsServiceServer(grpcServer, analyticsService)

	// 8. 启动 gRPC 和 HTTP Gateway
	grpcErrChan := make(chan error, 1)
	go func() {
		grpcAddr := fmt.Sprintf("%s:%d", config.Server.Grpc.Addr, config.Server.Grpc.Port)
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			grpcErrChan <- fmt.Errorf("failed to listen gRPC: %w", err)
			return
		}
		zap.S().Infof("Analytics Service gRPC server listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			grpcErrChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(grpcErrChan)
	}()

	httpServer, httpErrChan := startHTTPServer(context.Background(), config.Server.Grpc.Addr, config.Server.Grpc.Port, config.Server.Http.Addr, config.Server.Http.Port)
	if httpServer == nil {
		zap.S().Fatalf("failed to start HTTP server: %v", <-httpErrChan)
	}

	// 9. 优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down Analytics Service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down Analytics Service due to gRPC error...")
	case err := <-httpErrChan:
		zap.S().Errorf("HTTP server error: %v", err)
		zap.S().Info("Shutting down Analytics Service due to HTTP error...")
	}

	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("HTTP server shutdown error: %v", err)
	}
=======
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
>>>>>>> 04d1270d593e17e866ec0ca4dad1f5d56021f07d
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
<<<<<<< HEAD

// startHTTPServer 启动 HTTP Gateway
func startHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterAnalyticsServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for AnalyticsService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(gin.Recovery())
	// Add service-specific Gin routes here
	// For example:
	// r.GET("/analytics/report", handler.GetReport)

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
=======
>>>>>>> 04d1270d593e17e866ec0ca4dad1f5d56021f07d

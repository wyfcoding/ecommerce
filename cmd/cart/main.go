package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce/internal/cart/biz"
	"ecommerce/internal/cart/data"
	carthandler "ecommerce/internal/cart/handler"
	"ecommerce/internal/cart/service"
	configpkg "ecommerce/pkg/config" // Added this line
	"ecommerce/pkg/logging"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	configpkg.ServerConfig `toml:"server"` // Embed common server config
	Data                   struct {
		Redis struct {
			Addr         string `toml:"addr"`
			Password     string `toml:"password"`
			DB           int    `toml:"db"`
			ReadTimeout  string `toml:"read_timeout"`
			WriteTimeout string `toml:"write_timeout"`
		} `toml:"redis"`
		ProductService struct {
			Addr string `toml:"addr"`
		} `toml:"product_service"`
	} `toml:"data"`
	configpkg.LogConfig `toml:"log"` // Embed common log config
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/cart.toml", "config file path")
	flag.Parse()

	// 1. 加载配置
	var config Config
	if err := configpkg.LoadConfig(configPath, &config); err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger := logging.NewLogger(config.Log.Level, config.Log.Format, config.Log.Output)
	zap.ReplaceGlobals(logger)

	// 3. 初始化依赖
	dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	productServiceConn, err := grpc.DialContext(dialCtx, config.Data.ProductService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		zap.S().Fatalf("failed to connect to product service: %v", err)
	}
	defer productServiceConn.Close()

	readTimeout, err := time.ParseDuration(config.Data.Redis.ReadTimeout)
	if err != nil {
		zap.S().Fatalf("failed to parse redis read timeout: %v", err)
	}
	writeTimeout, err := time.ParseDuration(config.Data.Redis.WriteTimeout)
	if err != nil {
		zap.S().Fatalf("failed to parse redis write timeout: %v", err)
	}
	redisConfig := &data.RedisConfig{
		Addr:         config.Data.Redis.Addr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
	dataInstance, cleanup, err := data.NewData(redisConfig)
	if err != nil {
		zap.S().Fatalf("failed to new data: %v", err)
	}
	defer cleanup()

	// 4. 依赖注入 (DI)
	productClient := data.NewProductClient(productServiceConn)
	cartRepo := data.NewCartRepo(dataInstance)
	cartUsecase := biz.NewCartUsecase(cartRepo, productClient)
	cartService := service.NewCartService(cartUsecase)

	// 5. 启动 gRPC 和 HTTP Gateway
	grpcServer, grpcErrChan := carthandler.StartGRPCServer(cartService, config.Server.GRPC.Addr, config.Server.GRPC.Port)
	if grpcServer == nil {
		zap.S().Fatalf("failed to start gRPC server: %v", <-grpcErrChan)
	}
	httpServer, httpErrChan := carthandler.StartHTTPServer(context.Background(), config.Server.GRPC.Addr, config.Server.GRPC.Port, config.Server.HTTP.Addr, config.Server.HTTP.Port)
	if httpServer == nil {
		zap.S().Fatalf("failed to start HTTP server: %v", <-httpErrChan)
	}

	// 6. 等待中断信号或服务器错误以实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down cart service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down cart service due to gRPC error...")
	case err := <-httpErrChan:
		zap.S().Errorf("HTTP server error: %v", err)
		zap.S().Info("Shutting down cart service due to HTTP error...")
	}

	// 优雅地关闭服务器
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("HTTP server shutdown error: %v", err)
	}
}

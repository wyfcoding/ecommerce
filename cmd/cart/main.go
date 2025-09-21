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

	v1 "ecommerce/api/cart/v1"
	"ecommerce/internal/cart/biz"
	"ecommerce/internal/cart/data"
	"ecommerce/internal/cart/service"
	"ecommerce/pkg/logging"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Server struct {
		GRPC struct {
			Addr    string `toml:"addr"`
			Port    int    `toml:"port"`
			Timeout string `toml:"timeout"`
		} `toml:"grpc"`
	} `toml:"server"`
	Data struct {
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
	Log struct {
		Level  string `toml:"level"`
		Format string `toml:"format"`
		Output string `toml:"output"`
	} `toml:"log"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/cart.toml", "config file path")
	flag.Parse()

	// 1. 加载配置
	config, err := loadConfig(configPath)
	if err != nil {
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

	// 5. 启动 gRPC 服务器
	grpcServer, errChan := startGRPCServer(cartService, config.Server.GRPC.Addr, config.Server.GRPC.Port)
	if grpcServer == nil {
		zap.S().Fatalf("failed to start gRPC server: %v", <-errChan)
	}

	// 6. 等待中断信号或服务器错误以实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down cart service...")
	case err := <-errChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down cart service due to error...")
	}

	grpcServer.GracefulStop()
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func startGRPCServer(cartService *service.CartService, addr string, port int) (*grpc.Server, chan error) {
	errChan := make(chan error, 1)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %w", err)
		return nil, errChan
	}
	s := grpc.NewServer()
	v1.RegisterCartServer(s, cartService)

	zap.S().Infof("gRPC server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(errChan)
	}()
	return s, errChan
}

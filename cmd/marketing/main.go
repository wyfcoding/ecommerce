package main

import (
	"context"
	"flag"
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
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	v1 "ecommerce/api/marketing/v1"
	"ecommerce/internal/marketing/biz"
	"ecommerce/internal/marketing/data"
	"ecommerce/internal/marketing/service"
	"ecommerce/pkg/logging"
)

// Config 结构体用于映射 TOML 配置文件
type Config struct {
	Server struct {
		HTTP struct {
			Addr    string `toml:"addr"`
			Port    int    `toml:"port"`
			Timeout string `toml:"timeout"`
		} `toml:"http"`
		GRPC struct {
			Addr    string `toml:"addr"`
			Port    int    `toml:"port"`
			Timeout string `toml:"timeout"`
		} `toml:"grpc"`
	} `toml:"server"`
	Data struct {
		Database struct {
			DSN string `toml:"dsn"`
		} `toml:"database"`
		Redis struct {
			Addr         string `toml:"addr"`
			ReadTimeout  string `toml:"read_timeout"`
			WriteTimeout string `toml:"write_timeout"`
		} `toml:"redis"`
	} `toml:"data"`
	Log struct {
		Level  string `toml:"level"`
		Format string `toml:"format"`
		Output string `toml:"output"`
	} `toml:"log"`
}

// loadConfig 从指定路径加载并解析 TOML 配置文件
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

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/marketing.toml", "config file path")
	flag.Parse()

	// 1. 加载配置
	config, err := loadConfig(configPath)
	if err != nil {
		zap.S().Fatalf("failed to load config: %v", err)
	}

	// 2. 初始化日志
	logger := logging.NewLogger(config.Log.Level, config.Log.Format, config.Log.Output)
	zap.ReplaceGlobals(logger)

	// 3. 初始化依赖
	// 数据库连接
	db, err := gorm.Open(mysql.Open(config.Data.Database.DSN), &gorm.Config{})
	if err != nil {
		zap.S().Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移数据库表结构 (TODO: 仅在开发环境或需要时执行)
	// if err := db.AutoMigrate(&data.Coupon{}, &data.CouponTemplate{}); err != nil {
	// 	zap.S().Fatalf("failed to migrate database: %v", err)
	// }

	sqlDB, err := db.DB()
	if err != nil {
		zap.S().Fatalf("failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// Redis 客户端
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
	dataInstance, cleanup, err := data.NewData(db, redisConfig)
	if err != nil {
		zap.S().Fatalf("failed to new data: %v", err)
	}
	defer cleanup()

	// 4. 依赖注入 (DI)
	couponRepo := data.NewCouponRepo(dataInstance)
	marketingUsecase := biz.NewMarketingUsecase(couponRepo)
	marketingService := service.NewMarketingService(marketingUsecase)

	// 5. 启动 gRPC 和 HTTP Gateway
	grpcServer, grpcErrChan := startGRPCServer(marketingService, config.Server.GRPC.Addr, config.Server.GRPC.Port)
	if grpcServer == nil {
		zap.S().Fatalf("failed to start gRPC server: %v", <-grpcErrChan)
	}
	httpServer, httpErrChan := startHTTPServer(context.Background(), config.Server.GRPC.Addr, config.Server.GRPC.Port, config.Server.HTTP.Addr, config.Server.HTTP.Port)
	if httpServer == nil {
		zap.S().Fatalf("failed to start HTTP server: %v", <-httpErrChan)
	}

	// 6. 等待中断信号或服务器错误以实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down marketing service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down marketing service due to gRPC error...")
	case err := <-httpErrChan:
		zap.S().Errorf("HTTP server error: %v", err)
		zap.S().Info("Shutting down marketing service due to HTTP error...")
	}

	// 优雅地关闭服务器
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("HTTP server shutdown error: %v", err)
	}
}

// startGRPCServer 启动 gRPC 服务器
func startGRPCServer(marketingService *service.MarketingService, addr string, port int) (*grpc.Server, chan error) {
	errChan := make(chan error, 1)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %w", err)
		return nil, errChan
	}
	s := grpc.NewServer()
	v1.RegisterMarketingServer(s, marketingService)

	zap.S().Infof("gRPC server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(errChan)
	}()
	return s, errChan
}

// startHTTPServer 启动 HTTP Gateway，它会将 HTTP 请求代理到 gRPC 服务
func startHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterMarketingHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(gin.Recovery())
	// Add service-specific Gin routes here
	// For example:
	// r.GET("/marketing/coupons", handler.GetCoupons)

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

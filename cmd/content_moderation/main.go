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

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/BurntSushi/toml"

	v1 "ecommerce/api/content_moderation/v1"
	"ecommerce/internal/content_moderation/biz"
	"ecommerce/internal/content_moderation/data"
	"ecommerce/internal/content_moderation/service"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/snowflake"
	"ecommerce/pkg/database/redis"
)

// Config 结构体用于映射 TOML 配置文件
type Config struct {
	Server struct {
		HTTP struct {
			Addr    string `toml:"addr"`
			Port    int    `toml:"port"`
			Timeout time.Duration `toml:"timeout"`
		} `toml:"http"`
		GRPC struct {
			Addr    string `toml:"addr"`
			Port    int    `toml:"port"`
			Timeout time.Duration `toml:"timeout"`
		} `toml:"grpc"`
	} `toml:"server"`
	Data struct {
		Database struct {
			DSN string `toml:"dsn"`
		} `toml:"database"`
		Redis struct {
			Addr         string `toml:"addr"`
			Password     string `toml:"password"`
			DB           int    `toml:"db"`
			ReadTimeout  time.Duration `toml:"read_timeout"`
			WriteTimeout time.Duration `toml:"write_timeout"`
		} `toml:"redis"`
	} `toml:"data"`
	Snowflake struct {
		StartTime string `toml:"start_time"`
		MachineID int64  `toml:"machine_id"`
	} `toml:"snowflake"`
	Log struct {
		Level  string `toml:"level"`
		Format string `toml:"format"`
		Output string `toml:"output"`
	} `toml:"log"`
}

func main() {
	// 1. 加载配置
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/content_moderation.toml", "config file path")
	flag.Parse()

	config, err := loadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger := logging.NewLogger(config.Log.Level, config.Log.Format, config.Log.Output)
	zap.ReplaceGlobals(logger)

	// 3. 初始化雪花算法
	if err := snowflake.Init(config.Snowflake.StartTime, config.Snowflake.MachineID); err != nil {
		zap.S().Fatalf("failed to init snowflake: %v", err)
	}

	// 4. 依赖注入 (DI)
	dataInstance, cleanup, err := data.NewData(config.Data.Database.DSN)
	if err != nil {
		zap.S().Fatalf("failed to new data: %v", err)
	}
	defer cleanup()

	// 初始化 Redis
	redisClient, redisCleanup, err := redis.NewRedisClient(&redis.Config{
		Addr:         config.Data.Redis.Addr,
		Password:     config.Data.Redis.Password,
		DB:           config.Data.Redis.DB,
		ReadTimeout:  config.Data.Redis.ReadTimeout,
		WriteTimeout: config.Data.Redis.WriteTimeout,
		PoolSize:     10, // 默认值
		MinIdleConns: 5,  // 默认值
	})
	if err != nil {
		zap.S().Fatalf("failed to new redis client: %v", err)
	}
	defer redisCleanup()

	// 初始化业务层
	contentModerationRepo := data.NewContentModerationRepo(dataInstance, redisClient)
	contentModerationUsecase := biz.NewContentModerationUsecase(contentModerationRepo)

	// 初始化服务层
	contentModerationService := service.NewContentModerationService(contentModerationUsecase)

	// 5. 启动 gRPC 和 HTTP Gateway
	grpcServer, grpcErrChan := startGRPCServer(contentModerationService, config.Server.GRPC.Addr, config.Server.GRPC.Port)
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
		zap.S().Info("Shutting down content_moderation service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down content_moderation service due to gRPC error...")
	case err := <-httpErrChan:
		zap.S().Errorf("HTTP server error: %v", err)
		zap.S().Info("Shutting down content_moderation service due to HTTP error...")
	}

	// 优雅地关闭服务器
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("HTTP server shutdown error: %v", err)
	}
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

// startGRPCServer 启动 gRPC 服务器
func startGRPCServer(svc *service.ContentModerationService, addr string, port int) (*grpc.Server, chan error) {
	errChan := make(chan error, 1)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %w", err)
		return nil, errChan
	}
	s := grpc.NewServer()
	v1.RegisterContentModerationServiceServer(s, svc)

	zap.S().Infof("gRPC server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(errChan)
	}()
	return s, errChan
}

// startHTTPServer 启动 HTTP Gateway
func startHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterContentModerationServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for ContentModerationService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(gin.Recovery())
	// Add service-specific Gin routes here
	// For example:
	// r.POST("/content_moderation/text", handler.ModerateText)

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

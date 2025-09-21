package main

import (
	"context"
	"ecommerce/api/user/v1"
	"ecommerce/internal/user/biz"
	"ecommerce/internal/user/data"
	"ecommerce/internal/user/service"
	"ecommerce/pkg/jwt"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/snowflake"
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
)

// Config 结构体用于映射 TOML 配置文件
type Config struct {
	Server struct {
			Timeout time.Duration `toml:"timeout"`
		} `toml:"http"`GRPC struct {
			Addr    string `toml:"addr"`
			Port    int    `toml:"port"`
			Timeout time.Duration `toml:"timeout"`
		} `toml:"grpc"`
	} `toml:"server"`
	Data struct {
		Database struct {
			Driver string `toml:"driver"`
			DSN    string `toml:"dsn"`
		} `toml:"database"`
		Redis struct {
			Addr         string `toml:"addr"`
			Password     string `toml:"password"`
			DB           int    `toml:"db"`
			ReadTimeout  time.Duration `toml:"read_timeout"`
			WriteTimeout time.Duration `toml:"write_timeout"`
		} `toml:"redis"`
	} `toml:"data"`
	JWT struct {
		Secret   string `toml:"secret"`
		Issuer   string `toml:"issuer"`
		Expire   time.Duration `toml:"expire_duration"`
	} `toml:"jwt"`
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
	flag.StringVar(&configPath, "conf", "../../configs/user.toml", "config file path")
	flag.Parse()

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. 初始化日志
	logger := logging.NewLogger(config.Log.Level, config.Log.Format, config.Log.Output)
	zap.ReplaceGlobals(logger) // 将我们的 logger 设置为全局 logger

	// 3. 初始化雪花算法
	if err := snowflake.Init(config.Snowflake.StartTime, config.Snowflake.MachineID); err != nil {
		zap.S().Fatalf("failed to init snowflake: %v", err)
	}

	// 4. 依赖注入 (DI)
	// 从下到上构建依赖关系：data -> biz -> service
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

	userRepo := data.NewUserRepo(dataInstance)
	addressRepo := data.NewAddressRepo(dataInstance)

	userUsecase := biz.NewUserUsecase(userRepo, config.JWT.Secret, config.JWT.Issuer, config.JWT.Expire)
	addressUsecase := biz.NewAddressUsecase(addressRepo)

	userService := service.NewUserService(userUsecase, addressUsecase)

	// 5. 启动 gRPC 和 HTTP Gateway
	grpcServer, grpcErrChan := startGRPCServer(userService, config.Server.GRPC.Addr, config.Server.GRPC.Port)
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
		zap.S().Info("Shutting down servers...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down servers due to gRPC error...")
	case err := <-httpErrChan:
		zap.S().Errorf("HTTP server error: %v", err)
		zap.S().Info("Shutting down servers due to HTTP error...")
	}

	// 优雅地关闭服务器
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("HTTP server shutdown error: %v", err)
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
func startGRPCServer(userService *service.UserService, addr string, port int) (*grpc.Server, chan error) {
	errChan := make(chan error, 1)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %w", err)
		return nil, errChan
	}
	s := grpc.NewServer()
	v1.RegisterUserServer(s, userService)

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
	err := v1.RegisterUserHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway: %w", err)
		return nil, errChan
	}

	// 使用 Gin 作为 HTTP 服务器的引擎
	r := gin.Default()
	// 将 grpc-gateway 的处理器集成到 Gin
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
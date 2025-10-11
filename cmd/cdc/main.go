package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"

	"ecommerce/internal/cdc/biz"
	"ecommerce/internal/cdc/data"
	cdchandler "ecommerce/internal/cdc/handler"
	"ecommerce/internal/cdc/service"
	"ecommerce/pkg/database/redis"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/snowflake"
)

// Config 结构体用于映射 TOML 配置文件
type Config struct {
	Server struct {
		HTTP struct {
			Addr    string        `toml:"addr"`
			Port    int           `toml:"port"`
			Timeout time.Duration `toml:"timeout"`
		} `toml:"http"`
		GRPC struct {
			Addr    string        `toml:"addr"`
			Port    int           `toml:"port"`
			Timeout time.Duration `toml:"timeout"`
		} `toml:"grpc"`
	} `toml:"server"`
	Data struct {
		Database struct {
			DSN string `toml:"dsn"`
		} `toml:"database"`
		Redis struct {
			Addr         string        `toml:"addr"`
			Password     string        `toml:"password"`
			DB           int           `toml:"db"`
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
	flag.StringVar(&configPath, "conf", "./configs/cdc.toml", "config file path")
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
	cdcRepo := data.NewCdcRepo(dataInstance, redisClient)
	cdcUsecase := biz.NewCdcUsecase(cdcRepo)

	// 初始化服务层
	cdcService := service.NewCdcService(cdcUsecase)

	// 5. 启动 gRPC 和 HTTP Gateway
	grpcServer, grpcErrChan := cdchandler.StartGRPCServer(cdcService, config.Server.GRPC.Addr, config.Server.GRPC.Port)
	if grpcServer == nil {
		zap.S().Fatalf("failed to start gRPC server: %v", <-grpcErrChan)
	}
	httpServer, httpErrChan := cdchandler.StartHTTPServer(context.Background(), config.Server.GRPC.Addr, config.Server.GRPC.Port, config.Server.HTTP.Addr, config.Server.HTTP.Port)
	if httpServer == nil {
		zap.S().Fatalf("failed to start HTTP server: %v", <-httpErrChan)
	}

	// 6. 等待中断信号或服务器错误以实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down cdc service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down cdc service due to gRPC error...")
	case err := <-httpErrChan:
		zap.S().Errorf("HTTP server error: %v", err)
		zap.S().Info("Shutting down cdc service due to HTTP error...")
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

// startHTTPServer 启动 HTTP Gateway

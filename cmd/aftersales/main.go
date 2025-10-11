package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	// Project packages
	"ecommerce/pkg/config"
	"ecommerce/pkg/database/mysql"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	// Service-specific imports

	"ecommerce/internal/aftersales/biz"
	"ecommerce/internal/aftersales/data" // Assuming a model package
	aftersaleshandler "ecommerce/internal/aftersales/handler"
	"ecommerce/internal/aftersales/service"
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
		Database mysql.Config `toml:"database"` // 使用 pkg/database/mysql.Config
		// ... 其他服务客户端配置
	} `toml:"data"`
	Log     logging.Config `toml:"log"`   // 使用 pkg/logging.Config
	Trace   tracing.Config `toml:"trace"` // 使用 pkg/tracing.Config
	Metrics struct {
		Port string `toml:"port"`
	} `toml:"metrics"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/aftersales.toml", "config file path")
	flag.Parse()

	// 1. 加载配置
	var cfg Config
	if err := config.LoadConfig(configPath, &cfg); err != nil {
		zap.S().Fatalf("failed to load config: %v", err)
	}

	// 2. 初始化日志
	logger := logging.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output)
	zap.ReplaceGlobals(logger)
	defer logger.Sync() // 刷新任何缓冲的日志条目

	// 3. 初始化追踪
	_, cleanupTracing, err := tracing.InitTracer(&cfg.Trace)
	if err != nil {
		zap.S().Fatalf("failed to init tracing: %v", err)
	}
	defer cleanupTracing()

	// 4. 初始化指标暴露
	cleanupMetrics := metrics.ExposeHttp(cfg.Metrics.Port)
	defer cleanupMetrics()

	// 5. 初始化依赖
	// 数据库连接
	db, cleanupDB, err := mysql.NewGORMDB(&cfg.Data.Database)
	if err != nil {
		zap.S().Fatalf("failed to connect database: %v", err)
	}
	defer cleanupDB()

	// 自动迁移数据库表结构（假设售后服务有自己的模型）
	// if err := db.AutoMigrate(&aftersalesmodel.ReturnRequest{}, &aftersalesmodel.RefundRequest{}); err != nil {
	// 	zap.S().Fatalf("failed to migrate database: %v", err);
	// }

	// gRPC 客户端连接（如果售后服务需要调用其他服务）
	// dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// exampleClientConn, err := grpc.DialContext(dialCtx, cfg.Data.ExampleService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	// if err != nil {
	// 	zap.S().Fatalf("failed to connect to example service: %v", err);
	// }
	// defer exampleClientConn.Close()
	// exampleClient := examplev1.NewExampleClient(exampleClientConn)

	// 6. 依赖注入 (DI)
	dataRepo, cleanupData, err := data.NewData(db) // Assuming data.NewData takes db
	if err != nil {
		zap.S().Fatalf("failed to create data layer: %v", err)
	}
	defer cleanupData()

	aftersalesRepo := data.NewAftersalesRepo(dataRepo)
	aftersalesUsecase := biz.NewAftersalesUsecase(aftersalesRepo)
	aftersalesService := service.NewAftersalesService(aftersalesUsecase)

	// 7. 启动 gRPC 服务器
	grpcServer, grpcErrChan := aftersaleshandler.StartGRPCServer(aftersalesService, cfg.Server.GRPC.Addr, cfg.Server.GRPC.Port)
	if grpcServer == nil {
		zap.S().Fatalf("failed to start gRPC server: %v", <-grpcErrChan)
	}

	// 8. 启动 Gin HTTP 服务器
	ginServer, ginErrChan := aftersaleshandler.StartGinServer(aftersalesService, cfg.Server.HTTP.Addr, cfg.Server.HTTP.Port)
	if ginServer == nil {
		zap.S().Fatalf("failed to start Gin HTTP server: %v", <-ginErrChan)
	}

	// 9. 等待中断信号或服务器错误以实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down aftersales service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down aftersales service due to gRPC error...")
	case err := <-ginErrChan:
		zap.S().Errorf("Gin HTTP server error: %v", err)
		zap.S().Info("Shutting down aftersales service due to Gin HTTP error...")
	}

	// 优雅地关闭服务器
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ginServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("Gin HTTP server shutdown error: %v", err)
	}
}

// startGRPCServer 启动 gRPC 服务器

// startGinServer 启动 Gin HTTP 服务器

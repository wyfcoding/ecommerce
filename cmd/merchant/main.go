// Package main 商家服务启动入口
// 生成摘要：
// 1) 依赖注入
// 2) HTTP/gRPC 服务启动
// 3) 优雅关停
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	merchantv1 "github.com/wyfcoding/ecommerce/go-api/merchant/v1"
	"github.com/wyfcoding/ecommerce/internal/merchant/application"
	"github.com/wyfcoding/ecommerce/internal/merchant/infrastructure"
	"github.com/wyfcoding/ecommerce/internal/merchant/interfaces"
	"github.com/wyfcoding/pkg/messagequeue"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// noopEventPublisher 空操作事件发布者，用于开发环境
type noopEventPublisher struct{}

// Ensure noopEventPublisher implements messagequeue.EventPublisher
var _ messagequeue.EventPublisher = (*noopEventPublisher)(nil)

// Publish 发布事件（空操作）
func (p *noopEventPublisher) Publish(_ context.Context, _ string, _ string, _ any) error {
	return nil
}

// PublishInTx 在事务中发布事件（空操作）
func (p *noopEventPublisher) PublishInTx(_ context.Context, _ any, _ string, _ string, _ any) error {
	return nil
}

// Config 服务配置
type Config struct {
	HTTPPort    int
	GRPCPort    int
	MySQLDSN    string
	KafkaBroker string
	LogLevel    string
}

func main() {
	// 初始化日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// 加载配置（实际应从配置文件/环境变量加载）
	cfg := &Config{
		HTTPPort:    8080,
		GRPCPort:    9080,
		MySQLDSN:    "root:password@tcp(localhost:3306)/merchant?charset=utf8mb4&parseTime=True&loc=Local",
		KafkaBroker: "localhost:9092",
	}

	// 初始化数据库
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	// 自动迁移
	// db.AutoMigrate(
	// 	&domain.Merchant{},
	// 	&domain.BusinessLicense{},
	// 	&domain.BankAccount{},
	// 	&domain.Store{},
	// 	&domain.MerchantSettings{},
	// )

	// 初始化仓储
	merchantRepo := infrastructure.NewGormMerchantRepository(db)
	licenseRepo := infrastructure.NewGormBusinessLicenseRepository(db)
	bankAccountRepo := infrastructure.NewGormBankAccountRepository(db)
	storeRepo := infrastructure.NewGormStoreRepository(db)
	settingsRepo := infrastructure.NewGormMerchantSettingsRepository(db)

	// 初始化事件发布者（开发环境使用空操作实现，生产环境替换为 Kafka 实现）
	eventPublisher := &noopEventPublisher{}

	// 初始化应用层服务
	commandService := application.NewCommandService(
		merchantRepo,
		licenseRepo,
		bankAccountRepo,
		storeRepo,
		settingsRepo,
		eventPublisher,
		logger,
	)
	queryService := application.NewQueryService(
		merchantRepo,
		licenseRepo,
		bankAccountRepo,
		storeRepo,
		settingsRepo,
		logger,
	)

	// 初始化 HTTP Handler
	httpHandler := interfaces.NewHTTPHandler(commandService, queryService)
	grpcHandler := interfaces.NewGRPCHandler(commandService, queryService)

	// 创建 Gin 引擎
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 注册路由
	api := router.Group("/api/v1")
	httpHandler.RegisterRoutes(api)

	// 创建 HTTP 服务器
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer()
	merchantv1.RegisterMerchantServiceServer(grpcServer, grpcHandler)

	// 使用 errgroup 管理服务生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	// 启动 HTTP 服务
	g.Go(func() error {
		logger.Info("starting HTTP server", "port", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("HTTP server error: %w", err)
		}
		return nil
	})

	// 启动 gRPC 服务
	g.Go(func() error {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			return fmt.Errorf("failed to listen gRPC: %w", err)
		}
		logger.Info("starting gRPC server", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			return fmt.Errorf("gRPC server error: %w", err)
		}
		return nil
	})

	// 监听关闭信号
	g.Go(func() error {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigCh:
			logger.Info("received shutdown signal", "signal", sig)
			cancel()
		case <-ctx.Done():
		}

		// 优雅关停
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// 关闭 HTTP 服务
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown error", "error", err)
		}

		// 关闭 gRPC 服务
		grpcServer.GracefulStop()

		logger.Info("servers stopped gracefully")
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

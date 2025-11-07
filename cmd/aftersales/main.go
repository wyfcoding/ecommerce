package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/aftersales/handler"
	"ecommerce/internal/aftersales/repository"
	"ecommerce/internal/aftersales/service"
	"ecommerce/pkg/config"
	"ecommerce/pkg/database"
	"ecommerce/pkg/logging"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 初始化日志
	logger := logging.NewLogger(cfg.Log.Level, cfg.Log.Filename)
	defer logger.Sync()

	// 初始化数据库
	db, err := database.NewMySQL(cfg.MySQL)
	if err != nil {
		logger.Fatal("连接数据库失败", zap.Error(err))
	}

	// 初始化Redis
	redisClient, err := database.NewRedis(cfg.Redis)
	if err != nil {
		logger.Fatal("连接Redis失败", zap.Error(err))
	}

	// 初始化Repository
	repo := repository.NewAfterSalesRepo(db)

	// 初始化Service
	svc := service.NewAfterSalesService(repo, redisClient, logger)

	// 初始化Handler
	h := handler.NewAfterSalesHandler(svc, logger)

	// 初始化Gin
	r := gin.Default()
	
	// 注册路由
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 启动HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("启动服务器失败", zap.Error(err))
		}
	}()

	logger.Info("售后服务启动成功", zap.Int("port", cfg.Server.Port))

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("服务器强制关闭", zap.Error(err))
	}

	logger.Info("服务器已关闭")
}

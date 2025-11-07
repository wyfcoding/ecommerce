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

	"ecommerce/internal/warehouse/handler"
	"ecommerce/internal/warehouse/repository"
	"ecommerce/internal/warehouse/service"
	"ecommerce/pkg/config"
	"ecommerce/pkg/database"
	"ecommerce/pkg/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	logger := logging.NewLogger(cfg.Log.Level, cfg.Log.Filename)
	defer logger.Sync()

	db, err := database.NewMySQL(cfg.MySQL)
	if err != nil {
		logger.Fatal("连接数据库失败", zap.Error(err))
	}

	redisClient, err := database.NewRedis(cfg.Redis)
	if err != nil {
		logger.Fatal("连接Redis失败", zap.Error(err))
	}

	repo := repository.NewWarehouseRepo(db)
	svc := service.NewWarehouseService(repo, redisClient, logger)
	h := handler.NewWarehouseHandler(svc, logger)

	r := gin.Default()
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{Addr: ":8009", Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("启动服务器失败", zap.Error(err))
		}
	}()

	logger.Info("仓库管理服务启动成功", zap.Int("port", 8009))

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

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ecommerce/internal/wishlist/handler"
	"ecommerce/internal/wishlist/repository"
	"ecommerce/internal/wishlist/service"
)

func main() {
	// 1. 初始化配置和日志
	viper.SetConfigName("wishlist")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./configs")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %s", err)
	}
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 2. 初始化数据存储
	dsn := viper.GetString("data.database.dsn")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("连接数据库失败", zap.Error(err))
	}

	// 3. 初始化 gRPC 客户端 (此处省略)
	// ...

	// 4. 依赖注入
	wishlistRepo := repository.NewWishlistRepository(db)
	// 实际应传入 gRPC 客户端
	wishlistService := service.NewWishlistService(wishlistRepo, logger)
	wishlistHandler := handler.NewWishlistHandler(wishlistService, logger)

	// 5. 初始化 HTTP 引擎
	router := gin.Default()
	wishlistHandler.RegisterRoutes(router)

	// 6. 启动 HTTP 服务器
	httpAddr := fmt.Sprintf("%s:%d", viper.GetString("server.http.addr"), viper.GetInt("server.http.port"))
	srv := &http.Server{Addr: httpAddr, Handler: router}

	go func() {
		logger.Info("HTTP 服务器正在监听", zap.String("address", httpAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP 服务监听失败", zap.Error(err))
		}
	}()

	// 7. 优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("准备关闭 HTTP 服务器 ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("HTTP 服务器强制关闭", zap.Error(err))
	}

	logger.Info("HTTP 服务器已退出")
}
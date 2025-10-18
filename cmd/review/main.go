package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ecommerce/internal/review/handler"
	"ecommerce/internal/review/repository"
	"ecommerce/internal/review/service"
)

func main() {
	// 1. 初始化配置和日志
	viper.SetConfigName("review")
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
	reviewRepo := repository.NewReviewRepository(db)
	// 实际应传入 gRPC 客户端
	reviewService := service.NewReviewService(reviewRepo, logger)

	var wg sync.WaitGroup

	// 5. 启动 HTTP 服务器
	go func() {
		wg.Add(1)
		defer wg.Done()
		reviewHttpHandler := handler.NewReviewHandler(reviewService, logger)
		router := gin.Default()
		reviewHttpHandler.RegisterRoutes(router)

		httpAddr := fmt.Sprintf("%s:%d", viper.GetString("server.http.addr"), viper.GetInt("server.http.port"))
		srv := &http.Server{Addr: httpAddr, Handler: router}

		go func() {
			logger.Info("HTTP 服务器正在监听", zap.String("address", httpAddr))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("HTTP 服务监听失败", zap.Error(err))
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info("准备关闭 HTTP 服务器 ...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	// 6. 启动 gRPC 服务器
	go func() {
		wg.Add(1)
		defer wg.Done()
		grpcAddr := fmt.Sprintf("%s:%d", viper.GetString("server.grpc.addr"), viper.GetInt("server.grpc.port"))
		listener, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal("gRPC 监听失败", zap.Error(err))
		}
		grpcServer := grpc.NewServer()
		// 注册 gRPC 服务

		go func() {
			logger.Info("gRPC 服务器正在监听", zap.String("address", grpcAddr))
			grpcServer.Serve(listener)
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info("准备关闭 gRPC 服务器 ...")
		grpcServer.GracefulStop()
	}()

	wg.Wait()
	logger.Info("所有服务已退出")
}
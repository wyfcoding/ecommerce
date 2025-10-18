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

	"ecommerce/internal/aftersales/handler"
	"ecommerce/internal/aftersales/repository"
	"ecommerce/internal/aftersales/service"
)

func main() {
	// 1. 初始化配置和日志
	viper.SetConfigName("aftersales")
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

	// 3. 初始化 gRPC 客户端 (省略)
	// ...

	// 4. 依赖注入
	aftersalesRepo := repository.NewAftersalesRepository(db)
	aftersalesService := service.NewAftersalesService(aftersalesRepo, logger)

	var wg sync.WaitGroup

	// 5. 启动 HTTP 服务器
	go func() {
		wg.Add(1)
		defer wg.Done()
		aftersalesHttpHandler := handler.NewAftersalesHandler(aftersalesService, logger)
		router := gin.Default()
		aftersalesHttpHandler.RegisterRoutes(router)

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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	// 6. 启动 gRPC 服务器
	go func() {
		wg.Add(1)
		defer wg.Done()
		grpcAddr := fmt.Sprintf("%s:%d", viper.GetString("server.grpc.addr"), viper.GetInt("server.grpc.port"))
		listener, _ := net.Listen("tcp", grpcAddr)
		grpcServer := grpc.NewServer()
		// 注册 gRPC 服务

		go func() {
			logger.Info("gRPC 服务器正在监听", zap.String("address", grpcAddr))
			grpcServer.Serve(listener)
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		grpcServer.GracefulStop()
	}()

	wg.Wait()
	logger.Info("所有服务已退出")
}

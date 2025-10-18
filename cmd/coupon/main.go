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
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ecommerce/internal/coupon/handler"
	"ecommerce/internal/coupon/repository"
	"ecommerce/internal/coupon/service"
	// 伪代码
	// "ecommerce/internal/coupon/handler/grpc"
)

func main() {
	// 1. 初始化配置和日志
	viper.SetConfigName("coupon")
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
	redisAddr := viper.GetString("data.redis.addr")
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		logger.Fatal("连接 Redis 失败", zap.Error(err))
	}

	// 3. 依赖注入
	couponRepo := repository.NewCouponRepository(db)
	couponService := service.NewCouponService(couponRepo, logger)

	// 使用 WaitGroup 等待所有服务器关闭
	var wg sync.WaitGroup

	// 4. 启动 HTTP 服务器
	go func() {
		wg.Add(1)
		defer wg.Done()
		couponHttpHandler := handler.NewCouponHandler(couponService, logger)
		router := gin.Default()
		couponHttpHandler.RegisterRoutes(router)

		httpAddr := fmt.Sprintf("%s:%d", viper.GetString("server.http.addr"), viper.GetInt("server.http.port"))
		srv := &http.Server{Addr: httpAddr, Handler: router}

		go func() {
			logger.Info("HTTP 服务器正在监听", zap.String("address", httpAddr))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("HTTP 服务监听失败", zap.Error(err))
			}
		}()

		// 等待关闭信号
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
	}()

	// 5. 启动 gRPC 服务器
	go func() {
		wg.Add(1)
		defer wg.Done()
		grpcAddr := fmt.Sprintf("%s:%d", viper.GetString("server.grpc.addr"), viper.GetInt("server.grpc.port"))
		listener, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal("gRPC 监听失败", zap.Error(err))
		}
		grpcServer := grpc.NewServer()
		// couponGrpcHandler := grpc_handler.NewCouponServer(couponService)
		// pb.RegisterCouponServiceServer(grpcServer, couponGrpcHandler)

		go func() {
			logger.Info("gRPC 服务器正在监听", zap.String("address", grpcAddr))
			if err := grpcServer.Serve(listener); err != nil {
				logger.Error("gRPC 服务启动失败", zap.Error(err))
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info("准备关闭 gRPC 服务器 ...")
		grpcServer.GracefulStop()
		logger.Info("gRPC 服务器已退出")
	}()

	wg.Wait()
}

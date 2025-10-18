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
	"google.golang.org/grpc"

	"ecommerce/internal/admin/handler"
	"ecommerce/internal/admin/service"
)

func main() {
	// 1. 初始化配置和日志
	viper.SetConfigName("admin")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./configs")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %s", err)
	}
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 2. 初始化所有下游服务的 gRPC 客户端
	// 这是一个示例，实际项目中需要为每个服务都建立连接
	// userSvcAddr := viper.GetString("dependencies.user_service_grpc_addr")
	// userConn, err := grpc.Dial(userSvcAddr, grpc.WithInsecure())
	// if err != nil { logger.Fatal("连接用户服务失败", zap.Error(err)) }
	// defer userConn.Close()
	// userClient := userpb.NewUserServiceClient(userConn)
	// ... 对其他所有服务执行相同操作

	// 3. 依赖注入
	// 实际应传入所有 gRPC 客户端
	adminService := service.NewAdminService(logger)
	adminHandler := handler.NewAdminHandler(adminService, logger)

	// 4. 初始化 HTTP 引擎
	router := gin.Default()
	// TODO: 在此应用管理员认证中间件
	adminHandler.RegisterRoutes(router)

	// 5. 启动 HTTP 服务器
	httpAddr := fmt.Sprintf("%s:%d", viper.GetString("server.http.addr"), viper.GetInt("server.http.port"))
	srv := &http.Server{Addr: httpAddr, Handler: router}

	go func() {
		logger.Info("Admin HTTP 服务器正在监听", zap.String("address", httpAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP 服务监听失败", zap.Error(err))
		}
	}()

	// 6. 优雅停机
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
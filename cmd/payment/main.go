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

	"ecommerce/internal/payment/handler"
	"ecommerce/internal/payment/repository"
	"ecommerce/internal/payment/service"
	// 伪代码: "ecommerce/pkg/middleware"
)

func main() {
	// 1. 初始化配置
	vm := viper.New()
	vm.SetConfigName("payment")
	vm.SetConfigType("toml")
	vm.AddConfigPath("./configs")
	if err := vm.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %s \n", err)
	}

	// 2. 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("初始化 Zap logger 失败: %v", err)
	}
	defer logger.Sync()

	// 3. 初始化数据存储
	dsn := vm.GetString("data.database.dsn")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("连接数据库失败", zap.Error(err))
	}

	// 4. 初始化第三方 SDK、gRPC 客户端、消息队列等 (此处省略)
	// ...

	// 5. 依赖注入
	paymentRepo := repository.NewPaymentRepository(db)
	// 实际应传入支付网关客户端和 MQ 生产者
	paymentService := service.NewPaymentService(paymentRepo, logger)
	paymentHandler := handler.NewPaymentHandler(paymentService, logger)

	// 6. 初始化 HTTP 引擎 (Gin)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 注册路由
	// 注意：部分路由需要认证中间件，部分（webhook）则不需要
	paymentHandler.RegisterRoutes(router)

	// 7. 启动 HTTP 服务器
	httpAddr := fmt.Sprintf("%s:%d", vm.GetString("server.http.addr"), vm.GetInt("server.http.port"))
	srv := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP 服务监听失败", zap.Error(err))
		}
	}()

	// 8. 优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-
	logger.Info("准备关闭服务器 ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("服务器强制关闭: ", zap.Error(err))
	}

	logger.Info("服务器已退出")
}
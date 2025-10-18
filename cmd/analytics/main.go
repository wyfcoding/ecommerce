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
	"gorm.io/driver/clickhouse" // 假设使用 GORM 的 ClickHouse 驱动
	"gorm.io/gorm"

	"ecommerce/internal/analytics/handler"
	"ecommerce/internal/analytics/repository"
	"ecommerce/internal/analytics/service"
	// 伪代码
	// "ecommerce/pkg/mq/consumer"
)

func main() {
	// 1. 初始化配置和日志
	viper.SetConfigName("analytics")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./configs")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %s", err)
	}
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 2. 初始化数据存储 (ClickHouse)
	dsn := viper.GetString("data.database.dsn")
	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("连接 ClickHouse 失败", zap.Error(err))
	}

	// 3. 依赖注入
	analyticsRepo := repository.NewAnalyticsRepository(db)
	analyticsService := service.NewAnalyticsService(analyticsRepo, logger)

	// 4. 启动消息队列消费者
	go func() {
		// mqConsumer, _ := consumer.New(...)
		// handler := func(eventType string, body []byte) error {
		// 	 return analyticsService.ProcessEvent(context.Background(), eventType, body)
		// }
		// mqConsumer.StartConsuming("analytics.events", "all_events_fanout", "#", handler)
		logger.Info("消息队列消费者已启动 (模拟)")
	}()

	// 5. 启动 HTTP 服务器
	analyticsHttpHandler := handler.NewAnalyticsHandler(analyticsService, logger)
	router := gin.Default()
	analyticsHttpHandler.RegisterRoutes(router)

	httpAddr := fmt.Sprintf("%s:%d", viper.GetString("server.http.addr"), viper.GetInt("server.http.port"))
	srv := &http.Server{Addr: httpAddr, Handler: router}

	go func() {
		logger.Info("Analytics HTTP 服务器正在监听", zap.String("address", httpAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP 服务监听失败", zap.Error(err))
		}
	}()

	// 6. 优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("准备关闭服务 ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	// mqConsumer.Stop()

	logger.Info("服务已退出")
}

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ecommerce/internal/notification/repository"
	"ecommerce/internal/notification/service"
	// 伪代码: 模拟 MQ 消费者
	// "ecommerce/pkg/mq/consumer"
)

func main() {
	// 1. 初始化配置
	vper.SetConfigName("notification")
	vper.SetConfigType("toml")
	vper.AddConfigPath("./configs")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %s \n", err)
	}

	// 2. 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("初始化 Zap logger 失败: %v", err)
	}
	defer logger.Sync()

	// 3. 初始化数据存储
	dsn := viper.GetString("data.database.dsn")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("连接数据库失败", zap.Error(err))
	}

	// 4. 初始化第三方 SDK (Email, SMS, Push 客户端)
	// ...

	// 5. 依赖注入
	notiRepo := repository.NewNotificationRepository(db)
	// 实际应传入各通知渠道的客户端
	notiService := service.NewNotificationService(notiRepo, logger)

	// 6. 初始化并启动消息队列消费者
	// 这是此服务的核心驱动力
	// mqConsumer, err := consumer.NewRabbitMQConsumer(viper.GetString("mq.url"))
	// if err != nil {
	// 	 logger.Fatal("连接消息队列失败", zap.Error(err))
	// }

	// 获取要消费的队列列表
	// consumerQueues := viper.Get("mq.consumer_queues")
	// for _, q := range consumerQueues...
	// 	 go func(queueName, exchangeName, routingKey string) {
	// 		 handler := func(eventType string, body []byte) error {
	// 			 return notiService.ProcessEvent(context.Background(), eventType, body)
	// 		 }
	// 		 if err := mqConsumer.StartConsuming(queueName, exchangeName, routingKey, handler); err != nil {
	// 			 logger.Error("启动消费者失败", zap.String("queue", queueName), zap.Error(err))
	// 		 }
	// 	}(q...)
	// }
	logger.Info("消息队列消费者已启动 (模拟)")

	// 7. (可选) 启动 gRPC 服务器
	// ...

	// 8. 优雅停机
	// 程序将持续运行，直到接收到终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("准备关闭服务 ...")

	// 关闭消费者和服务器
	// mqConsumer.Stop()
	// grpcServer.GracefulStop()

	logger.Info("服务已退出")
}
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ecommerce/internal/logistics/repository"
	"ecommerce/internal/logistics/service"
	// 伪代码: 模拟 gRPC handler 和 MQ 消费者
	// "ecommerce/internal/logistics/handler/grpc"
	// "ecommerce/pkg/mq/consumer"
)

func main() {
	// 1. 初始化配置
	vm := viper.New()
	vm.SetConfigName("logistics")
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

	// 4. 初始化第三方 SDK (EasyPost client)
	// ...

	// 5. 依赖注入
	logisticsRepo := repository.NewLogisticsRepository(db)
	// 实际应传入 EasyPost client 和 MQ 生产者
	logisticsService := service.NewLogisticsService(logisticsRepo, logger)

	// 6. 启动消息队列消费者
	go func() {
		// mqConsumer, _ := consumer.New(...)
		// mqConsumer.RegisterHandler("order.processing", logisticsService.HandleOrderProcessing)
		// mqConsumer.Start()
		logger.Info("消息队列消费者已启动 (模拟)")
	}()

	// 7. 启动 gRPC 服务器
	grpcAddr := fmt.Sprintf("%s:%d", vm.GetString("server.grpc.addr"), vm.GetInt("server.grpc.port"))
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("gRPC 监听失败", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	// 伪代码: 注册 gRPC 服务
	// logistics_grpc_handler := grpc_handler.NewLogisticsServer(logisticsService)
	// pb.RegisterLogisticsServiceServer(grpcServer, logistics_grpc_handler)
	logger.Info("gRPC 服务器正在监听", zap.String("address", grpcAddr))

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error("gRPC 服务启动失败", zap.Error(err))
		}
	}()

	// 8. 优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-
	logger.Info("准备关闭服务 ...")

	grpcServer.GracefulStop()
	// mqConsumer.Stop()

	logger.Info("服务已退出")
}
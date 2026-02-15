// 生成摘要：
// - 实现 Fulfillment 服务启动入口，使用标准化 app.Builder 引导
// - 完成依赖注入：MySQL -> GORM Repository -> Command/Query Services -> gRPC/HTTP Handlers
// - 注册 Kafka 消费者监听 order.confirmed 事件，实现自动履约逻辑
// - 集成 Prometheus 指标、Jaeger 追踪和结构化日志

package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	pb "github.com/wyfcoding/ecommerce/go-api/fulfillment/v1"
	orderv1 "github.com/wyfcoding/ecommerce/go-api/order/v1"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/application"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/infrastructure/mq"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/infrastructure/persistence/mysql"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/interfaces"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/interfaces/consumer"
	fulfillmentgrpc "github.com/wyfcoding/ecommerce/internal/fulfillment/interfaces/grpc"
	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// BootstrapName 服务唯一标识
const BootstrapName = "fulfillment"

// Config 扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	Fulfillment      struct {
		AutoAssignPicking     bool   `mapstructure:"auto_assign_picking"`
		PriorityStrategy      string `mapstructure:"priority_strategy"`
		PickingTimeoutMinutes int    `mapstructure:"picking_timeout_minutes"`
	} `mapstructure:"fulfillment"`
}

// AppContext 应用上下文
type AppContext struct {
	Config         *Config
	CommandService *application.CommandService
	QueryService   *application.QueryService
	OrderClient    orderv1.OrderServiceClient
}

func main() {
	if err := app.NewBuilder[*Config, *AppContext](BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		Build().
		Run(); err != nil {
		slog.Error("fulfillment service bootstrap failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, ctx *AppContext) {
	pb.RegisterFulfillmentServiceServer(s, fulfillmentgrpc.NewHandler(ctx.CommandService, ctx.QueryService))
}

func registerGin(r *gin.Engine, ctx *AppContext) {
	api := r.Group("/api/v1")
	h := interfaces.NewHTTPHandler(ctx.CommandService, ctx.QueryService)
	h.RegisterRoutes(api)
}

func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	c := cfg
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	// 1. 初始化数据库
	db, err := database.NewDB(c.Data.Database, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("database init error: %w", err)
	}

	// 2. 初始化 Kafka 生产者
	producer := kafka.NewProducer(&c.MessageQueue.Kafka, logger, m)
	eventPublisher := mq.NewKafkaEventPublisher(producer, logger.Logger)

	// 3. 构建仓储与应用服务
	fulfillmentRepo := mysql.NewFulfillmentRepository(db)
	cmdSvc := application.NewCommandService(fulfillmentRepo, eventPublisher, logger.Logger)
	querySvc := application.NewQueryService(fulfillmentRepo, logger.Logger)

	// 4. 初始化下游 gRPC 客户端 (Order Service)
	orderAddr := c.GetGRPCAddr("order")
	orderConn, err := grpc.Dial(orderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		bootLog.Warn("failed to connect order service", "addr", orderAddr, "error", err)
	}
	var orderClient orderv1.OrderServiceClient
	if orderConn != nil {
		orderClient = orderv1.NewOrderServiceClient(orderConn)
	}

	// 5. 注册 Kafka 消费者
	orderConfirmedConsumer := kafka.NewConsumer(&configpkg.KafkaConfig{
		Brokers: c.MessageQueue.Kafka.Brokers,
		Topic:   "order.confirmed",
		GroupID: "fulfillment-order-confirmed-group",
	}, logger, m)

	handler := consumer.NewOrderConfirmedHandler(cmdSvc, querySvc, orderClient, logger.Logger)

	// 在后台启动消费者
	stopConsumer := make(chan struct{})
	go func() {
		bootLog.Info("starting order confirmed consumer")
		if err := orderConfirmedConsumer.Consume(context.Background(), handler.Handle); err != nil {
			bootLog.Error("order confirmed consumer stopped", "error", err)
		}
	}()

	cleanup := func() {
		bootLog.Info("shutting down fulfillment service...")
		if sqlDB, err := db.RawDB().DB(); err == nil && sqlDB != nil {
			sqlDB.Close()
		}
		if producer != nil {
			producer.Close()
		}
		if orderConfirmedConsumer != nil {
			orderConfirmedConsumer.Close()
		}
		if orderConn != nil {
			orderConn.Close()
		}
		close(stopConsumer)
	}

	return &AppContext{
		Config:         c,
		CommandService: cmdSvc,
		QueryService:   querySvc,
		OrderClient:    orderClient,
	}, cleanup, nil
}

// Package main 履约服务启动入口
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	orderv1 "github.com/wyfcoding/ecommerce/go-api/order/v1"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/application"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/infrastructure"
	fulfillmentconsumer "github.com/wyfcoding/ecommerce/internal/fulfillment/interfaces/consumer"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/interfaces"
	pkgconfig "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/metrics"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type kafkaEventPublisher struct {
	producer *kafka.Producer
	logger   *slog.Logger
}

var _ messagequeue.EventPublisher = (*kafkaEventPublisher)(nil)

func (p *kafkaEventPublisher) Publish(ctx context.Context, topic, key string, event any) error {
	if p == nil || p.producer == nil {
		return nil
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if err := p.producer.PublishToTopic(ctx, topic, []byte(key), payload); err != nil {
		if p.logger != nil {
			p.logger.Error("failed to publish fulfillment event", "topic", topic, "error", err)
		}
		return err
	}
	return nil
}

func (p *kafkaEventPublisher) PublishInTx(ctx context.Context, _ any, topic, key string, event any) error {
	return p.Publish(ctx, topic, key, event)
}

// Config 服务配置
type Config struct {
	HTTPPort      int
	GRPCPort      int
	MySQLDSN      string
	KafkaBroker   string
	OrderGRPCAddr string
	LogLevel      string
}

func main() {
	// 初始化日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// 加载配置
	cfg := &Config{
		HTTPPort:      8081,
		GRPCPort:      9081,
		MySQLDSN:      "root:password@tcp(localhost:3306)/fulfillment?charset=utf8mb4&parseTime=True&loc=Local",
		KafkaBroker:   "localhost:9092",
		OrderGRPCAddr: "localhost:50051",
	}

	// 初始化数据库
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	// 初始化仓储
	fulfillmentRepo := infrastructure.NewGormFulfillmentRepository(db)

	// 初始化消息发布
	brokers := []string{cfg.KafkaBroker}
	if cfg.KafkaBroker == "" {
		brokers = []string{"localhost:9092"}
	}
	infraLogger := logging.NewLogger("fulfillment", "bootstrap")
	metricsImpl := metrics.NewMetrics("fulfillment")
	producer := kafka.NewProducer(&pkgconfig.KafkaConfig{
		Brokers: brokers,
	}, infraLogger, metricsImpl)
	eventPublisher := &kafkaEventPublisher{producer: producer, logger: logger}

	// 初始化应用层服务
	commandService := application.NewCommandService(
		fulfillmentRepo,
		eventPublisher,
		logger,
	)
	queryService := application.NewQueryService(
		fulfillmentRepo,
		logger,
	)

	// 初始化订单客户端（用于自动创建履约）
	orderConn, err := grpc.Dial(cfg.OrderGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to connect order service", "addr", cfg.OrderGRPCAddr, "error", err)
	}
	var orderClient orderv1.OrderServiceClient
	if orderConn != nil {
		orderClient = orderv1.NewOrderServiceClient(orderConn)
	}

	var orderConfirmedConsumer *kafka.Consumer
	if orderClient != nil {
		orderConfirmedConsumer = kafka.NewConsumer(&pkgconfig.KafkaConfig{
			Brokers: brokers,
			Topic:   "order.confirmed",
			GroupID: "fulfillment-order-confirmed-group",
		}, infraLogger, metricsImpl)
		orderConfirmedConsumer.Start(context.Background(), 3, fulfillmentconsumer.NewOrderConfirmedHandler(
			commandService,
			queryService,
			orderClient,
			logger,
		).Handle)
	}

	// 初始化 HTTP Handler
	httpHandler := interfaces.NewHTTPHandler(commandService, queryService)

	// 创建 Gin 引擎
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 注册路由
	api := router.Group("/api/v1")
	httpHandler.RegisterRoutes(api)

	// 创建 HTTP 服务器
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer()

	// 服务生命周期管理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	// 启动 HTTP 服务
	g.Go(func() error {
		logger.Info("starting HTTP server", "port", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("HTTP server error: %w", err)
		}
		return nil
	})

	// 启动 gRPC 服务
	g.Go(func() error {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			return fmt.Errorf("failed to listen gRPC: %w", err)
		}
		logger.Info("starting gRPC server", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			return fmt.Errorf("gRPC server error: %w", err)
		}
		return nil
	})

	// 监听关闭信号
	g.Go(func() error {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigCh:
			logger.Info("received shutdown signal", "signal", sig)
			cancel()
		case <-ctx.Done():
		}

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown error", "error", err)
		}

		grpcServer.GracefulStop()
		if orderConfirmedConsumer != nil {
			_ = orderConfirmedConsumer.Close()
		}
		if producer != nil {
			_ = producer.Close()
		}
		if orderConn != nil {
			_ = orderConn.Close()
		}

		logger.Info("servers stopped gracefully")
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

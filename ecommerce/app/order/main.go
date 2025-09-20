package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	orderV1 "ecommerce/ecommerce/api/order/v1"
	"ecommerce/ecommerce/app/order/internal/biz"
	"ecommerce/ecommerce/app/order/internal/data"
	"ecommerce/ecommerce/app/order/internal/service"
	"ecommerce/ecommerce/pkg/snowflake"

	"git.example.com/your_org/genesis/pkg/metrics"
	"git.example.com/your_org/genesis/pkg/tracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

const (
	orderServicePort        = ":9003"
	productServiceAddr      = "localhost:9001"
	cartServiceAddr         = "localhost:9002"
	orderServiceMetricsPort = "9093" // 为 order-service 分配一个指标端口
	jaegerEndpoint          = "http://localhost:14268/api/traces"
)

func main() {
	// --- 1. 初始化依赖 ---

	// --- 1. 初始化和暴露 Metrics ---
	metrics.InitMetrics()
	go metrics.ExposeHttp(orderServiceMetricsPort)

	// --- 新增：初始化和关闭 Tracer ---
	tp, err := tracing.InitTracer("order-service", jaegerEndpoint)
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// 初始化 Snowflake (machineID 应唯一)
	if err := snowflake.Init("2025-01-01", 3); err != nil {
		log.Fatalf("failed to init snowflake: %v", err)
	}

	// 初始化数据库连接 (连接到 genesis_order 数据库)
	dsn := "root:your_password@tcp(127.0.0.1:3306)/genesis_order?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 在 gRPC Client 端应用拦截器链
	productSvcConn, err := grpc.Dial(productServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor( // 使用拦截器链
			grpc_prometheus.UnaryClientInterceptor,
			otelgrpc.UnaryClientInterceptor(), // 添加 Tracing 客户端拦截器
		),
	)
	if err != nil {
		log.Fatalf("failed to connect to product service: %v", err)
	}
	defer productSvcConn.Close()

	// 初始化到 Cart Service 的 gRPC 客户端连接
	cartSvcConn, err := grpc.Dial(cartServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			grpc_prometheus.UnaryClientInterceptor,
			otelgrpc.UnaryClientInterceptor(), // 添加 Tracing 客户端拦截器
		),
	)
	if err != nil {
		log.Fatalf("failed to connect to cart service: %v", err)
	}
	defer cartSvcConn.Close()

	// --- 2. 依赖注入 (DI) ---
	productGreeter := data.NewProductGreeter(productSvcConn)
	cartGreeter := data.NewCartGreeter(cartSvcConn)
	orderRepo := data.NewOrderRepo(db)

	orderUsecase := biz.NewOrderUsecase(orderRepo, productGreeter, cartGreeter)
	orderService := service.NewOrderService(orderUsecase)

	// --- 3. 启动 gRPC 服务器 ---
	lis, err := net.Listen("tcp", orderServicePort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// 在 gRPC Server 端应用拦截器链
	s := grpc.NewServer(
		grpc.WithChainUnaryInterceptor( // 使用拦截器链
			grpc_prometheus.UnaryServerInterceptor,
			otelgrpc.UnaryServerInterceptor(), // 添加 Tracing 服务端拦截器
		),
	)

	// 为 gRPC Server 注册 prometheus 指标
	grpc_prometheus.Register(s)
	orderV1.RegisterOrderServer(s, orderService)

	log.Printf("gRPC server (order-service) listening at %s", orderServicePort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

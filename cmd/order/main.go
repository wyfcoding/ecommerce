package main

import (
	"fmt"
	"time"

	"ecommerce/api/order/v1"
	"ecommerce/internal/order/biz"
	"ecommerce/internal/order/data"
	"ecommerce/internal/order/handler"
	"ecommerce/internal/order/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/metrics"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Config is the service-specific configuration structure.
type Config struct {
	configpkg.Config
	Services struct {
		ProductService struct {
			Addr string `toml:"addr"`
		}
		CartService struct {
			Addr string `toml:"addr"`
		}
		PaymentService struct {
			Addr string `toml:"addr"`
		}
	} `toml:"services"`
	Data struct {
		configpkg.DataConfig
		Database struct {
			configpkg.DatabaseConfig
			LogLevel      gormlogger.LogLevel `toml:"log_level"`
			SlowThreshold time.Duration     `toml:"slow_threshold"`
		} `toml:"database"`
	} `toml:"data"`
}

func main() {
	app.NewBuilder("order").
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithMetrics("9092").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	v1.RegisterOrderServer(s, srv.(v1.OrderServer))
}

func registerGin(e *gin.Engine, srv interface{}) {
	orderHandler := handler.NewOrderHandler(srv.(*service.OrderService))
	// e.g., e.POST("/v1/orders", orderHandler.CreateOrder)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	// --- Downstream gRPC clients ---
	productServiceConn, err := grpc.Dial(config.Services.ProductService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to product service: %w", err)
	}

	cartServiceConn, err := grpc.Dial(config.Services.CartService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		productServiceConn.Close()
		return nil, nil, fmt.Errorf("failed to connect to cart service: %w", err)
	}

	paymentServiceConn, err := grpc.Dial(config.Services.PaymentService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		productServiceConn.Close()
		cartServiceConn.Close()
		return nil, nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}

	// --- Data layer ---
	db, err := gorm.Open(mysql.Open(config.Data.Database.DSN), &gorm.Config{
		Logger: gormlogger.New(zap.L(), config.Data.Database.LogLevel, config.Data.Database.SlowThreshold),
	})
	if err != nil {
		productServiceConn.Close()
		cartServiceConn.Close()
		paymentServiceConn.Close()
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(&data.Order{}, &data.OrderItem{}); err != nil {
		return nil, nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// --- DI (Data -> Biz -> Service) ---
	dataInstance := data.NewData(db)
	orderRepo := data.NewOrderRepo(dataInstance)
	productClient := data.NewProductClient(productServiceConn)
	cartClient := data.NewCartClient(cartServiceConn)
	paymentClient := data.NewPaymentClient(paymentServiceConn)
	transaction := data.NewTransaction(dataInstance)
	orderUsecase := biz.NewOrderUsecase(transaction, orderRepo, productClient, cartClient, paymentClient)
	orderService := service.NewOrderService(orderUsecase)

	cleanup := func() {
		sqlDB.Close()
		productServiceConn.Close()
		cartServiceConn.Close()
		paymentServiceConn.Close()
	}

	return orderService, cleanup, nil
}

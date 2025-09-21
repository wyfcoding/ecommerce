package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "ecommerce/api/order/v1"
	"ecommerce/internal/order/biz"
	"ecommerce/internal/order/data"
	"ecommerce/internal/order/service"
	"ecommerce/pkg/logging"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/BurntSushi/toml"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	Server struct {
		GRPC struct {
			Addr    string `toml:"addr"`
			Port    int    `toml:"port"`
			Timeout string `toml:"timeout"`
		} `toml:"grpc"`
	} `toml:"server"`
	Data struct {
		Database struct {
			DSN string `toml:"dsn"`
		} `toml:"database"`
		ProductService struct {
			Addr string `toml:"addr"`
		} `toml:"product_service"`
		CartService struct {
			Addr string `toml:"addr:"`
		} `toml:"cart_service"`
	} `toml:"data"`
	Log struct {
		Level  string `toml:"level"`
		Format string `toml:"format"`
		Output string `toml:"output"`
	} `toml:"log"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/order.toml", "config file path")
	flag.Parse()

	// 1. 加载配置
	config, err := loadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger := logging.NewLogger(config.Log.Level, config.Log.Format, config.Log.Output)
	zap.ReplaceGlobals(logger)

	// 3. 初始化依赖
	// gRPC 客户端连接
	dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	productServiceConn, err := grpc.DialContext(dialCtx, config.Data.ProductService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		zap.S().Fatalf("failed to connect to product service: %v", err)
	}
	defer productServiceConn.Close()

	cartServiceConn, err := grpc.Dial(config.Data.CartService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Fatalf("failed to connect to cart service: %v", err)
	}
	defer cartServiceConn.Close()

	// 数据库连接
	db, err := gorm.Open(mysql.Open(config.Data.Database.DSN), &gorm.Config{})
	if err != nil {
		zap.S().Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移数据库表结构
	if err := db.AutoMigrate(&data.Order{}, &data.OrderItem{}); err != nil {
		zap.S().Fatalf("failed to migrate database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.S().Fatalf("failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// 4. 依赖注入 (DI)
	dataInstance := data.NewData(db)
	orderRepo := data.NewOrderRepo(dataInstance)
	productClient := data.NewProductClient(productServiceConn)
	cartClient := data.NewCartClient(cartServiceConn)
	transaction := data.NewTransaction(dataInstance)
	orderUsecase := biz.NewOrderUsecase(transaction, orderRepo, productClient, cartClient)
	orderService := service.NewOrderService(orderUsecase)

	// 5. 启动 gRPC 服务器
	grpcServer, errChan := startGRPCServer(orderService, config.Server.GRPC.Addr, config.Server.GRPC.Port)
	if grpcServer == nil {
		zap.S().Fatalf("failed to start gRPC server: %v", <-errChan)
	}

	// 6. 等待中断信号或服务器错误以实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down order service...")
	case err := <-errChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down order service due to error...")
	}

	grpcServer.GracefulStop()
}

func loadConfig(path string) (*Config, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	return &c, toml.Unmarshal(f, &c)
}

func startGRPCServer(orderService *service.OrderService, addr string, port int) (*grpc.Server, chan error) {
	errChan := make(chan error, 1)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %w", err)
		return nil, errChan
	}
	s := grpc.NewServer()
	v1.RegisterOrderServer(s, orderService)

	zap.S().Infof("gRPC server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(errChan)
	}()
	return s, errChan
}

package main

import (
	"context"
	"ecommerce/internal/flashsale/biz"
	"ecommerce/internal/flashsale/data"
	"ecommerce/internal/flashsale/service"
	"fmt"
	"log"
	"time"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	orderv1 "ecommerce/api/order/v1"
	flashsalehandler "ecommerce/internal/flashsale/handler"
)

// Config holds the application configuration.
type Config struct {
	Service  ServiceConfig  `toml:"service"`
	Database DatabaseConfig `toml:"database"`
	Redis    RedisConfig    `toml:"redis"`
	Data     DataConfig     `toml:"data"`
}

type ServiceConfig struct {
	Port string `toml:"port"`
}

type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBName   string `toml:"dbname"`
}

type RedisConfig struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
}

type DataConfig struct {
	OrderService struct {
		Addr string `toml:"addr"`
	} `toml:"order_service"`
}

func main() {
	// ======== 1. Initialize Dependencies (e.g., Config, Logger, DB) ========

	// Load configuration from TOML file
	var conf Config
	if _, err := toml.DecodeFile("../../configs/flashsale.toml", &conf); err != nil {
		log.Fatalf("failed to load config file: %v", err)
	}

	// Construct MySQL DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Database.User,
		conf.Database.Password,
		conf.Database.Host,
		conf.Database.Port,
		conf.Database.DBName,
	)

	// Connect to MySQL database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&data.FlashSaleEvent{}, &data.FlashSaleProduct{} /*, &data.FlashSaleOrder{} */)
	if err != nil {
		log.Fatalf("failed to migrate schema: %v", err)
	}

	log.Println("Successfully connected to database and migrated schema.")

	// gRPC Client for Order Service
	dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	orderServiceConn, err := grpc.DialContext(dialCtx, conf.Data.OrderService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to order service: %v", err)
	}
	defer orderServiceConn.Close()
	orderClient := orderv1.NewOrderClient(orderServiceConn)

	// ======== 2. Wire up the application layers (Dependency Injection) ========

	dataRepo, cleanup, err := data.NewData(db, conf.Redis.Addr, conf.Redis.Password)
	if err != nil {
		log.Fatalf("failed to create data layer: %v", err)
	}
	defer cleanup()

	flashSaleRepo := data.NewFlashSaleRepo(dataRepo)
	distributedLocker := data.NewDistributedLocker(dataRepo)
	orderServiceClient := data.NewOrderServiceClient(orderClient)
	flashSaleUsecase := biz.NewFlashSaleUsecase(flashSaleRepo, distributedLocker, orderServiceClient)
	flashSaleService := service.NewFlashSaleService(flashSaleUsecase)

	log.Println("Application layers wired successfully.")

	// ======== 3. Start the Server (e.g., HTTP, gRPC) ========

	flashsalehandler.StartHTTPServer(flashSaleService, conf.Service.Port)

	_ = flashSaleService
}

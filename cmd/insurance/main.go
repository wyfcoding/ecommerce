package main

import (
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	v1 "github.com/wyfcoding/ecommerce/go-api/insurance/v1"
	"github.com/wyfcoding/ecommerce/internal/insurance/application"
	"github.com/wyfcoding/ecommerce/internal/insurance/infrastructure/messaging"
	persistence_mysql "github.com/wyfcoding/ecommerce/internal/insurance/infrastructure/persistence/mysql"
	grpc_server "github.com/wyfcoding/ecommerce/internal/insurance/interfaces/grpc"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Config (Simplified)
	dsn := os.Getenv("To be replaced by config loader")
	if dsn == "" {
		dsn = "root:password@tcp(127.0.0.1:3306)/ecommerce_insurance?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// 3. Database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto Migrate
	err = db.AutoMigrate(persistence_mysql.GetModels()...)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// 4. Layers
	repo := persistence_mysql.NewInsuranceRepository(db)

	// Mock Event Publisher for skeleton (replace with real Kafka publisher)
	eventPub := &messaging.NoOpEventPublisher{}

	app := application.NewInsuranceService(repo, eventPub, logger)
	svc := grpc_server.NewServer(app)

	// 5. Server
	lis, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	v1.RegisterInsuranceServiceServer(s, svc)

	go func() {
		logger.Info("server started", "addr", ":9091")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")
	s.GracefulStop()
}

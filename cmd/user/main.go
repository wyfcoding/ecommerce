package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/application"
	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/ecommerce/internal/user/infrastructure/messaging"
	"github.com/wyfcoding/ecommerce/internal/user/infrastructure/persistence"
	grpcInterface "github.com/wyfcoding/ecommerce/internal/user/interfaces/grpc"
	httpInterface "github.com/wyfcoding/ecommerce/internal/user/interfaces/http"

	"log/slog"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	pb "github.com/wyfcoding/ecommerce/goapi/user/v1"
	// Assuming pkg/config is used but locally we are using specific structs if needed.
	// For this task, I will mock/simplify config loading or use what's there.
	// Since I overwrote main.go previously with simplified version, I will stick to it but add Kafka generic setup.
)

func main() {
	// 1. 初始化 Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. 初始化数据库 (Simplified)
	dsn := "root:root@tcp(127.0.0.1:3306)/ecommerce_user?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&domain.User{}, &domain.Address{})

	// 3. 初始化 Kafka Publisher
	brokers := []string{"localhost:9092"}
	publisher := messaging.NewKafkaPublisher(brokers, logger)
	defer publisher.Close()

	// 4. 初始化 Repositories
	userRepo := persistence.NewUserRepository(db)
	addressRepo := persistence.NewAddressRepository(db)

	// 5. 初始化 Services
	// topic: "user-events" from config
	cmdService := application.NewUserCommandService(
		userRepo,
		addressRepo,
		publisher,
		"user-events",
		"secret",
		"ecommerce",
		24*time.Hour,
		nil,
		logger,
	)
	queryService := application.NewUserQuery(userRepo, addressRepo, nil, logger)

	// 6. 初始化 Handlers
	httpHandler := httpInterface.NewUserHandler(cmdService, queryService)
	grpcHandler := grpcInterface.NewGrpcHandler(cmdService, queryService)

	// 7. 启动 gRPC Server
	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, grpcHandler)
	go func() {
		logger.Info("gRPC server listening at :9001")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// 8. 启动 HTTP Server
	r := gin.Default()
	httpHandler.RegisterHandlers(r)

	go func() {
		logger.Info("HTTP server listening at :8001")
		if err := r.Run(":8001"); err != nil {
			log.Fatalf("failed to run server: %v", err)
		}
	}()

	// 9. 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")
}

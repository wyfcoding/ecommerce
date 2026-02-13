package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/wyfcoding/ecommerce/go-api/livestream/v1"
	"github.com/wyfcoding/ecommerce/internal/livestream/application"
	"github.com/wyfcoding/ecommerce/internal/livestream/domain"
	"github.com/wyfcoding/ecommerce/internal/livestream/infrastructure"
	"github.com/wyfcoding/ecommerce/internal/livestream/interfaces"
	"github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := &config.Config{}
	if err := config.Load("configs/livestream/config.toml", cfg); err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := logging.NewLogger(cfg.Server.Name, "main", cfg.Log.Level)

	m := metrics.NewMetrics(cfg.Server.Name)

	db, err := database.NewDB(cfg.Data.Database, cfg.CircuitBreaker, logger, m)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	if err := db.DB.AutoMigrate(&domain.Room{}, &domain.Product{}); err != nil {
		logger.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	repo := infrastructure.NewGormLivestreamRepository(db.DB, rdb)
	cmdSvc := application.NewLivestreamCommandService(repo, nil, logger)
	querySvc := application.NewLivestreamQueryService(repo, logger)
	app := application.NewLivestreamApplicationService(cmdSvc, querySvc)
	handler := interfaces.NewLivestreamHandler(app)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPC.Port))
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterLivestreamServiceServer(s, handler)
	reflection.Register(s)

	fmt.Printf("%s listening at %v\n", cfg.Server.Name, lis.Addr())

	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Error("failed to serve", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	s.GracefulStop()
	logger.Info("server stopped")
}

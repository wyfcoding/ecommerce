package main

import (
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/wyfcoding/ecommerce/go-api/crossborder/v1"
	"github.com/wyfcoding/ecommerce/internal/crossborder/application"
	"github.com/wyfcoding/ecommerce/internal/crossborder/domain"
	persistence_mysql "github.com/wyfcoding/ecommerce/internal/crossborder/infrastructure/persistence/mysql"
	"github.com/wyfcoding/ecommerce/internal/crossborder/interfaces"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(127.0.0.1:3306)/ecommerce_crossborder?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&domain.CustomsDeclaration{}, &domain.DeclarationItem{}, &domain.HSCode{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	repo := persistence_mysql.NewCrossBorderRepository(db)
	app := application.NewCrossBorderService(
		repo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		logger,
	)
	handler := interfaces.NewCrossBorderHandler(app, repo)

	lis, err := net.Listen("tcp", ":9095")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCrossBorderServiceServer(s, handler)

	go func() {
		logger.Info("server started", "addr", ":9095")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")
	s.GracefulStop()
}

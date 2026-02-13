package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/wyfcoding/ecommerce/go-api/merchantsettlement/v1"
	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/application"
	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/domain"
	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/infrastructure"
	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/interfaces"
	"github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := &config.Config{}
	if err := config.Load("configs/merchantsettlement/config.toml", cfg); err != nil {
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

	if err := db.DB.AutoMigrate(&domain.Settlement{}); err != nil {
		logger.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	settlementRepo := infrastructure.NewGormSettlementRepository(db.DB)
	bankAccountRepo := infrastructure.NewGormMerchantBankAccountRepository(db.DB)
	configRepo := infrastructure.NewGormMerchantSettlementConfigRepository(db.DB)

	app := application.NewMerchantSettlementService(
		settlementRepo,
		bankAccountRepo,
		configRepo,
		nil,
		nil,
	)
	handler := interfaces.NewMerchantSettlementHandler(app)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPC.Port))
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterMerchantSettlementServiceServer(s, handler)
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

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	v1 "github.com/wyfcoding/ecommerce/go-api/tax/v1"
	"github.com/wyfcoding/ecommerce/internal/tax/application"
	"github.com/wyfcoding/ecommerce/internal/tax/domain"
	persistence_mysql "github.com/wyfcoding/ecommerce/internal/tax/infrastructure/persistence/mysql"
	grpc_server "github.com/wyfcoding/ecommerce/internal/tax/interfaces/grpc"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// noopCrossBorderTaxRepository 提供跨境税务仓储的兜底实现，避免启动阶段因未接入而阻塞构建。
type noopCrossBorderTaxRepository struct{}

func (noopCrossBorderTaxRepository) FindConfig(_ context.Context, _, _ string, _ domain.TradeType) (*domain.CrossBorderTaxConfig, error) {
	return nil, fmt.Errorf("cross-border tax config repository is not configured")
}

func (noopCrossBorderTaxRepository) SaveConfig(_ context.Context, _ *domain.CrossBorderTaxConfig) error {
	return fmt.Errorf("cross-border tax config repository is not configured")
}

func (noopCrossBorderTaxRepository) FindAgreement(_ context.Context, _ string) (*domain.PreferentialAgreement, error) {
	return nil, fmt.Errorf("cross-border tax config repository is not configured")
}

func (noopCrossBorderTaxRepository) ListAgreements(_ context.Context, _, _ string) ([]*domain.PreferentialAgreement, error) {
	return nil, fmt.Errorf("cross-border tax config repository is not configured")
}

func main() {
	// 1. Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Config (Simplified for skeleton)
	dsn := os.Getenv("To be replaced by config loader")
	if dsn == "" {
		dsn = "root:password@tcp(127.0.0.1:3306)/ecommerce_tax?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// 3. Database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto Migrate
	err = db.AutoMigrate(
		&persistence_mysql.TaxRuleModel{},
		&persistence_mysql.TaxInvoiceModel{},
		&persistence_mysql.TaxExemptionModel{},
	)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// 4. Layers
	repo := persistence_mysql.NewTaxRepository(db)
	crossBorderRepo := noopCrossBorderTaxRepository{}
	app := application.NewTaxService(repo, crossBorderRepo, nil, logger)
	svc := grpc_server.NewServer(app)

	// 5. Server
	lis, err := net.Listen("tcp", ":9090") // Port should be from config
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	v1.RegisterTaxServiceServer(s, svc)

	go func() {
		logger.Info("server started", "addr", ":9090")
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

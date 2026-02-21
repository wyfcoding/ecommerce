package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	pb "github.com/wyfcoding/ecommerce/go-api/risk/v1"
	riskapp "github.com/wyfcoding/ecommerce/internal/risk/application"
	riskmysql "github.com/wyfcoding/ecommerce/internal/risk/infrastructure/persistence/mysql"
	riskgrpc "github.com/wyfcoding/ecommerce/internal/risk/interfaces/grpc"
	riskhttp "github.com/wyfcoding/ecommerce/internal/risk/interfaces/http"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/server"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dsn := envOrDefault("RISK_DSN", "root:password@tcp(127.0.0.1:3306)/ecommerce_risk?charset=utf8mb4&parseTime=True&loc=Local")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	if err := db.AutoMigrate(
		&riskmysql.RiskAnalysisResultModel{},
		&riskmysql.BlacklistModel{},
		&riskmysql.DeviceFingerprintModel{},
		&riskmysql.UserBehaviorModel{},
		&riskmysql.RiskRuleModel{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{envOrDefault("RISK_REDIS_ADDR", "127.0.0.1:6379")}})
	defer func() { _ = redisClient.Close() }()

	repo := riskmysql.NewRiskRepository(db, redisClient)
	cmdSvc := riskapp.NewRiskSecurityCommandService(repo, nil, logger)
	querySvc := riskapp.NewRiskSecurityQueryService(repo, nil, nil, nil, nil, logger)

	httpHandler := riskhttp.NewHandler(cmdSvc, querySvc, logger)
	engine := server.NewDefaultGinEngine(gin.Recovery())
	engine.GET("/api/v1/risk/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})
	httpHandler.RegisterRoutes(engine.Group("/api/v1"))

	httpAddr := envOrDefault("RISK_HTTP_ADDR", ":9206")
	httpSrv := server.NewGinServer(engine, httpAddr, logger)
	go func() {
		if err := httpSrv.Start(context.Background()); err != nil {
			slog.Error("http server exit", "error", err)
		}
	}()

	grpcAddr := envOrDefault("RISK_GRPC_ADDR", ":9306")
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcAddr, err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterRiskServiceServer(grpcSrv, riskgrpc.NewServer(cmdSvc, querySvc))
	go func() {
		logger.Info("risk gRPC started", "addr", grpcAddr)
		if serveErr := grpcSrv.Serve(lis); serveErr != nil {
			logger.Error("risk grpc exit", "error", serveErr)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = httpSrv.Stop(context.Background())
	grpcSrv.GracefulStop()
	slog.Info("service risk gracefully stopped")
}

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

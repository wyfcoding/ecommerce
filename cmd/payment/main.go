package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	v1 "github.com/wyfcoding/ecommerce/go-api/payment/v1"
	settlementv1 "github.com/wyfcoding/ecommerce/go-api/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/application"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/ecommerce/internal/payment/infrastructure/gateway"
	"github.com/wyfcoding/ecommerce/internal/payment/infrastructure/persistence"
	"github.com/wyfcoding/ecommerce/internal/payment/infrastructure/risk"
	grpcServer "github.com/wyfcoding/ecommerce/internal/payment/interfaces/grpc"
	"github.com/wyfcoding/ecommerce/pkg/app"
	configpkg "github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/databases/sharding"
	"github.com/wyfcoding/ecommerce/pkg/idgen"
	"github.com/wyfcoding/ecommerce/pkg/logging"
	"github.com/wyfcoding/ecommerce/pkg/metrics"

	paymenthttp "github.com/wyfcoding/ecommerce/internal/payment/interfaces/http"
	"google.golang.org/grpc"
)

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

func main() {
	app.NewBuilder("payment").
		WithConfig(&Config{}).
		WithService(func(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
			c := cfg.(*Config)

			// Database Sharding Manager
			logger := logging.NewLogger("payment-service", "app")
			shardingManager, err := sharding.NewManager(c.Data.Shards, logger)
			if err != nil {
				// Fallback to single DB
				if len(c.Data.Shards) == 0 {
					logger.Info("No shards configured, using single database connection")
					shardingManager, err = sharding.NewManager([]configpkg.DatabaseConfig{c.Data.Database}, logger)
					if err != nil {
						return nil, nil, err
					}
				} else {
					return nil, nil, err
				}
			}

			cleanupDB := func() {
				shardingManager.Close()
			}

			// Repositories
			paymentRepo := persistence.NewPaymentRepository(shardingManager)
			refundRepo := persistence.NewRefundRepository(shardingManager)
			channelRepo := persistence.NewChannelRepository(shardingManager)

			// ID Generator
			idGenerator, err := idgen.NewSnowflakeGenerator(c.Snowflake)
			if err != nil {
				cleanupDB()
				return nil, nil, err
			}

			// Domain Services
			riskService := risk.NewRiskService()

			// Gateways
			gateways := map[domain.GatewayType]domain.PaymentGateway{
				domain.GatewayTypeAlipay: gateway.NewAlipayGateway(),
				domain.GatewayTypeWechat: gateway.NewWechatGateway(),
				domain.GatewayTypeStripe: gateway.NewStripeGateway(),
				domain.GatewayTypeMock:   gateway.NewAlipayGateway(), // Default to Alipay for mock
			}

			// Settlement Client (Direct Call)
			// TODO: Use Service Discovery / Config
			conn, err := grpc.Dial("localhost:9006", grpc.WithInsecure())
			if err != nil {
				cleanupDB()
				return nil, nil, err
			}
			settlementCli := settlementv1.NewSettlementServiceClient(conn)

			// Applicatopn Service
			appService := application.NewPaymentApplicationService(
				paymentRepo,
				refundRepo,
				channelRepo,
				riskService,
				idGenerator,
				gateways,
				settlementCli,
				logger.Logger,
			)

			// gRPC Server
			srv := grpcServer.NewServer(appService)

			return srv, func() {
				cleanupDB()
			}, nil
		}).
		WithGRPC(func(s *grpc.Server, svc interface{}) {
			v1.RegisterPaymentServer(s, svc.(v1.PaymentServer))
		}).
		WithGin(registerGin).
		Build().
		Run()
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*grpcServer.Server).App
	handler := paymenthttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for payment service (DDD)")
}

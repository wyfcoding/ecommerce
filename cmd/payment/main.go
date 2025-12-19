package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	v1 "github.com/wyfcoding/ecommerce/go-api/payment/v1"
	settlementv1 "github.com/wyfcoding/ecommerce/go-api/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/application"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/ecommerce/internal/payment/infrastructure/gateway"
	"github.com/wyfcoding/ecommerce/internal/payment/infrastructure/persistence"
	"github.com/wyfcoding/ecommerce/internal/payment/infrastructure/risk"
	grpcServer "github.com/wyfcoding/ecommerce/internal/payment/interfaces/grpc"
	paymenthttp "github.com/wyfcoding/ecommerce/internal/payment/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases/sharding"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

const BootstrapName = "payment"

type AppContext struct {
	Config     *configpkg.Config
	AppService *application.PaymentService
	Clients    *ServiceClients
}

type ServiceClients struct {
	Order      *grpc.ClientConn
	User       *grpc.ClientConn
	Settlement *grpc.ClientConn
}

func main() {
	app.NewBuilder(BootstrapName).
		WithConfig(&configpkg.Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(middleware.CORS()).
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, svc interface{}) {
	ctx := svc.(*AppContext)
	v1.RegisterPaymentServer(s, grpcServer.NewServer(ctx.AppService))
}

func registerGin(e *gin.Engine, svc interface{}) {
	ctx := svc.(*AppContext)
	handler := paymenthttp.NewHandler(ctx.AppService, slog.Default())
	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...")

	// 1. Database Sharding Manager
	shardingManager, err := sharding.NewManager(c.Data.Shards, logging.Default())
	if err != nil {
		if len(c.Data.Shards) == 0 {
			slog.Info("No shards configured, using single database connection")
			shardingManager, err = sharding.NewManager([]configpkg.DatabaseConfig{c.Data.Database}, logging.Default())
			if err != nil {
				return nil, nil, fmt.Errorf("failed to initialize single db manager: %w", err)
			}
		} else {
			return nil, nil, fmt.Errorf("failed to initialize sharding manager: %w", err)
		}
	}

	// 2. ID Generator
	idGenerator, err := idgen.NewSnowflakeGenerator(c.Snowflake)
	if err != nil {
		shardingManager.Close()
		return nil, nil, fmt.Errorf("failed to initialize id generator: %w", err)
	}

	// 3. Downstream Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		shardingManager.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. Infrastructure & Application
	paymentRepo := persistence.NewPaymentRepository(shardingManager)
	refundRepo := persistence.NewRefundRepository(shardingManager)
	channelRepo := persistence.NewChannelRepository(shardingManager)
	riskService := risk.NewRiskService()

	gateways := map[domain.GatewayType]domain.PaymentGateway{
		domain.GatewayTypeAlipay: gateway.NewAlipayGateway(),
		domain.GatewayTypeWechat: gateway.NewWechatGateway(),
		domain.GatewayTypeStripe: gateway.NewStripeGateway(),
		domain.GatewayTypeMock:   gateway.NewAlipayGateway(),
	}

	// Create Settlement Client Wrapper
	// Note: If clients.Settlement is nil, this will panic if used.
	// Ensure Settlement service is configured or handle nil gracefully in ServiceClients.
	var settlementCli settlementv1.SettlementServiceClient
	if clients.Settlement != nil {
		settlementCli = settlementv1.NewSettlementServiceClient(clients.Settlement)
	}

	// 5. Initialize Granular Application Services
	processor := application.NewPaymentProcessor(
		paymentRepo,
		channelRepo,
		riskService,
		idGenerator,
		gateways,
		logging.Default().Logger,
	)
	callbackHandler := application.NewCallbackHandler(
		paymentRepo,
		gateways,
		logging.Default().Logger,
	)
	refundService := application.NewRefundService(
		paymentRepo,
		refundRepo,
		idGenerator,
		gateways,
		logging.Default().Logger,
	)
	paymentQuery := application.NewPaymentQuery(paymentRepo)

	// 6. Create PaymentService Facade
	appService := application.NewPaymentService(
		processor,
		callbackHandler,
		refundService,
		paymentQuery,
		settlementCli,
		logging.Default().Logger,
	)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		shardingManager.Close()
	}

	return &AppContext{
		Config:     c,
		AppService: appService,
		Clients:    clients,
	}, cleanup, nil
}

package gatewayhandler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	adminV1 "ecommerce/api/admin/v1"
	analyticsV1 "ecommerce/api/analytics/v1"
	assetV1 "ecommerce/api/asset/v1"
	authV1 "ecommerce/api/auth/v1"
	cartV1 "ecommerce/api/cart/v1"
	configV1 "ecommerce/api/config/v1"
	customerServiceV1 "ecommerce/api/customer_service/v1"
	dataIngestionV1 "ecommerce/api/data_ingestion/v1"
	dataProcessingV1 "ecommerce/api/data_processing/v1"
	logisticsV1 "ecommerce/api/logistics/v1"
	marketingV1 "ecommerce/api/marketing/v1"
	notificationV1 "ecommerce/api/notification/v1"
	orderV1 "ecommerce/api/order/v1"
	paymentV1 "ecommerce/api/payment/v1"
	pricingV1 "ecommerce/api/pricing/v1"
	productV1 "ecommerce/api/product/v1"
	recommendationV1 "ecommerce/api/recommendation/v1"
	riskSecurityV1 "ecommerce/api/risk_security/v1"
	settlementV1 "ecommerce/api/settlement/v1"
	userV1 "ecommerce/api/user/v1"
	"ecommerce/pkg/logging"
)

// RegisterServiceHandlers 动态注册所有 gRPC 服务处理器
func RegisterServiceHandlers(ctx context.Context, gwmux *runtime.ServeMux, servicesConfig map[string]struct{ Addr string }, opts []grpc.DialOption) {
	// 注册用户服务
	if userService, ok := servicesConfig["user"]; ok {
		if err := userV1.RegisterUserHandlerFromEndpoint(ctx, gwmux, userService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register user service: %v", err)
		}
		zap.S().Infof("Registered user service at %s", userService.Addr)
	}

	// 注册商品服务
	if productService, ok := servicesConfig["product"]; ok {
		if err := productV1.RegisterProductHandlerFromEndpoint(ctx, gwmux, productService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register product service: %v", err)
		}
		zap.S().Infof("Registered product service at %s", productService.Addr)
	}

	// 注册购物车服务
	if cartService, ok := servicesConfig["cart"]; ok {
		if err := cartV1.RegisterCartHandlerFromEndpoint(ctx, gwmux, cartService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register cart service: %v", err)
		}
		zap.S().Infof("Registered cart service at %s", cartService.Addr)
	}

	// 注册订单服务
	if orderService, ok := servicesConfig["order"]; ok {
		if err := orderV1.RegisterOrderHandlerFromEndpoint(ctx, gwmux, orderService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register order service: %v", err)
		}
		zap.S().Infof("Registered order service at %s", orderService.Addr)
	}

	// 注册管理服务
	if adminService, ok := servicesConfig["admin"]; ok {
		if err := adminV1.RegisterAdminHandlerFromEndpoint(ctx, gwmux, adminService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register admin service: %v", err)
		}
		zap.S().Infof("Registered admin service at %s", adminService.Addr)
	}

	// 注册营销服务
	if marketingService, ok := servicesConfig["marketing"]; ok {
		if err := marketingV1.RegisterMarketingHandlerFromEndpoint(ctx, gwmux, marketingService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register marketing service: %v", err)
		}
		zap.S().Infof("Registered marketing service at %s", marketingService.Addr)
	}

	// 注册资产服务
	if assetService, ok := servicesConfig["asset"]; ok {
		if err := assetV1.RegisterAssetServiceHandlerFromEndpoint(ctx, gwmux, assetService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register asset service: %v", err)
		}
		zap.S().Infof("Registered asset service at %s", assetService.Addr)
	}

	// 注册分析服务
	if analyticsService, ok := servicesConfig["analytics"]; ok {
		if err := analyticsV1.RegisterAnalyticsServiceHandlerFromEndpoint(ctx, gwmux, analyticsService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register analytics service: %v", err)
		}
		zap.S().Infof("Registered analytics service at %s", analyticsService.Addr)
	}

	// 注册数据摄取服务
	if dataIngestionService, ok := servicesConfig["data_ingestion"]; ok {
		if err := dataIngestionV1.RegisterDataIngestionServiceHandlerFromEndpoint(ctx, gwmux, dataIngestionService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register data ingestion service: %v", err)
		}
		zap.S().Infof("Registered data ingestion service at %s", dataIngestionService.Addr)
	}

	// 注册数据处理服务
	if dataProcessingService, ok := servicesConfig["data_processing"]; ok {
		if err := dataProcessingV1.RegisterDataProcessingServiceHandlerFromEndpoint(ctx, gwmux, dataProcessingService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register data processing service: %v", err)
		}
		zap.S().Infof("Registered data processing service at %s", dataProcessingService.Addr)
	}

	// 注册认证服务
	if authService, ok := servicesConfig["auth"]; ok {
		if err := authV1.RegisterAuthServiceHandlerFromEndpoint(ctx, gwmux, authService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register auth service: %v", err)
		}
		zap.S().Infof("Registered auth service at %s", authService.Addr)
	}

	// 注册定价服务
	if pricingService, ok := servicesConfig["pricing"]; ok {
		if err := pricingV1.RegisterPricingServiceHandlerFromEndpoint(ctx, gwmux, pricingService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register pricing service: %v", err)
		}
		zap.S().Infof("Registered pricing service at %s", pricingService.Addr)
	}

	// 注册物流服务
	if logisticsService, ok := servicesConfig["logistics"]; ok {
		if err := logisticsV1.RegisterLogisticsServiceHandlerFromEndpoint(ctx, gwmux, logisticsService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register logistics service: %v", err)
		}
		zap.S().Infof("Registered logistics service at %s", logisticsService.Addr)
	}

	// 注册库存服务
	if inventoryService, ok := servicesConfig["inventory"]; ok {
		if err := inventoryV1.RegisterInventoryServiceHandlerFromEndpoint(ctx, gwmux, inventoryService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register inventory service: %v", err)
		}
		zap.S().Infof("Registered inventory service at %s", inventoryService.Addr)
	}

	// 注册支付服务
	if paymentService, ok := servicesConfig["payment"]; ok {
		if err := paymentV1.RegisterPaymentServiceHandlerFromEndpoint(ctx, gwmux, paymentService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register payment service: %v", err)
		}
		zap.S().Infof("Registered payment service at %s", paymentService.Addr)
	}

	// 注册客户服务
	if customerService, ok := servicesConfig["customer_service"]; ok {
		if err := customerServiceV1.RegisterCustomerServiceHandlerFromEndpoint(ctx, gwmux, customerService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register customer service: %v", err)
		}
		zap.S().Infof("Registered customer service at %s", customerService.Addr)
	}

	// 注册风控与安全服务
	if riskSecurityService, ok := servicesConfig["risk_security"]; ok {
		if err := riskSecurityV1.RegisterRiskSecurityServiceHandlerFromEndpoint(ctx, gwmux, riskSecurityService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register risk security service: %v", err)
		}
		zap.S().Infof("Registered risk security service at %s", riskSecurityService.Addr)
	}

	// 注册结算服务
	if settlementService, ok := servicesConfig["settlement"]; ok {
		if err := settlementV1.RegisterSettlementServiceHandlerFromEndpoint(ctx, gwmux, settlementService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register settlement service: %v", err)
		}
		zap.S().Infof("Registered settlement service at %s", settlementService.Addr)
	}

	// 注册通知服务
	if notificationService, ok := servicesConfig["notification"]; ok {
		if err := notificationV1.RegisterNotificationServiceHandlerFromEndpoint(ctx, gwmux, notificationService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register notification service: %v", err)
		}
		zap.S().Infof("Registered notification service at %s", notificationService.Addr)
	}

	// 注册配置服务
	if configService, ok := servicesConfig["config"]; ok {
		if err := configV1.RegisterConfigServiceHandlerFromEndpoint(ctx, gwmux, configService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register config service: %v", err)
		}
		zap.S().Infof("Registered config service at %s", configService.Addr)
	}

	// 注册推荐服务 (already had productV1, but this is the correct one)
	if recommendationService, ok := servicesConfig["recommendation"]; ok {
		if err := recommendationV1.RegisterRecommendationServiceHandlerFromEndpoint(ctx, gwmux, recommendationService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register recommendation service: %v", err)
		}
		zap.S().Infof("Registered recommendation service at %s", recommendationService.Addr)
	}
}
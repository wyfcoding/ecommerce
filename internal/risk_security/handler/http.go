package risksecurityhandler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "ecommerce/api/risk_security/v1"
	"ecommerce/internal/risk_security/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer starts the HTTP Gateway, which proxies HTTP requests to gRPC services.
func StartHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int, riskSecurityService *service.RiskSecurityService) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterRiskSecurityServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for RiskSecurityService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery())

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/risk-security")
	{
		api.POST("/anti-fraud-check", performAntiFraudCheckHandler(riskSecurityService))
	}

	r.Any("/*any", gin.WrapH(mux))

	httpEndpoint := fmt.Sprintf("%s:%d", httpAddr, httpPort)
	server := &http.Server{
		Addr:    httpEndpoint,
		Handler: r,
	}

	zap.S().Infof("HTTP server listening at %s", httpEndpoint)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to serve HTTP: %w", err)
		}
		close(errChan)
	}()
	return server, errChan
}

func performAntiFraudCheckHandler(s *service.RiskSecurityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID         string            `json:"user_id"`
			IPAddress      string            `json:"ip_address"`
			DeviceInfo     string            `json:"device_info"`
			OrderID        string            `json:"order_id"`
			Amount         uint64            `json:"amount"`
			AdditionalData map[string]string `json:"additional_data"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		result, err := s.PerformAntiFraudCheck(
			c.Request.Context(),
			req.UserID,
			req.IPAddress,
			req.DeviceInfo,
			req.OrderID,
			req.Amount,
			req.AdditionalData,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

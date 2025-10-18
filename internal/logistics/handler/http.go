package logisticshandler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "ecommerce/api/logistics/v1"
	"ecommerce/internal/logistics/model"
	"ecommerce/internal/logistics/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer 启动 HTTP Gateway
func StartHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int, logisticsService *service.LogisticsService) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterLogisticsServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for LogisticsService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/logistics")
	{
		api.POST("/shipping-cost", calculateShippingCostHandler(logisticsService))
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

func calculateShippingCostHandler(s *service.LogisticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OriginAddress      *model.AddressInfo `json:"origin_address"`
			DestinationAddress *model.AddressInfo `json:"destination_address"`
			Items              []*model.ItemInfo  `json:"items"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		cost, err := s.CalculateShippingCost(c.Request.Context(), req.OriginAddress, req.DestinationAddress, req.Items)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"shipping_cost": cost})
	}
}

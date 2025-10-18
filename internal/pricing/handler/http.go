package pricinghandler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "ecommerce/api/pricing/v1"
	"ecommerce/internal/pricing/model"
	"ecommerce/internal/pricing/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer 启动 HTTP Gateway
func StartHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int, pricingService *service.PricingService) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterPricingServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for PricingService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/pricing")
	{
		api.POST("/calculate-final-price", calculateFinalPriceHandler(pricingService))
		api.POST("/calculate-dynamic-price", calculateDynamicPriceHandler(pricingService))
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

func calculateFinalPriceHandler(s *service.PricingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID     uint64              `json:"user_id"`
			Items      []*model.SkuPriceInfo `json:"items"`
			CouponCode string              `json:"coupon_code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		totalOriginalPrice, totalDiscountAmount, finalPrice, err := s.CalculateFinalPrice(c.Request.Context(), req.UserID, req.Items, req.CouponCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total_original_price": totalOriginalPrice,
			"total_discount_amount": totalDiscountAmount,
			"final_price":          finalPrice,
		})
	}
}

func calculateDynamicPriceHandler(s *service.PricingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ProductID     uint64            `json:"product_id"`
			UserID        uint64            `json:"user_id"`
			ContextFeatures map[string]string `json:"context_features"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		dynamicPrice, explanation, err := s.CalculateDynamicPrice(c.Request.Context(), req.ProductID, req.UserID, req.ContextFeatures)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"dynamic_price": dynamicPrice, "explanation": explanation})
	}
}

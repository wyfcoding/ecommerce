package inventoryhandler

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

	v1 "ecommerce/api/inventory/v1"
	"ecommerce/internal/inventory/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer 启动 HTTP Gateway
func StartHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int, inventoryService *service.InventoryService) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterInventoryServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for InventoryService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/inventory")
	{
		api.GET("/stock/:sku_id", getStockHandler(inventoryService))
		api.POST("/stock/deduct", deductStockHandler(inventoryService))
		api.POST("/stock/release", releaseStockHandler(inventoryService))
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

func getStockHandler(s *service.InventoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid SKU ID"})
			return
		}
		stock, err := s.GetStock(c.Request.Context(), skuID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, stock)
	}
}

func deductStockHandler(s *service.InventoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SkuID    uint64 `json:"sku_id"`
			Quantity uint32 `json:"quantity"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if err := s.DeductStock(c.Request.Context(), req.SkuID, req.Quantity); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "stock deducted successfully"})
	}
}

func releaseStockHandler(s *service.InventoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SkuID    uint64 `json:"sku_id"`
			Quantity uint32 `json:"quantity"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if err := s.ReleaseStock(c.Request.Context(), req.SkuID, req.Quantity); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "stock released successfully"})
	}
}

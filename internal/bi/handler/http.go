package bihandler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "ecommerce/api/bi/v1"
	"ecommerce/internal/bi/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer 启动 HTTP Gateway
func StartHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int, biService *service.BiService) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterBIServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for BiService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/bi")
	{
		api.GET("/sales/overview", getSalesOverviewHandler(biService))
		api.GET("/products/top-selling", getTopSellingProductsHandler(biService))
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

func getSalesOverviewHandler(s *service.BiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var startDate, endDate *time.Time
		if sdStr := c.Query("start_date"); sdStr != "" {
			sd, err := time.Parse("2006-01-02", sdStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
				return
			}
			startDate = &sd
		}
		if edStr := c.Query("end_date"); edStr != "" {
			ed, err := time.Parse("2006-01-02", edStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
				return
			}
			endDate = &ed
		}

		overview, err := s.GetSalesOverview(c.Request.Context(), startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, overview)
	}
}

func getTopSellingProductsHandler(s *service.BiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.ParseUint(limitStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}

		var startDate, endDate *time.Time
		// Similar date parsing as in getSalesOverviewHandler
		if sdStr := c.Query("start_date"); sdStr != "" {
			sd, err := time.Parse("2006-01-02", sdStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
				return
			}
			startDate = &sd
		}
		if edStr := c.Query("end_date"); edStr != "" {
			ed, err := time.Parse("2006-01-02", edStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
				return
			}
			endDate = &ed
		}

		products, err := s.GetTopSellingProducts(c.Request.Context(), uint32(limit), startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, products)
	}
}

package salesforecastinghandler

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

	v1 "ecommerce/api/sales_forecasting/v1"
	"ecommerce/internal/sales_forecasting/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer starts the HTTP Gateway, which proxies HTTP requests to gRPC services.
func StartHTTPServer(svc *service.SalesForecastingService, addr string, port int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", addr, port)

	err := v1.RegisterSalesForecastingServiceHandlerFromEndpoint(context.Background(), mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for SalesForecastingService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/sales-forecasting")
	{
		api.GET("/products/:product_id", getProductSalesForecastHandler(svc))
		api.POST("/models/train", trainSalesForecastModelHandler(svc))
	}

	r.Any("/*any", gin.WrapH(mux))

	httpEndpoint := fmt.Sprintf("%s:%d", addr, port)
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

func getProductSalesForecastHandler(s *service.SalesForecastingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID, err := strconv.ParseUint(c.Param("product_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
			return
		}
		forecastDays, err := strconv.ParseUint(c.DefaultQuery("forecast_days", "7"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid forecast days"})
			return
		}

		forecasts, err := s.GetProductSalesForecast(c.Request.Context(), productID, uint32(forecastDays))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, forecasts)
	}
}

func trainSalesForecastModelHandler(s *service.SalesForecastingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ModelName  string            `json:"model_name"`
			DataSource string            `json:"data_source"`
			Parameters map[string]string `json:"parameters"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		jobID, status, err := s.TrainSalesForecastModel(c.Request.Context(), req.ModelName, req.DataSource, req.Parameters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"job_id": jobID, "status": status})
	}
}

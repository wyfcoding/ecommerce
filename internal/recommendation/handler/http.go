package recommendationhandler

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

	v1 "ecommerce/api/recommendation/v1"
	"ecommerce/internal/recommendation/model"
	"ecommerce/internal/recommendation/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer starts the HTTP Gateway.
func StartHTTPServer(svc *service.RecommendationService, addr string, port int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", addr, port)

	err := v1.RegisterRecommendationServiceHandlerFromEndpoint(context.Background(), mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for RecommendationService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/recommendations")
	{
		api.GET("/users/:user_id", getRecommendedProductsHandler(svc))
		api.POST("/relationships", indexProductRelationshipHandler(svc))
		api.GET("/graph/:product_id", getGraphRecommendedProductsHandler(svc))
		api.POST("/advanced", getAdvancedRecommendedProductsHandler(svc))
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

func getRecommendedProductsHandler(s *service.RecommendationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		count, err := strconv.ParseInt(c.DefaultQuery("count", "10"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid count"})
			return
		}

		products, err := s.GetRecommendedProducts(c.Request.Context(), userID, int32(count))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, products)
	}
}

func indexProductRelationshipHandler(s *service.RecommendationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.ProductRelationship
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if err := s.IndexProductRelationship(c.Request.Context(), &req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "product relationship indexed"})
	}
}

func getGraphRecommendedProductsHandler(s *service.RecommendationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("product_id")
		count, err := strconv.ParseInt(c.DefaultQuery("count", "10"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid count"})
			return
		}

		products, err := s.GetGraphRecommendedProducts(c.Request.Context(), productID, int32(count))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, products)
	}
}

func getAdvancedRecommendedProductsHandler(s *service.RecommendationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID          string            `json:"user_id"`
			Count           int32             `json:"count"`
			ContextFeatures map[string]string `json:"context_features"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		products, explanation, err := s.GetAdvancedRecommendedProducts(c.Request.Context(), req.UserID, req.Count, req.ContextFeatures);
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"products": products, "explanation": explanation})
	}
}

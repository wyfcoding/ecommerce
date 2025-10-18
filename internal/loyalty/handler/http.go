package loyaltyhandler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ecommerce/internal/loyalty/model"
	"ecommerce/internal/loyalty/service"
	"ecommerce/pkg/logging"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// StartHTTPServer starts the HTTP server.
func StartHTTPServer(loyaltyService *service.LoyaltyService, addr string, port int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	r := gin.New()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Register routes
	registerRoutes(r, loyaltyService)

	httpEndpoint := fmt.Sprintf("%s:%d", addr, port)
	server := &http.Server{
		Addr:    httpEndpoint,
		Handler: r,
	}

	zap.S().Infof("Gin HTTP server listening at %s", httpEndpoint)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to serve Gin HTTP: %w", err)
		}
		close(errChan)
	}()
	return server, errChan
}

func registerRoutes(r *gin.Engine, svc *service.LoyaltyService) {
	api := r.Group("/api/v1/loyalty")
	{
		api.GET("/profiles/:user_id", getUserLoyaltyProfileHandler(svc))
		api.POST("/points/add", addPointsHandler(svc))
		api.POST("/points/deduct", deductPointsHandler(svc))
		api.GET("/transactions/:user_id", listPointsTransactionsHandler(svc))
		api.PUT("/profiles/:user_id/level", updateUserLevelHandler(svc))
	}
}

func getUserLoyaltyProfileHandler(s *service.LoyaltyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		profile, err := s.GetUserLoyaltyProfile(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, profile)
	}
}

func addPointsHandler(s *service.LoyaltyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID string `json:"user_id"`
			Points int64  `json:"points"`
			Reason string `json:"reason"`
			OrderID string `json:"order_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		profile, transaction, err := s.AddPoints(c.Request.Context(), req.UserID, req.Points, req.Reason, req.OrderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"profile": profile, "transaction": transaction})
	}
}

func deductPointsHandler(s *service.LoyaltyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID string `json:"user_id"`
			Points int64  `json:"points"`
			Reason string `json:"reason"`
			OrderID string `json:"order_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		profile, transaction, err := s.DeductPoints(c.Request.Context(), req.UserID, req.Points, req.Reason, req.OrderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"profile": profile, "transaction": transaction})
	}
}

func listPointsTransactionsHandler(s *service.LoyaltyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 32)
		pageToken, _ := strconv.ParseInt(c.DefaultQuery("page_token", "0"), 10, 32)
		transactions, total, err := s.ListPointsTransactions(c.Request.Context(), userID, int32(pageSize), int32(pageToken))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"transactions": transactions, "total": total})
	}
}

func updateUserLevelHandler(s *service.LoyaltyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		var req struct {
			NewLevel string `json:"new_level"`
			Reason   string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		profile, err := s.UpdateUserLevel(c.Request.Context(), userID, req.NewLevel, req.Reason)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, profile)
	}
}

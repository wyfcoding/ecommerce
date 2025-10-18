package frauddetectionhandler

import (
	"context"
	"fmt"
	"net/http"

	"ecommerce/internal/fraud_detection/service"
	"ecommerce/pkg/logging"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// StartHTTPServer starts the HTTP server.
func StartHTTPServer(fraudDetectionService *service.FraudDetectionService, addr string, port int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	r := gin.New()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Register routes
	registerRoutes(r, fraudDetectionService)

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

func registerRoutes(r *gin.Engine, svc *service.FraudDetectionService) {
	api := r.Group("/api/v1/fraud-detection")
	{
		api.POST("/evaluate", evaluateTransactionHandler(svc))
		api.GET("/evaluations/:transaction_id", getEvaluationStatusHandler(svc))
		api.POST("/report", reportFraudHandler(svc))
	}
}

func evaluateTransactionHandler(s *service.FraudDetectionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TransactionID     string            `json:"transaction_id"`
			UserID            string            `json:"user_id"`
			Amount            float64           `json:"amount"`
			Currency          string            `json:"currency"`
			PaymentMethodType string            `json:"payment_method_type"`
			IpAddress         string            `json:"ip_address"`
			UserAgent         string            `json:"user_agent"`
			AdditionalData    map[string]string `json:"additional_data"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		evaluation, err := s.EvaluateTransaction(
			c.Request.Context(),
			req.TransactionID,
			req.UserID,
			req.Amount,
			req.Currency,
			req.PaymentMethodType,
			req.IpAddress,
			req.UserAgent,
			req.AdditionalData,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, evaluation)
	}
}

func getEvaluationStatusHandler(s *service.FraudDetectionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		transactionID := c.Param("transaction_id")
		evaluation, err := s.GetEvaluationStatus(c.Request.Context(), transactionID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, evaluation)
	}
}

func reportFraudHandler(s *service.FraudDetectionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TransactionID string            `json:"transaction_id"`
			UserID        string            `json:"user_id"`
			ReportReason  string            `json:"report_reason"`
			Evidence      map[string]string `json:"evidence"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		report, err := s.ReportFraud(c.Request.Context(), req.TransactionID, req.UserID, req.ReportReason, req.Evidence)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, report)
	}
}

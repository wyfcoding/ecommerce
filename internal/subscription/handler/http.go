package subscriptionhandler

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

	v1 "ecommerce/api/subscription/v1"
	"ecommerce/internal/subscription/model"
	"ecommerce/internal/subscription/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer starts the HTTP Gateway.
func StartHTTPServer(svc *service.SubscriptionService, addr string, port int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", addr, port)

	err := v1.RegisterSubscriptionServiceHandlerFromEndpoint(context.Background(), mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for SubscriptionService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/subscriptions")
	{
		// Subscription Plans
		api.POST("/plans", createSubscriptionPlanHandler(svc))
		api.GET("/plans/:id", getSubscriptionPlanHandler(svc))
		api.GET("/plans", listSubscriptionPlansHandler(svc))
		api.PUT("/plans/:id", updateSubscriptionPlanHandler(svc))

		// User Subscriptions
		api.POST("/users", createUserSubscriptionHandler(svc))
		api.GET("/users/:id", getUserSubscriptionHandler(svc))
		api.PUT("/users/:id/cancel", cancelUserSubscriptionHandler(svc))
		api.PUT("/users/:id", updateUserSubscriptionHandler(svc))
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

func createSubscriptionPlanHandler(s *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name           string  `json:"name"`
			Description    string  `json:"description"`
			Price          float64 `json:"price"`
			Currency       string  `json:"currency"`
			RecurrenceType string  `json:"recurrence_type"`
			DurationMonths int32   `json:"duration_months"`
			IsActive       bool    `json:"is_active"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		plan, err := s.CreateSubscriptionPlan(c.Request.Context(), req.Name, req.Description, req.Price, req.Currency, req.RecurrenceType, req.DurationMonths, req.IsActive)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, plan)
	}
}

func getSubscriptionPlanHandler(s *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan ID"})
			return
		}
		plan, err := s.GetSubscriptionPlan(c.Request.Context(), uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, plan)
	}
}

func listSubscriptionPlansHandler(s *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		activeOnly, _ := strconv.ParseBool(c.DefaultQuery("active_only", "false"))
		plans, total, err := s.ListSubscriptionPlans(c.Request.Context(), activeOnly)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"plans": plans, "total": total})
	}
}

func updateSubscriptionPlanHandler(s *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan ID"})
			return
		}
		var req model.SubscriptionPlan
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.ID = uint(id)
		plan, err := s.UpdateSubscriptionPlan(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, plan)
	}
}

func createUserSubscriptionHandler(s *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID          string `json:"user_id"`
			PlanID          string `json:"plan_id"`
			PaymentMethodID string `json:"payment_method_id"`
			AutoRenew       bool   `json:"auto_renew"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		sub, err := s.CreateUserSubscription(c.Request.Context(), req.UserID, req.PlanID, req.PaymentMethodID, req.AutoRenew)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, sub)
	}
}

func getUserSubscriptionHandler(s *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
			return
		}
		sub, err := s.GetUserSubscription(c.Request.Context(), uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sub)
	}
}

func cancelUserSubscriptionHandler(s *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
			return
		}
		sub, err := s.CancelUserSubscription(c.Request.Context(), uint(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sub)
	}
}

func updateUserSubscriptionHandler(s *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
			return
		}
		var req struct {
			PlanID          string `json:"plan_id"`
			PaymentMethodID string `json:"payment_method_id"`
			AutoRenew       bool   `json:"auto_renew"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		sub, err := s.UpdateUserSubscription(c.Request.Context(), uint(id), req.PlanID, req.PaymentMethodID, req.AutoRenew)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sub)
	}
}

package flashsalehandler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ecommerce/internal/flashsale/model"
	"ecommerce/internal/flashsale/service"
	"ecommerce/pkg/logging"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// StartHTTPServer starts the HTTP server.
func StartHTTPServer(flashSaleService *service.FlashSaleService, addr string, port int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	r := gin.New()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Register routes
	registerRoutes(r, flashSaleService)

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

func registerRoutes(r *gin.Engine, svc *service.FlashSaleService) {
	api := r.Group("/api/v1/flashsale")
	{
		api.POST("/events", createFlashSaleEventHandler(svc))
		api.GET("/events/:id", getFlashSaleEventHandler(svc))
		api.GET("/events/active", listActiveFlashSaleEventsHandler(svc))
		api.POST("/participate", participateInFlashSaleHandler(svc))
		api.GET("/products/:event_id/:product_id", getFlashSaleProductDetailsHandler(svc))
	}
}

func createFlashSaleEventHandler(svc *service.FlashSaleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string                  `json:"name"`
			Description string                  `json:"description"`
			StartTime   time.Time               `json:"start_time"`
			EndTime     time.Time               `json:"end_time"`
			Products    []*model.FlashSaleProduct `json:"products"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		event, err := svc.CreateFlashSaleEvent(c.Request.Context(), req.Name, req.Description, req.StartTime, req.EndTime, req.Products)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, event)
	}
}

func getFlashSaleEventHandler(svc *service.FlashSaleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
			return
		}
		event, err := svc.GetFlashSaleEvent(c.Request.Context(), uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, event)
	}
}

func listActiveFlashSaleEventsHandler(svc *service.FlashSaleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		events, _, err := svc.ListActiveFlashSaleEvents(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, events)
	}
}

func participateInFlashSaleHandler(svc *service.FlashSaleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			EventID   uint   `json:"event_id"`
			ProductID string `json:"product_id"`
			UserID    string `json:"user_id"`
			Quantity  int32  `json:"quantity"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		orderID, status, err := svc.ParticipateInFlashSale(c.Request.Context(), req.EventID, req.ProductID, req.UserID, req.Quantity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"order_id": orderID, "status": status})
	}
}

func getFlashSaleProductDetailsHandler(svc *service.FlashSaleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
			return
		}
		productID := c.Param("product_id")

		product, err := svc.GetFlashSaleProductDetails(c.Request.Context(), uint(eventID), productID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, product)
	}
}

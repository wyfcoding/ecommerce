package customerservicehandler

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

	v1 "ecommerce/api/customer_service/v1"
	"ecommerce/internal/customer_service/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer 启动 HTTP Gateway
func StartHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int, customerServiceService *service.CustomerServiceService) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterCustomerServiceServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for CustomerServiceService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/customer-service")
	{
		api.POST("/tickets", createTicketHandler(customerServiceService))
		api.GET("/tickets/:ticket_id", getTicketHandler(customerServiceService))
		api.POST("/tickets/:ticket_id/messages", addTicketMessageHandler(customerServiceService))
		api.GET("/users/:user_id/tickets", listTicketsHandler(customerServiceService))
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

func createTicketHandler(s *service.CustomerServiceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID      uint64 `json:"user_id"`
			Subject     string `json:"subject"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		ticket, err := s.CreateTicket(c.Request.Context(), req.UserID, req.Subject, req.Description)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, ticket)
	}
}

func getTicketHandler(s *service.CustomerServiceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketID := c.Param("ticket_id")
		ticket, err := s.GetTicket(c.Request.Context(), ticketID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ticket)
	}
}

func addTicketMessageHandler(s *service.CustomerServiceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketID := c.Param("ticket_id")
		var req struct {
			SenderID   uint64 `json:"sender_id"`
			SenderType string `json:"sender_type"`
			Content    string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		message, err := s.AddTicketMessage(c.Request.Context(), ticketID, req.SenderID, req.SenderType, req.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, message)
	}
}

func listTicketsHandler(s *service.CustomerServiceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		status := c.Query("status")
		tickets, err := s.ListTickets(c.Request.Context(), userID, status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tickets)
	}
}

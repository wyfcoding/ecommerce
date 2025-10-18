package notificationhandler

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

	v1 "ecommerce/api/notification/v1"
	"ecommerce/internal/notification/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer 启动 HTTP Gateway
func StartHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int, notificationService *service.NotificationService) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterNotificationServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for NotificationService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/notifications")
	{
		api.POST("/send", sendNotificationHandler(notificationService))
		api.GET("/users/:user_id", listNotificationsHandler(notificationService))
		api.PUT("/:notification_id/read", markNotificationAsReadHandler(notificationService))
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

func sendNotificationHandler(s *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID  uint64 `json:"user_id"`
			Type    string `json:"type"`
			Title   string `json:"title"`
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		notification, err := s.SendNotification(c.Request.Context(), req.UserID, req.Type, req.Title, req.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, notification)
	}
}

func listNotificationsHandler(s *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}
		includeRead, _ := strconv.ParseBool(c.DefaultQuery("include_read", "false"))
		pageSize, _ := strconv.ParseUint(c.DefaultQuery("page_size", "10"), 10, 32)
		pageNum, _ := strconv.ParseUint(c.DefaultQuery("page_num", "1"), 10, 32)

		notifications, total, err := s.ListNotifications(c.Request.Context(), userID, includeRead, uint32(pageSize), uint32(pageNum))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"notifications": notifications, "total": total})
	}
}

func markNotificationAsReadHandler(s *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		notificationID := c.Param("notification_id")
		if err := s.MarkNotificationAsRead(c.Request.Context(), notificationID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
	}
}

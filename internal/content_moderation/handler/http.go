package contentmoderationhandler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "ecommerce/api/content_moderation/v1"
	"ecommerce/internal/content_moderation/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer 启动 HTTP Gateway
func StartHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int, contentModerationService *service.ContentModerationService) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterContentModerationServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for ContentModerationService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/content-moderation")
	{
		api.POST("/text", moderateTextHandler(contentModerationService))
		api.POST("/image", moderateImageHandler(contentModerationService))
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

func moderateTextHandler(s *service.ContentModerationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ContentID   string `json:"content_id"`
			ContentType string `json:"content_type"`
			UserID      string `json:"user_id"`
			TextContent string `json:"text_content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		result, err := s.ModerateText(
			c.Request.Context(),
			req.ContentID,
			req.ContentType,
			req.UserID,
			req.TextContent,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func moderateImageHandler(s *service.ContentModerationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ContentID   string `json:"content_id"`
			ContentType string `json:"content_type"`
			UserID      string `json:"user_id"`
			ImageURL    string `json:"image_url"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		result, err := s.ModerateImage(
			c.Request.Context(),
			req.ContentID,
			req.ContentType,
			req.UserID,
			req.ImageURL,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

package aimodelhandler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/ai_model/service"
	"ecommerce/pkg/logging"
)

// StartGinServer 启动 Gin HTTP 服务器
func StartGinServer(svc *service.AiModelService, addr string, port int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	r := gin.New()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// --- Register service-specific HTTP routes here ---
	// Example:
	// aiModelHandler := handler.NewAiModelHandler(svc)
	// r.POST("/api/v1/ai_model/predict", aiModelHandler.Predict)
	//
	// --- Application-level Degradation (降级) ---
	// Degradation logic is highly application-specific.
	// For example, if a downstream service is unavailable, you might:
	// - Return cached data.
	// - Return a default response.
	// - Redirect to a static error page.
	// - Use a simplified version of the feature.
	// This typically involves checking the health/status of dependencies
	// or implementing fallback logic within your handlers.
	// --------------------------------------------------

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

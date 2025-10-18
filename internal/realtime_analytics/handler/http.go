package realtimeanalyticshandler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "ecommerce/api/realtime_analytics/v1"
	"ecommerce/internal/realtime_analytics/service"
	"ecommerce/pkg/logging"
)

// StartHTTPServer 启动 HTTP Gateway
func StartHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int, realtimeAnalyticsService *service.RealtimeAnalyticsService) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterRealtimeAnalyticsServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for RealtimeAnalyticsService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // Use project's GinLogger

	// Add service-specific Gin routes here
	api := r.Group("/api/v1/realtime-analytics")
	{
		api.GET("/sales-metrics", getRealtimeSalesMetricsHandler(realtimeAnalyticsService))
		api.GET("/user-activity", getRealtimeUserActivityHandler(realtimeAnalyticsService))
		api.POST("/user-behavior", recordUserBehaviorHandler(realtimeAnalyticsService))
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

func getRealtimeSalesMetricsHandler(s *service.RealtimeAnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics, err := s.GetRealtimeSalesMetrics(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, metrics)
	}
}

func getRealtimeUserActivityHandler(s *service.RealtimeAnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		activity, err := s.GetRealtimeUserActivity(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, activity)
	}
}

func recordUserBehaviorHandler(s *service.RealtimeAnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID       string            `json:"user_id"`
			BehaviorType string            `json:"behavior_type"`
			ItemID       string            `json:"item_id"`
			Properties   map[string]string `json:"properties"`
			EventTime    time.Time         `json:"event_time"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if err := s.RecordUserBehavior(
			c.Request.Context(),
			req.UserID,
			req.BehaviorType,
			req.ItemID,
			req.Properties,
			req.EventTime,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user behavior recorded"})
	}
}

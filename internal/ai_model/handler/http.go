package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	v1 "ecommerce/api/ai_model/v1"
	"ecommerce/internal/ai_model/service"
	"ecommerce/pkg/jwt"
	"ecommerce/pkg/logging"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterRoutes 注册所有的 HTTP 路由。
func RegisterRoutes(r *gin.Engine, aiModelService *service.AIModelService) {
	// 健康检查路由
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1/ai_model")

	// 用户相关路由，需要用户认证
	userProtected := api.Group("/")
	userProtected.Use(userAuthMiddleware())
	{ // 推荐系统
		userProtected.GET("/recommendations/products", getProductRecommendationsUserHandler(aiModelService))
		userProtected.GET("/recommendations/related_products/:product_id", getRelatedProductsUserHandler(aiModelService))
		userProtected.GET("/recommendations/feed", getPersonalizedFeedUserHandler(aiModelService))
		// NLP
		userProtected.POST("/nlp/analyze_review_sentiment", analyzeReviewSentimentUserHandler(aiModelService))
	}

	// 公开路由 (例如，某些图像识别可能不需要用户认证)
	api.POST("/image/recognize", recognizeImageContentHandler(aiModelService))
	api.POST("/image/search_by_image", searchImageByImageHandler(aiModelService))

	// 管理员相关路由，需要管理员认证
	adminProtected := api.Group("/admin")
	adminProtected.Use(adminAuthMiddleware())
	{ // 模型管理
		adminProtected.POST("/models/deploy", deployModelAdminHandler(aiModelService))
		adminProtected.GET("/models/:deployment_id/status", getModelStatusAdminHandler(aiModelService))
		adminProtected.POST("/models/retrain", retrainModelAdminHandler(aiModelService))
		// 欺诈检测 (通常由内部服务调用，但也可以暴露给管理员查看)
		adminProtected.POST("/fraud_detection/score", getFraudScoreAdminHandler(aiModelService))
		// NLP
		adminProtected.POST("/nlp/extract_keywords", extractKeywordsFromTextAdminHandler(aiModelService))
		adminProtected.POST("/nlp/summarize_text", summarizeTextAdminHandler(aiModelService))
	}
}

// StartHTTPServer 启动 HTTP Gateway。
// 它将 gRPC 服务通过 grpc-gateway 暴露为 HTTP/JSON 接口。
func StartHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int, aiModelService *service.AIModelService) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := gwruntime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	// 注册 AIModelService 的 gRPC-Gateway 处理器
	err := v1.RegisterAIModelServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for AIModelService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // 使用项目统一的 GinLogger

	// 将 Gin 路由和 gRPC-Gateway 路由合并
	RegisterRoutes(r, aiModelService) // 注册 Gin 自己的路由
	r.Any("/*any", gin.WrapH(mux)) // 将所有未匹配的请求转发给 gRPC-Gateway

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

// --- HTTP Handlers (User) ---

// getProductRecommendationsUserHandler 处理用户获取商品推荐请求。
func getProductRecommendationsUserHandler(s *service.AIModelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.GetProductRecommendationsRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			zap.S().Warnf("getProductRecommendationsUserHandler: failed to bind query: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}

		// 从上下文获取用户ID并设置到请求中
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}
		req.UserId = userID.(uint64)

		resp, err := s.GetProductRecommendations(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp.Recommendations)
	}
}

// getRelatedProductsUserHandler 处理用户获取相关商品请求。
func getRelatedProductsUserHandler(s *service.AIModelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID, err := strconv.ParseUint(c.Param("product_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
			return
		}

		var req v1.GetRelatedProductsRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			zap.S().Warnf("getRelatedProductsUserHandler: failed to bind query: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}
		req.ProductId = productID

		resp, err := s.GetRelatedProducts(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp.RelatedProducts)
	}
}

// getPersonalizedFeedUserHandler 处理用户获取个性化内容流请求。
func getPersonalizedFeedUserHandler(s *service.AIModelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.GetPersonalizedFeedRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			zap.S().Warnf("getPersonalizedFeedUserHandler: failed to bind query: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}

		// 从上下文获取用户ID并设置到请求中
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}
		req.UserId = userID.(uint64)

		resp, err := s.GetPersonalizedFeed(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp.FeedItems)
	}
}

// analyzeReviewSentimentUserHandler 处理用户分析评论情感请求。
func analyzeReviewSentimentUserHandler(s *service.AIModelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.AnalyzeReviewSentimentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			zap.S().Warnf("analyzeReviewSentimentUserHandler: failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := s.AnalyzeReviewSentiment(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// --- HTTP Handlers (Public) ---

// recognizeImageContentHandler 处理识别图片内容请求。
func recognizeImageContentHandler(s *service.AIModelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.RecognizeImageContentRequest
		// 尝试从 JSON body 绑定，如果失败则从 form data 绑定 (支持文件上传)
		if err := c.ShouldBindJSON(&req); err != nil {
			// 如果 JSON 绑定失败，尝试 form data
			if err := c.ShouldBind(&req); err != nil {
				zap.S().Warnf("recognizeImageContentHandler: failed to bind request: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body or form data"})
				return
			}
		}

		// TODO: 处理文件上传的 image_data

		resp, err := s.RecognizeImageContent(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// searchImageByImageHandler 处理通过图片搜索相似商品请求。
func searchImageByImageHandler(s *service.AIModelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.SearchImageByImageRequest
		// 尝试从 JSON body 绑定，如果失败则从 form data 绑定 (支持文件上传)
		if err := c.ShouldBindJSON(&req); err != nil {
			// 如果 JSON 绑定失败，尝试 form data
			if err := c.ShouldBind(&req); err != nil {
				zap.S().Warnf("searchImageByImageHandler: failed to bind request: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body or form data"})
				return
			}
		}

		// TODO: 处理文件上传的 image_data

		resp, err := s.SearchImageByImage(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// --- HTTP Handlers (Admin) ---

// deployModelAdminHandler 处理管理员部署AI模型请求。
func deployModelAdminHandler(s *service.AIModelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.DeployModelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			zap.S().Warnf("deployModelAdminHandler: failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := s.DeployModel(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

// getModelStatusAdminHandler 处理管理员获取模型状态请求。
func getModelStatusAdminHandler(s *service.AIModelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		deploymentID := c.Param("deployment_id")

		resp, err := s.GetModelStatus(c.Request.Context(), &v1.GetModelStatusRequest{DeploymentId: deploymentID})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// retrainModelAdminHandler 处理管理员触发模型重新训练请求。
func retrainModelAdminHandler(s *service.AIModelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.RetrainModelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			zap.S().Warnf("retrainModelAdminHandler: failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := s.RetrainModel(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// getFraudScoreAdminHandler 处理管理员获取欺诈评分请求。
func getFraudScoreAdminHandler(s *service.AIModelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.GetFraudScoreRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			zap.S().Warnf("getFraudScoreAdminHandler: failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := s.GetFraudScore(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// extractKeywordsFromTextAdminHandler 处理管理员从文本中提取关键词请求。
func extractKeywordsFromTextAdminHandler(s *service.AIModelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.ExtractKeywordsFromTextRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			zap.S().Warnf("extractKeywordsFromTextAdminHandler: failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := s.ExtractKeywordsFromText(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// summarizeTextAdminHandler 处理管理员总结长文本内容请求。
func summarizeTextAdminHandler(s *service.AIModelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.SummarizeTextRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			zap.S().Warnf("summarizeTextAdminHandler: failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := s.SummarizeText(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// --- Middleware ---

// userAuthMiddleware 是一个 Gin 中间件，用于验证用户 JWT Token 的有效性。
func userAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Gin context 中获取 JWT Secret，这里简化为硬编码，实际应从配置或服务中获取
		// TODO: 从配置中获取 JWT Secret
		jwtSecret := "your-user-jwt-secret" // 替换为实际的用户服务 JWT Secret

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		claims, err := jwt.ParseToken(tokenString, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// 将解析出的用户信息存储在 context 中，以便后续的 handler 使用。
		// 同时通过 gRPC metadata 传递给 service 层。
		c.Set("userID", claims.UserID)
		md := metadata.Pairs("x-user-id", strconv.FormatUint(claims.UserID, 10), "x-username", claims.Username)
		newCtx := metadata.NewIncomingContext(c.Request.Context(), md)
		c.Request = c.Request.WithContext(newCtx)
		c.Next()
	}
}

// adminAuthMiddleware 是一个 Gin 中间件，用于验证管理员 JWT Token 的有效性。
func adminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Gin context 中获取 JWT Secret，这里简化为硬编码，实际应从配置或服务中获取
		// TODO: 从配置中获取 JWT Secret
		jwtSecret := "your-admin-jwt-secret" // 替换为实际的管理员服务 JWT Secret

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		claims, err := jwt.ParseToken(tokenString, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// 将解析出的管理员用户信息存储在 context 中，以便后续的 handler 使用。
		// 同时通过 gRPC metadata 传递给 service 层。
		c.Set("adminUserID", claims.UserID)
		md := metadata.Pairs("x-admin-user-id", strconv.FormatUint(claims.UserID, 10), "x-admin-username", claims.Username, "x-is-admin", "true")
		newCtx := metadata.NewIncomingContext(c.Request.Context(), md)
		c.Request = c.Request.WithContext(newCtx)
		c.Next()
	}
}

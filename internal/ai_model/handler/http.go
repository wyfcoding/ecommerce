package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	v1 "ecommerce/api/ai_model/v1"
	"ecommerce/internal/ai_model/service"
	"ecommerce/pkg/jwt"
	"ecommerce/pkg/logging"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

// AIModelHandler 负责处理 AI 模型服务的 HTTP 请求。
type AIModelHandler struct {
	svc    service.AIModelService // 业务逻辑服务接口
	logger *zap.Logger
}

// NewAIModelHandler 创建一个新的 AIModelHandler 实例。
func NewAIModelHandler(svc service.AIModelService, logger *zap.Logger) *AIModelHandler {
	return &AIModelHandler{svc: svc, logger: logger}
}

// RegisterRoutes 注册所有的 HTTP 路由。
func (h *AIModelHandler) RegisterRoutes(r *gin.Engine) {
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
	userProtected.Use(h.userAuthMiddleware())
	{ // 推荐系统
		userProtected.GET("/recommendations/products", h.getProductRecommendationsUserHandler())
		userProtected.GET("/recommendations/related_products/:product_id", h.getRelatedProductsUserHandler())
		userProtected.GET("/recommendations/feed", h.getPersonalizedFeedUserHandler())
		// NLP
		userProtected.POST("/nlp/analyze_review_sentiment", h.analyzeReviewSentimentUserHandler())
	}

	// 公开路由 (例如，某些图像识别可能不需要用户认证)
	api.POST("/image/recognize", h.recognizeImageContentHandler())
	api.POST("/image/search_by_image", h.searchImageByImageHandler())

	// 管理员相关路由，需要管理员认证
	adminProtected := api.Group("/admin")
	adminProtected.Use(h.adminAuthMiddleware())
	{ // 模型管理
		adminProtected.POST("/models/deploy", h.deployModelAdminHandler())
		adminProtected.GET("/models/:deployment_id/status", h.getModelStatusAdminHandler())
		adminProtected.POST("/models/retrain", h.retrainModelAdminHandler())
		// 欺诈检测 (通常由内部服务调用，但也可以暴露给管理员查看)
		adminProtected.POST("/fraud_detection/score", h.getFraudScoreAdminHandler())
		// NLP
		adminProtected.POST("/nlp/extract_keywords", h.extractKeywordsFromTextAdminHandler())
		adminProtected.POST("/nlp/summarize_text", h.summarizeTextAdminHandler())
	}
}

// --- HTTP Handlers (User) ---

// getProductRecommendationsUserHandler 处理用户获取商品推荐请求。
func (h *AIModelHandler) getProductRecommendationsUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.GetProductRecommendationsRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			h.logger.Warn("GetProductRecommendationsUserHandler: failed to bind query", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}

		// 从上下文获取用户ID并设置到请求中
		userID, exists := c.Get("userID")
		if !exists {
			h.logger.Warn("GetProductRecommendationsUserHandler: user ID not found in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}
		req.UserId = userID.(uint64)

		resp, err := h.svc.GetProductRecommendations(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to get product recommendations", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp.Recommendations)
	}
}

// getRelatedProductsUserHandler 处理用户获取相关商品请求。
func (h *AIModelHandler) getRelatedProductsUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		productID, err := strconv.ParseUint(c.Param("product_id"), 10, 64)
		if err != nil {
			h.logger.Warn("GetRelatedProductsUserHandler: invalid product ID format", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
			return
		}

		var req v1.GetRelatedProductsRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			h.logger.Warn("GetRelatedProductsUserHandler: failed to bind query", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}
		req.ProductId = productID

		resp, err := h.svc.GetRelatedProducts(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to get related products", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp.RelatedProducts)
	}
}

// getPersonalizedFeedUserHandler 处理用户获取个性化内容流请求。
func (h *AIModelHandler) getPersonalizedFeedUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.GetPersonalizedFeedRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			h.logger.Warn("GetPersonalizedFeedUserHandler: failed to bind query", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}

		// 从上下文获取用户ID并设置到请求中
		userID, exists := c.Get("userID")
		if !exists {
			h.logger.Warn("GetPersonalizedFeedUserHandler: user ID not found in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}
		req.UserId = userID.(uint64)

		resp, err := h.svc.GetPersonalizedFeed(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to get personalized feed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp.FeedItems)
	}
}

// analyzeReviewSentimentUserHandler 处理用户分析评论情感请求。
func (h *AIModelHandler) analyzeReviewSentimentUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.AnalyzeReviewSentimentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.logger.Warn("AnalyzeReviewSentimentUserHandler: failed to bind JSON", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := h.svc.AnalyzeReviewSentiment(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to analyze review sentiment", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// --- HTTP Handlers (Public) ---

// recognizeImageContentHandler 处理识别图片内容请求。
func (h *AIModelHandler) recognizeImageContentHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.RecognizeImageContentRequest
		// 尝试从 JSON body 绑定，如果失败则从 form data 绑定 (支持文件上传)
		if err := c.ShouldBindJSON(&req); err != nil {
			// 如果 JSON 绑定失败，尝试 form data
			if err := c.ShouldBind(&req); err != nil {
				h.logger.Warn("RecognizeImageContentHandler: failed to bind request", zap.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body or form data"})
				return
			}
		}

		// TODO: 处理文件上传的 image_data

		resp, err := h.svc.RecognizeImageContent(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to recognize image content", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// searchImageByImageHandler 处理通过图片搜索相似商品请求。
func (h *AIModelHandler) searchImageByImageHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.SearchImageByImageRequest
		// 尝试从 JSON body 绑定，如果失败则从 form data 绑定 (支持文件上传)
		if err := c.ShouldBindJSON(&req); err != nil {
			// 如果 JSON 绑定失败，尝试 form data
			if err := c.ShouldBind(&req); err != nil {
				h.logger.Warn("SearchImageByImageHandler: failed to bind request", zap.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body or form data"})
				return
			}
		}

		// TODO: 处理文件上传的 image_data

		resp, err := h.svc.SearchImageByImage(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to search image by image", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// --- HTTP Handlers (Admin) ---

// deployModelAdminHandler 处理管理员部署AI模型请求。
func (h *AIModelHandler) deployModelAdminHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.DeployModelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.logger.Warn("DeployModelAdminHandler: failed to bind JSON", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := h.svc.DeployModel(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to deploy AI model", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

// getModelStatusAdminHandler 处理管理员获取模型状态请求。
func (h *AIModelHandler) getModelStatusAdminHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		deploymentID := c.Param("deployment_id")

		resp, err := h.svc.GetModelStatus(c.Request.Context(), &v1.GetModelStatusRequest{DeploymentId: deploymentID})
		if err != nil {
			h.logger.Error("Failed to get model status", zap.Error(err), zap.String("deployment_id", deploymentID))
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// retrainModelAdminHandler 处理管理员触发模型重新训练请求。
func (h *AIModelHandler) retrainModelAdminHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.RetrainModelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.logger.Warn("RetrainModelAdminHandler: failed to bind JSON", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := h.svc.RetrainModel(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to retrain AI model", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// getFraudScoreAdminHandler 处理管理员获取欺诈评分请求。
func (h *AIModelHandler) getFraudScoreAdminHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.GetFraudScoreRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.logger.Warn("GetFraudScoreAdminHandler: failed to bind JSON", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := h.svc.GetFraudScore(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to get fraud score", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// extractKeywordsFromTextAdminHandler 处理管理员从文本中提取关键词请求。
func (h *AIModelHandler) extractKeywordsFromTextAdminHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.ExtractKeywordsFromTextRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.logger.Warn("ExtractKeywordsFromTextAdminHandler: failed to bind JSON", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := h.svc.ExtractKeywordsFromText(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to extract keywords from text", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// summarizeTextAdminHandler 处理管理员总结长文本内容请求。
func (h *AIModelHandler) summarizeTextAdminHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.SummarizeTextRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.logger.Warn("SummarizeTextAdminHandler: failed to bind JSON", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := h.svc.SummarizeText(c.Request.Context(), &req)
		if err != nil {
			h.logger.Error("Failed to summarize text", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// --- Middleware ---

// userAuthMiddleware 是一个 Gin 中间件，用于验证用户 JWT Token 的有效性。
func (h *AIModelHandler) userAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Gin context 中获取 JWT Secret，这里简化为硬编码，实际应从配置或服务中获取
		// TODO: 从配置中获取 JWT Secret
		jwtSecret := "your-user-jwt-secret" // 替换为实际的用户服务 JWT Secret

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			h.logger.Warn("UserAuthMiddleware: authorization header is missing")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			h.logger.Warn("UserAuthMiddleware: invalid authorization header format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		claims, err := jwt.ParseToken(tokenString, jwtSecret)
		if err != nil {
			h.logger.Warn("UserAuthMiddleware: invalid or expired token", zap.Error(err))
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
func (h *AIModelHandler) adminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Gin context 中获取 JWT Secret，这里简化为硬编码，实际应从配置或服务中获取
		// TODO: 从配置中获取 JWT Secret
		jwtSecret := "your-admin-jwt-secret" // 替换为实际的管理员服务 JWT Secret

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			h.logger.Warn("AdminAuthMiddleware: authorization header is missing")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			h.logger.Warn("AdminAuthMiddleware: invalid authorization header format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		claims, err := jwt.ParseToken(tokenString, jwtSecret)
		if err != nil {
			h.logger.Warn("AdminAuthMiddleware: invalid or expired token", zap.Error(err))
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
package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ecommerce/api/user/v1"
	"ecommerce/internal/user/service"
	"ecommerce/pkg/jwt"
	"ecommerce/pkg/logging"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RegisterRoutes 注册所有的 HTTP 路由。
func RegisterRoutes(r *gin.Engine, userService *service.UserService) {
	// 健康检查路由
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1/users")

	// 公开路由，无需认证
	api.POST("/register", registerUserHandler(userService))
	api.POST("/login", loginUserHandler(userService))

	// 受保护的路由组，需要 JWT 认证
	protected := api.Group("/")
	protected.Use(authMiddleware(userService.GetJwtSecret()))
	{
		protected.GET("/:user_id", getUserHandler(userService))
		protected.PUT("/:user_id", updateUserHandler(userService))

		// 地址相关的路由
		protected.POST("/:user_id/addresses", createAddressHandler(userService))
		protected.GET("/:user_id/addresses/:address_id", getAddressHandler(userService))
		protected.GET("/:user_id/addresses", listAddressesHandler(userService))
		protected.PUT("/:user_id/addresses/:address_id", updateAddressHandler(userService))
		protected.DELETE("/:user_id/addresses/:address_id", deleteAddressHandler(userService))
	}
}

// --- HTTP Handlers ---

// registerUserHandler 处理用户注册请求。
func registerUserHandler(s *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.RegisterByPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			zap.S().Warnf("registerUserHandler: failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := s.RegisterByPassword(c.Request.Context(), &req)
		if err != nil {
			// service层已记录详细错误, 此处只返回通用错误信息
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

// loginUserHandler 处理用户登录请求。
func loginUserHandler(s *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.LoginByPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			zap.S().Warnf("loginUserHandler: failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resp, err := s.LoginByPassword(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// getUserHandler 处理获取用户信息的请求。
func getUserHandler(s *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}

		resp, err := s.GetUserByID(c.Request.Context(), &v1.GetUserByIDRequest{UserId: userID})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp.User)
	}
}

// updateUserHandler 处理更新用户信息的请求。
func updateUserHandler(s *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}

		var req struct {
			Nickname *string `json:"nickname"`
			Avatar   *string `json:"avatar"`
			Gender   *int32  `json:"gender"`
			Birthday *string `json:"birthday"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			zap.S().Warnf("updateUserHandler: failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		updateReq := &v1.UpdateUserInfoRequest{
			UserId:   userID,
			Nickname: req.Nickname,
			Avatar:   req.Avatar,
			Gender:   req.Gender,
		}

		if req.Birthday != nil {
			birthdayTime, err := time.Parse("2006-01-02", *req.Birthday)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid birthday format. Use YYYY-MM-DD"})
				return
			}
			updateReq.Birthday = timestamppb.New(birthdayTime)
		}

		resp, err := s.UpdateUserInfo(c.Request.Context(), updateReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp.User)
	}
}

// createAddressHandler 处理创建地址的请求。
func createAddressHandler(s *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}

		var req v1.AddAddressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			zap.S().Warnf("createAddressHandler: failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		req.UserId = userID

		resp, err := s.AddAddress(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

// getAddressHandler 处理获取单个地址的请求。
func getAddressHandler(s *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}
		addressID, err := strconv.ParseUint(c.Param("address_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID format"})
			return
		}

		resp, err := s.GetAddress(c.Request.Context(), &v1.GetAddressRequest{UserId: userID, Id: addressID})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// listAddressesHandler 处理获取地址列表的请求。
func listAddressesHandler(s *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}

		resp, err := s.ListAddresses(c.Request.Context(), &v1.ListAddressesRequest{UserId: userID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp.Addresses)
	}
}

// updateAddressHandler 处理更新地址的请求。
func updateAddressHandler(s *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}
		addressID, err := strconv.ParseUint(c.Param("address_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID format"})
			return
		}

		var req v1.UpdateAddressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			zap.S().Warnf("updateAddressHandler: failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		req.UserId = userID
		req.Id = addressID

		resp, err := s.UpdateAddress(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// deleteAddressHandler 处理删除地址的请求。
func deleteAddressHandler(s *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}
		addressID, err := strconv.ParseUint(c.Param("address_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID format"})
			return
		}

		_, err = s.DeleteAddress(c.Request.Context(), &v1.DeleteAddressRequest{UserId: userID, Id: addressID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}

// --- Middleware ---

// authMiddleware 是一个 Gin 中间件，用于验证 JWT Token 的有效性。
func authMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
package middleware

import (
	"net/http"
	"strings"

	"ecommerce/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 中获取 Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "missing authorization header",
			})
			c.Abort()
			return
		}

		// 验证 Bearer Token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 解析 Token
		claims, err := jwt.ParseToken(token, jwtSecret)
		if err != nil {
			zap.S().Warnf("failed to parse token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		// 设置 gRPC metadata (用于转发到后端服务)
		c.Request.Header.Set("x-user-id", claims.UserID)
		c.Request.Header.Set("x-username", claims.Username)

		c.Next()
	}
}

// OptionalAuthMiddleware 可选的认证中间件 (不强制要求登录)
func OptionalAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		claims, err := jwt.ParseToken(token, jwtSecret)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Request.Header.Set("x-user-id", claims.UserID)
		c.Request.Header.Set("x-username", claims.Username)

		c.Next()
	}
}

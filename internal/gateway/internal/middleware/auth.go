package middleware

import (
	"ecommerce/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 定义一个 context key 用于在 Gin 的上下文中传递和获取用户信息
const ContextUserClaimsKey = "userClaims"

// JWTAuthMiddleware 是一个校验 JWT 的 Gin 中-间件
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 HTTP 请求头中获取 Authorization 信息
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			// 如果没有 Authorization 头，返回 401 未授权错误
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "请求头中缺少 Authorization 认证信息",
			})
			c.Abort() // 终止请求处理链
			return
		}

		// Authorization 头的格式是 "Bearer <token>"，我们需要解析出 token 部分
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization 认证信息格式错误，应为 Bearer {token}",
			})
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 使用我们 pkg 包中的函数解析和验证 token
		claims, err := jwt.ParseToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的 Token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Token 验证成功，将解析出的用户信息（claims）存入 Gin 的上下文中
		// 后续的处理函数（Handler）就可以通过 c.Get(ContextUserClaimsKey) 来获取这些信息
		c.Set(ContextUserClaimsKey, claims)

		// 继续处理请求链中的下一个中间件或处理器
		c.Next()
	}
}

package middleware

import (
	"context"
	"net/http"
	"strings"

	"ecommerce/ecommerce/pkg/jwt"
)

// 定义一个 context key 用于传递用户信息
type userClaimsKey struct{}

// JWTAuthMiddleware 是一个校验 JWT 的中间件
func JWTAuthMiddleware(secretKey []byte, unprotectedPaths []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否是免认证路径
			for _, path := range unprotectedPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 从 Header 中获取 Token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			// 解析 Token
			claims, err := jwt.ParseToken(tokenString, secretKey)
			if err != nil {
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Token 验证成功，将用户信息存入 context
			ctx := context.WithValue(r.Context(), userClaimsKey{}, claims)
			r = r.WithContext(ctx)

			// 继续处理请求
			next.ServeHTTP(w, r)
		})
	}
}

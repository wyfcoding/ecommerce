package service

import (
	"context"
	"strings"

	"ecommerce/internal/admin/biz"
	"ecommerce/pkg/jwt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const adminUserIDKey contextKey = "adminUserID"

var (
	// 定义不需要认证的白名单方法
	// 格式为 /package.Service/Method
	authWhitelist = map[string]bool{
		"/admin.v1.Admin/AdminLogin": true,
	}
)

// AuthInterceptor 是一个 gRPC 拦截器，用于处理认证。
type AuthInterceptor struct {
	authUsecase *biz.AuthUsecase
	jwtSecret   string
}

// NewAuthInterceptor 是 AuthInterceptor 的构造函数。
func NewAuthInterceptor(authUC *biz.AuthUsecase, jwtSecret string) *AuthInterceptor {
	return &AuthInterceptor{
		authUsecase: authUC,
		jwtSecret:   jwtSecret,
	}
}

// Auth 是一个 gRPC UnaryInterceptor，用于验证 JWT Token。
func (i *AuthInterceptor) Auth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 1. 检查是否是白名单方法
	if authWhitelist[info.FullMethod] {
		return handler(ctx, req)
	}

	// 2. 从 metadata 中获取 JWT Token
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "缺少认证元数据")
	}
	token := md.Get("authorization")
	if len(token) == 0 || !strings.HasPrefix(token[0], "Bearer ") {
		return nil, status.Errorf(codes.Unauthenticated, "认证头格式不正确")
	}

	// 3. 解析并验证 Token
	tokenString := strings.TrimPrefix(token[0], "Bearer ")
	claims, err := jwt.ParseToken(tokenString, i.jwtSecret)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Token 无效或已过期: %v", err)
	}

	// 4. 将用户ID等信息放入 context，传递给后续的 RPC 处理函数
	ctx = context.WithValue(ctx, adminUserIDKey{}, claims.UserID)

	// 5. 继续处理 RPC 请求
	return handler(ctx, req)
}

package service

import (
	"context"
	"strings"

	"ecommerce/ecommerce/app/admin/internal/biz"
	"ecommerce/ecommerce/pkg/jwt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	uc        *biz.AuthUsecase
	jwtSecret []byte
}

func NewAuthInterceptor(uc *biz.AuthUsecase, jwtSecret string) *AuthInterceptor {
	return &AuthInterceptor{uc: uc, jwtSecret: []byte(jwtSecret)}
}

// Auth 是主要的拦截器方法
func (i *AuthInterceptor) Auth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 白名单：登录接口直接放行
	if info.FullMethod == "/api.admin.v1.Admin/AdminLogin" {
		return handler(ctx, req)
	}

	// 从 metadata 中获取 token
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is required")
	}

	// 解析 token
	tokenString := strings.TrimPrefix(authHeaders[0], "Bearer ")
	claims, err := jwt.ParseToken(tokenString, i.jwtSecret)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// 权限校验
	adminUserID := uint(claims.UserID)
	hasPermission, err := i.uc.CheckPermission(ctx, adminUserID, info.FullMethod)
	if err != nil {
		return nil, status.Error(codes.Internal, "permission check failed")
	}
	if !hasPermission {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	// 权限校验通过，将用户信息注入 context，方便后续 service 使用
	ctx = context.WithValue(ctx, "admin_user_id", adminUserID)

	return handler(ctx, req)
}

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	// 伪代码: 模拟 user 服务的 gRPC 客户端
	// userpb "ecommerce/gen/user/v1"
)

// Claims 定义了 JWT 中存储的自定义声明
type Claims struct {
	UserID uint   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// AuthService 定义了认证服务的业务逻辑接口
type AuthService interface {
	Register(ctx context.Context, email, password string) (uint, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshTokenString string) (newAccessToken string, err error)
	ValidateToken(ctx context.Context, tokenString string) (*Claims, error)
}

// authService 是接口的具体实现
type authService struct {
	logger         *zap.Logger
	jwtSecret      []byte
	jwtIssuer      string
	accessTokenTTL time.Duration
	refreshTokenTTL time.Duration
	bryptCost      int
	// userClient     userpb.UserServiceClient
}

// NewAuthService 创建一个新的 authService 实例
func NewAuthService(logger *zap.Logger, jwtSecret, jwtIssuer string, accessTTL, refreshTTL time.Duration, cost int) AuthService {
	return &authService{
		logger:         logger,
		jwtSecret:      []byte(jwtSecret),
		jwtIssuer:      jwtIssuer,
		accessTokenTTL: accessTTL,
		refreshTokenTTL: refreshTTL,
		bryptCost:      cost,
	}
}

// Register 处理用户注册
func (s *authService) Register(ctx context.Context, email, password string) (uint, error) {
	s.logger.Info("Registering new user", zap.String("email", email))

	// 1. 检查用户是否已存在 (通过调用用户服务)
	// existingUser, err := s.userClient.GetUserByEmail(ctx, &userpb.GetUserByEmailRequest{Email: email})
	// if existingUser != nil { return 0, fmt.Errorf("用户已存在") }

	// 2. 对密码进行哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), s.brycptCost)
	if err != nil {
		return 0, fmt.Errorf("密码哈希失败: %w", err)
	}

	// 3. 调用用户服务创建新用户
	// newUser, err := s.userClient.CreateUser(ctx, &userpb.CreateUserRequest{Email: email, PasswordHash: string(hashedPassword)})
	// if err != nil { return 0, fmt.Errorf("创建用户失败: %w", err) }
	// return newUser.Id, nil

	// 伪代码返回
	return 1, nil
}

// Login 处理用户登录
func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	s.logger.Info("User login attempt", zap.String("email", email))

	// 1. 从用户服务获取用户信息
	// user, err := s.userClient.GetUserByEmail(ctx, &userpb.GetUserByEmailRequest{Email: email})
	// if err != nil || user == nil { return "", "", fmt.Errorf("用户不存在或获取失败") }
	// 伪造用户数据
	user := struct{ ID uint; PasswordHash string }{1, "$2a$12$DGGz.fK6p/y3eixvj9lM9.L2Fp9S4l.E2oGsr2y56JzF.5p3p3p3p"}

	// 2. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", fmt.Errorf("密码错误")
	}

	// 3. 生成 Access Token 和 Refresh Token
	accessToken, err := s.generateToken(user.ID, []string{"user"}, s.accessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("生成 access token 失败: %w", err)
	}
	refreshToken, err := s.generateToken(user.ID, []string{"user"}, s.refreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("生成 refresh token 失败: %w", err)
	}

	return accessToken, refreshToken, nil
}

// RefreshToken 使用刷新令牌获取新的访问令牌
func (s *authService) RefreshToken(ctx context.Context, refreshTokenString string) (string, error) {
	claims, err := s.ValidateToken(ctx, refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("无效的 refresh token: %w", err)
	}

	// 检查是否是过期的 access token (可选，取决于设计)

	// 生成新的 access token
	newAccessToken, err := s.generateToken(claims.UserID, claims.Roles, s.accessTokenTTL)
	if err != nil {
		return "", fmt.Errorf("生成新 access token 失败: %w", err)
	}

	return newAccessToken, nil
}

// ValidateToken 验证令牌的有效性
func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名算法: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("令牌解析失败: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("无效的令牌")
	}
}

// generateToken 是一个内部辅助函数，用于生成 JWT
func (s *authService) generateToken(userID uint, roles []string, ttl time.Duration) (string, error) {
	expirationTime := time.Now().Add(ttl)
	claims := &Claims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    s.jwtIssuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
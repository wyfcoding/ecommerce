package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/pkg/algorithm"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/jwt"
	"github.com/wyfcoding/pkg/security"
)

// Auth 定义了 Auth 相关的服务逻辑。
type Auth struct {
	userRepo  domain.UserRepository
	jwtSecret string
	jwtIssuer string
	jwtExpiry time.Duration
	antiBot   *algorithm.AntiBotDetector
	logger    *slog.Logger
}

// NewAuth 定义了 NewAuth 相关的服务逻辑。
func NewAuth(
	userRepo domain.UserRepository,
	jwtSecret string,
	jwtIssuer string,
	jwtExpiry time.Duration,
	logger *slog.Logger,
) *Auth {
	return &Auth{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtIssuer: jwtIssuer,
		jwtExpiry: jwtExpiry,
		antiBot:   algorithm.NewAntiBotDetector(),
		logger:    logger,
	}
}

// Register 注册一个新用户
func (s *Auth) Register(ctx context.Context, username, password, email, phone string) (uint64, error) {
	// 1. 检查用户名是否已存在
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check existing user", "username", username, "error", err)
		return 0, err
	}
	if existingUser != nil {
		return 0, errors.New("username already exists")
	}

	// 2. 对密码进行哈希处理
	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to hash password", "username", username, "error", err)
		return 0, err
	}

	// 3. 创建用户实体
	user, err := domain.NewUser(username, email, hashedPassword, phone)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create user entity", "username", username, "error", err)
		return 0, err
	}

	// 4. 生成用户ID
	user.ID = uint(idgen.GenID())

	// 5. 保存用户
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.ErrorContext(ctx, "failed to save user", "username", username, "error", err)
		return 0, err
	}
	s.logger.InfoContext(ctx, "user registered successfully", "user_id", user.ID, "username", username)

	return uint64(user.ID), nil
}

// Login 用户登录
func (s *Auth) Login(ctx context.Context, username, password, ip string) (string, int64, error) {
	// 1. 进行机器人行为检测
	behavior := algorithm.UserBehavior{
		UserID:    0,
		IP:        ip,
		Timestamp: time.Now(),
		Action:    "login",
	}
	if isBot, _ := s.antiBot.IsBot(behavior); isBot {
		return "", 0, errors.New("bot detected")
	}

	// 2. 查找用户
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to find user by username", "username", username, "error", err)
		return "", 0, err
	}
	if user == nil {
		return "", 0, errors.New("invalid credentials")
	}

	// 3. 验证密码
	if !security.CheckPassword(password, user.Password) {
		return "", 0, errors.New("invalid credentials")
	}

	// 4. 生成JWT token
	token, err := jwt.GenerateToken(uint64(user.ID), user.Username, s.jwtSecret, s.jwtIssuer, s.jwtExpiry, nil)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate token", "user_id", user.ID, "error", err)
		return "", 0, err
	}
	s.logger.InfoContext(ctx, "user logged in successfully", "user_id", user.ID, "username", username)

	return token, time.Now().Add(s.jwtExpiry).Unix(), nil
}

// CheckBot 检查是否为机器人
func (s *Auth) CheckBot(ctx context.Context, userID uint64, ip string) bool {
	behavior := algorithm.UserBehavior{
		UserID:    userID,
		IP:        ip,
		Timestamp: time.Now(),
		Action:    "check",
	}
	isBot, _ := s.antiBot.IsBot(behavior)
	return isBot
}

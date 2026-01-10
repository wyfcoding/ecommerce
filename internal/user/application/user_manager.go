package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/domain" // AntiBotDetector
	// UserBehavior

	"github.com/wyfcoding/pkg/algorithm/infra"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/jwt"
	"github.com/wyfcoding/pkg/security"
)

// UserManager 处理所有写操作
type UserManager struct {
	userRepo    domain.UserRepository
	addressRepo domain.AddressRepository
	jwtSecret   string
	jwtIssuer   string
	jwtExpiry   time.Duration
	antiBot     *infra.AntiBotDetector
	logger      *slog.Logger
}

func NewUserManager(
	userRepo domain.UserRepository,
	addressRepo domain.AddressRepository,
	jwtSecret string,
	jwtIssuer string,
	jwtExpiry time.Duration,
	antiBot *infra.AntiBotDetector,
	logger *slog.Logger,
) *UserManager {
	return &UserManager{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		jwtSecret:   jwtSecret,
		jwtIssuer:   jwtIssuer,
		jwtExpiry:   jwtExpiry,
		antiBot:     antiBot,
		logger:      logger,
	}
}

// Register 处理新用户注册流程。
// 流程：查重 -> 密码哈希 -> 实体创建 -> 分布式 ID 分配 -> 持久化。
func (m *UserManager) Register(ctx context.Context, req *RegisterRequest) (*domain.User, error) {
	// 1. 账号唯一性检查
	u, err := m.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to check existing user", "username", req.Username, "error", err)
		return nil, err
	}
	if u != nil {
		return nil, errors.New("username already exists")
	}

	// 2. 密码加密存储
	hashed, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 3. 构造领域实体
	user, err := domain.NewUser(req.Username, req.Email, hashed, req.Phone)
	if err != nil {
		return nil, err
	}

	// 4. 分配分布式全局唯一 ID
	user.ID = uint(idgen.GenID())

	// 5. 执行数据持久化
	if err := m.userRepo.Save(ctx, user); err != nil {
		m.logger.ErrorContext(ctx, "failed to save user", "error", err)
		return nil, err
	}

	m.logger.InfoContext(ctx, "user registered successfully", "user_id", user.ID, "username", user.Username)
	return user, nil
}

// Login 处理用户登录认证。
// 流程：机器人检测 -> 用户查找 -> 密码校验 -> 签发 JWT 令牌。
func (m *UserManager) Login(ctx context.Context, username, password, ip string) (string, int64, error) {
	// 1. 基础安全防护：调用算法库进行机器人识别
	behavior := &infra.UserBehavior{
		IP:        ip,
		Timestamp: time.Now(),
		Action:    "login",
		UserID:    0, // Login stage might not have ID yet or use 0
	}
	if isBot, reason := m.antiBot.IsBot(behavior); isBot {
		m.logger.WarnContext(ctx, "bot detected during login", "ip", ip, "username", username, "reason", reason)
		return "", 0, errors.New("bot detected")
	}

	// 2. 查找用户记录
	user, err := m.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", 0, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return "", 0, errors.New("invalid credentials")
	}

	// 3. 比对加密后的密码
	if !security.CheckPassword(password, user.Password) {
		return "", 0, errors.New("invalid credentials")
	}

	// 4. 签发具备权限标识的 JWT 令牌
	token, err := jwt.GenerateToken(uint64(user.ID), user.Username, nil, m.jwtSecret, m.jwtIssuer, m.jwtExpiry)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate token: %w", err)
	}

	m.logger.InfoContext(ctx, "user logged in", "user_id", user.ID, "username", user.Username, "ip", ip)
	return token, time.Now().Add(m.jwtExpiry).Unix(), nil
}

// UpdateProfile 修改用户个人资料。
func (m *UserManager) UpdateProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) (*domain.User, error) {
	user, err := m.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	var birthday *time.Time
	if req.Birthday != "" {
		t, err := time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			return nil, fmt.Errorf("invalid birthday format: %w", err)
		}
		birthday = &t
	}

	user.UpdateProfile(req.Nickname, req.Avatar, req.Gender, birthday)
	if err := m.userRepo.Update(ctx, user); err != nil {
		m.logger.ErrorContext(ctx, "failed to update user profile", "user_id", userID, "error", err)
		return nil, err
	}

	m.logger.InfoContext(ctx, "user profile updated", "user_id", userID)
	return user, nil
}

// AddAddress 为指定用户添加一条新的收货地址。
func (m *UserManager) AddAddress(ctx context.Context, userID uint, req *AddressDTO) (*domain.Address, error) {
	addr := domain.NewAddress(userID, req.RecipientName, req.PhoneNumber, req.Province, req.City, req.District, req.DetailedAddress, req.PostalCode, req.IsDefault)

	if err := m.addressRepo.Save(ctx, addr); err != nil {
		m.logger.ErrorContext(ctx, "failed to save address", "user_id", userID, "error", err)
		return nil, err
	}

	// 如果设为默认地址，则调用仓储逻辑取消该用户其他地址的默认标记
	if req.IsDefault {
		if err := m.addressRepo.SetDefault(ctx, userID, addr.ID); err != nil {
			m.logger.ErrorContext(ctx, "failed to set default address", "user_id", userID, "address_id", addr.ID, "error", err)
		}
	}

	m.logger.InfoContext(ctx, "address added", "user_id", userID, "address_id", addr.ID)
	return addr, nil
}

// UpdateAddress 修改指定的收货地址信息。
func (m *UserManager) UpdateAddress(ctx context.Context, userID, addressID uint, req *AddressDTO) (*domain.Address, error) {
	addr, err := m.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return nil, err
	}
	if addr == nil || addr.UserID != userID {
		return nil, errors.New("address not found or permission denied")
	}

	addr.RecipientName = req.RecipientName
	addr.PhoneNumber = req.PhoneNumber
	addr.Province = req.Province
	addr.City = req.City
	addr.District = req.District
	addr.DetailedAddress = req.DetailedAddress
	addr.PostalCode = req.PostalCode

	if err := m.addressRepo.Update(ctx, addr); err != nil {
		m.logger.ErrorContext(ctx, "failed to update address", "address_id", addressID, "error", err)
		return nil, err
	}

	if req.IsDefault {
		if err := m.addressRepo.SetDefault(ctx, userID, addr.ID); err != nil {
			m.logger.ErrorContext(ctx, "failed to set default address", "user_id", userID, "address_id", addr.ID, "error", err)
		}
		addr.IsDefault = true
	}

	m.logger.InfoContext(ctx, "address updated", "user_id", userID, "address_id", addressID)
	return addr, nil
}

// DeleteAddress 物理删除指定的收货地址。
func (m *UserManager) DeleteAddress(ctx context.Context, userID, addressID uint) error {
	addr, err := m.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return err
	}
	if addr == nil || addr.UserID != userID {
		return errors.New("address not found")
	}

	if err := m.addressRepo.Delete(ctx, addressID); err != nil {
		m.logger.ErrorContext(ctx, "failed to delete address", "address_id", addressID, "error", err)
		return err
	}

	m.logger.InfoContext(ctx, "address deleted", "user_id", userID, "address_id", addressID)
	return nil
}

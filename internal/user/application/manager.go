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

// UserManager 处理所有写操作
type UserManager struct {
	userRepo    domain.UserRepository
	addressRepo domain.AddressRepository
	jwtSecret   string
	jwtIssuer   string
	jwtExpiry   time.Duration
	antiBot     *algorithm.AntiBotDetector
	logger      *slog.Logger
}

func NewUserManager(
	userRepo domain.UserRepository,
	addressRepo domain.AddressRepository,
	jwtSecret string,
	jwtIssuer string,
	jwtExpiry time.Duration,
	antiBot *algorithm.AntiBotDetector,
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

// Register 注册用户
func (m *UserManager) Register(ctx context.Context, req *RegisterRequest) (*domain.User, error) {
	// 1. Check existing
	if u, _ := m.userRepo.FindByUsername(ctx, req.Username); u != nil {
		return nil, errors.New("username already exists")
	}

	// 2. Hash Password
	hashed, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 3. Create Entity
	user, err := domain.NewUser(req.Username, req.Email, hashed, req.Phone)
	if err != nil {
		return nil, err
	}

	// 4. Gen ID
	user.ID = uint(idgen.GenID())

	// 5. Save
	if err := m.userRepo.Save(ctx, user); err != nil {
		m.logger.Error("failed to save user", "err", err)
		return nil, err
	}

	return user, nil
}

// Login 登录
func (m *UserManager) Login(ctx context.Context, username, password, ip string) (string, int64, error) {
	// 1. AntiBot
	behavior := algorithm.UserBehavior{
		IP:        ip,
		Timestamp: time.Now(),
		Action:    "login",
	}
	if isBot, _ := m.antiBot.IsBot(behavior); isBot {
		return "", 0, errors.New("bot detected")
	}

	// 2. Find User
	user, err := m.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", 0, err
	}
	if user == nil {
		return "", 0, errors.New("invalid credentials")
	}

	// 3. Check Password
	if !security.CheckPassword(password, user.Password) {
		return "", 0, errors.New("invalid credentials")
	}

	// 4. Generate Token
	// 【修正】：适配统一的 6 参数签名
	token, err := jwt.GenerateToken(uint64(user.ID), user.Username, nil, m.jwtSecret, m.jwtIssuer, m.jwtExpiry)
	if err != nil {
		return "", 0, err
	}

	return token, time.Now().Add(m.jwtExpiry).Unix(), nil
}

// UpdateProfile 更新信息
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
		if err == nil {
			birthday = &t
		}
	}

	user.UpdateProfile(req.Nickname, req.Avatar, req.Gender, birthday)
	if err := m.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// AddAddress 添加地址
func (m *UserManager) AddAddress(ctx context.Context, userID uint, req *AddressDTO) (*domain.Address, error) {
	addr := domain.NewAddress(userID, req.RecipientName, req.PhoneNumber, req.Province, req.City, req.District, req.DetailedAddress, req.PostalCode, req.IsDefault)

	if err := m.addressRepo.Save(ctx, addr); err != nil {
		return nil, err
	}

	if req.IsDefault {
		_ = m.addressRepo.SetDefault(ctx, userID, addr.ID)
	}

	return addr, nil
}

// UpdateAddress 更新地址
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
		return nil, err
	}

	if req.IsDefault {
		_ = m.addressRepo.SetDefault(ctx, userID, addr.ID)
		addr.IsDefault = true
	}

	return addr, nil
}

// DeleteAddress 删除地址
func (m *UserManager) DeleteAddress(ctx context.Context, userID, addressID uint) error {
	addr, err := m.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return err
	}
	if addr == nil || addr.UserID != userID {
		return errors.New("address not found")
	}
	return m.addressRepo.Delete(ctx, addressID)
}

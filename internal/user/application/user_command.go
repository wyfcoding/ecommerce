package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/pkg/algorithm/infra"
	"github.com/wyfcoding/pkg/cache"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/jwt"
	"github.com/wyfcoding/pkg/security"
)

// UserCommandService 定义用户服务的所有写操作接口
type UserCommandService struct {
	userRepo    domain.UserRepository
	addressRepo domain.AddressRepository
	publisher   domain.EventPublisher // Event Publisher Injected
	cache       cache.Cache           // Cache Injected
	jwtSecret   string
	jwtIssuer   string
	jwtExpiry   time.Duration
	antiBot     *infra.AntiBotDetector
	logger      *slog.Logger
	topic       string // Kafka Topic
}

// NewUserCommandService 创建用户命令服务实例
func NewUserCommandService(
	userRepo domain.UserRepository,
	addressRepo domain.AddressRepository,
	publisher domain.EventPublisher,
	cache cache.Cache,
	topic string,
	jwtSecret string,
	jwtIssuer string,
	jwtExpiry time.Duration,
	antiBot *infra.AntiBotDetector,
	logger *slog.Logger,
) *UserCommandService {
	return &UserCommandService{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		publisher:   publisher,
		cache:       cache,
		topic:       topic,
		jwtSecret:   jwtSecret,
		jwtIssuer:   jwtIssuer,
		jwtExpiry:   jwtExpiry,
		antiBot:     antiBot,
		logger:      logger,
	}
}

// Register 处理注册命令
func (s *UserCommandService) Register(ctx context.Context, cmd *CreateUserCommand) (*domain.User, error) {
	// 1. 唯一性检查
	existingUser, err := s.userRepo.FindByUsername(ctx, cmd.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 2. 密码加密
	hashedPassword, err := security.HashPassword(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 3. 创建领域对象
	user, err := domain.NewUser(
		cmd.Username,
		cmd.Email,
		hashedPassword,
		cmd.Phone,
	)
	if err != nil {
		return nil, err
	}

	// 4. 生成分布式ID
	user.ID = uint(idgen.GenID())

	// 5. 保存
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// 6. 发布事件 (UserCreatedEvent)
	event := domain.UserCreatedEvent{
		Header: domain.EventHeader{
			ID:        fmt.Sprintf("%d", idgen.GenID()),
			TraceID:   "", // Should extract from context
			Source:    "user.service",
			CreatedAt: time.Now(),
		},
		User: user,
	}
	// Best effort delivery. In strict EDA, use Outbox pattern.
	if s.publisher != nil {
		if err := s.publisher.Publish(ctx, s.topic, fmt.Sprintf("%d", user.ID), event); err != nil {
			s.logger.ErrorContext(ctx, "failed to publish UserCreatedEvent", "error", err, "user_id", user.ID)
			// Decide: fail usage or ignore? Typically for non-critical downstream, ignore.
			// But for strict consistency, might want to rollback. GORM transaction required.
		}
	}

	s.logger.InfoContext(ctx, "user registered", "user_id", user.ID)
	return user, nil
}

// Login 处理登录命令
func (s *UserCommandService) Login(ctx context.Context, cmd *LoginCommand, ip string) (*LoginResponse, error) {
	// 1. 反爬与风控检查
	behavior := &infra.UserBehavior{
		IP:        ip,
		Timestamp: time.Now(),
		Action:    "login",
	}
	if isBot, reason := s.antiBot.IsBot(behavior); isBot {
		s.logger.WarnContext(ctx, "bot login attempt blocked", "ip", ip, "reason", reason)
		return nil, errors.New("security check failed")
	}

	// 2. 用户验证
	user, err := s.userRepo.FindByUsername(ctx, cmd.Username)
	if err != nil {
		return nil, err
	}
	if user == nil || !security.CheckPassword(cmd.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	// 3. 生成Token
	token, err := jwt.GenerateToken(uint64(user.ID), user.Username, nil, s.jwtSecret, s.jwtIssuer, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:     token,
		ExpiresIn: int(s.jwtExpiry.Seconds()),
		User:      toUserDTO(user),
	}, nil
}

// UpdateProfile 处理更新资料命令
func (s *UserCommandService) UpdateProfile(ctx context.Context, cmd *UpdateUserCommand) error {
	user, err := s.userRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.UpdateProfile(cmd.Nickname, cmd.Avatar, cmd.Gender, cmd.Birthday)

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Cache Invalidation
	_ = s.cache.Delete(ctx, fmt.Sprintf("user:%d", user.ID))

	// 发布 UserUpdatedEvent
	event := domain.UserUpdatedEvent{
		Header: domain.EventHeader{
			ID:        fmt.Sprintf("%d", idgen.GenID()),
			Source:    "user.service",
			CreatedAt: time.Now(),
		},
		UserID:    user.ID,
		UpdatedAt: time.Now(),
	}
	if s.publisher != nil {
		_ = s.publisher.Publish(ctx, s.topic, fmt.Sprintf("%d", user.ID), event)
	}

	return nil
}

// AddAddress 处理添加地址命令
func (s *UserCommandService) AddAddress(ctx context.Context, cmd *AddAddressCommand) (*domain.Address, error) {
	addr := domain.NewAddress(
		cmd.UserID,
		cmd.RecipientName,
		cmd.PhoneNumber,
		cmd.Province,
		cmd.City,
		cmd.District,
		cmd.DetailedAddress,
		cmd.PostalCode,
		cmd.IsDefault,
	)

	if err := s.addressRepo.Save(ctx, addr); err != nil {
		return nil, fmt.Errorf("failed to save address: %w", err)
	}

	// 如果设为默认，则更新其他地址
	if cmd.IsDefault {
		if err := s.addressRepo.SetDefault(ctx, cmd.UserID, addr.ID); err != nil {
			s.logger.ErrorContext(ctx, "failed to set default address", "user_id", cmd.UserID, "address_id", addr.ID, "error", err)
		}
	}

	return addr, nil
	// Note: Ideally we should invalidate user cache here too if user profile includes addresses
	// But deferring for now to keep it simple or adding explicit delete:
	// _ = s.cache.Delete(ctx, fmt.Sprintf("user:%d", cmd.UserID))
}

// UpdateAddress 处理更新地址命令
func (s *UserCommandService) UpdateAddress(ctx context.Context, cmd *UpdateAddressCommand) error {
	addr, err := s.addressRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if addr == nil || addr.UserID != cmd.UserID {
		return errors.New("address not found or permission denied")
	}

	addr.RecipientName = cmd.RecipientName
	addr.PhoneNumber = cmd.PhoneNumber
	addr.Province = cmd.Province
	addr.City = cmd.City
	addr.District = cmd.District
	addr.DetailedAddress = cmd.DetailedAddress
	addr.PostalCode = cmd.PostalCode

	if err := s.addressRepo.Update(ctx, addr); err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	if cmd.IsDefault {
		if err := s.addressRepo.SetDefault(ctx, cmd.UserID, addr.ID); err != nil {
			s.logger.ErrorContext(ctx, "failed to set default address", "user_id", cmd.UserID, "address_id", addr.ID, "error", err)
		}
	}
	return nil
}

// DeleteAddress 处理删除地址命令
func (s *UserCommandService) DeleteAddress(ctx context.Context, userID, addressID uint) error {
	addr, err := s.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return err
	}
	if addr == nil || addr.UserID != userID {
		return errors.New("address not found")
	}

	if err := s.addressRepo.Delete(ctx, addressID); err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}
	return nil
}

// 辅助函数：转换 User 实体到 DTO
func toUserDTO(user *domain.User) *UserDTO {
	dto := &UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Gender:    user.Gender,
		Birthday:  user.Birthday,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	return dto
}

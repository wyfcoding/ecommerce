package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/pkg/algos/infra"
	"github.com/wyfcoding/pkg/contextx"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/jwt"
	"github.com/wyfcoding/pkg/messagequeue"
	"github.com/wyfcoding/pkg/security"
)

// UserCommandService 定义用户服务的所有写操作接口
type UserCommandService struct {
	userRepo        domain.UserRepository
	addressRepo     domain.AddressRepository
	userReadRepo    domain.UserReadRepository
	addressReadRepo domain.AddressReadRepository
	searchRepo      domain.UserSearchRepository
	publisher       messagequeue.EventPublisher
	jwtSecret       string
	jwtIssuer       string
	jwtExpiry       time.Duration
	antiBot         *infra.AntiBotDetector
	logger          *slog.Logger
	topic           string
}

// NewUserCommandService 创建用户命令服务实例
func NewUserCommandService(
	userRepo domain.UserRepository,
	addressRepo domain.AddressRepository,
	userReadRepo domain.UserReadRepository,
	addressReadRepo domain.AddressReadRepository,
	searchRepo domain.UserSearchRepository,
	publisher messagequeue.EventPublisher,
	topic string,
	jwtSecret string,
	jwtIssuer string,
	jwtExpiry time.Duration,
	antiBot *infra.AntiBotDetector,
	logger *slog.Logger,
) *UserCommandService {
	return &UserCommandService{
		userRepo:        userRepo,
		addressRepo:     addressRepo,
		userReadRepo:    userReadRepo,
		addressReadRepo: addressReadRepo,
		searchRepo:      searchRepo,
		publisher:       publisher,
		topic:           topic,
		jwtSecret:       jwtSecret,
		jwtIssuer:       jwtIssuer,
		jwtExpiry:       jwtExpiry,
		antiBot:         antiBot,
		logger:          logger,
	}
}

// Register 处理注册命令
func (s *UserCommandService) Register(ctx context.Context, cmd CreateUserCommand) (*domain.User, error) {
	if cmd.Username == "" || cmd.Email == "" || cmd.Password == "" {
		return nil, errors.New("username/email/password are required")
	}

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

	// 5. 保存 + Outbox
	if err := s.userRepo.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.Save(txCtx, user); err != nil {
			return fmt.Errorf("failed to save user: %w", err)
		}
		if s.publisher == nil {
			return nil
		}
		event := domain.UserCreatedEvent{
			UserID:     user.ID,
			Username:   user.Username,
			Email:      user.Email,
			Phone:      user.Phone,
			Status:     int8(user.Status),
			CreatedAt:  time.Now().UnixMilli(),
			OccurredOn: time.Now(),
		}
		return s.publisher.PublishInTx(txCtx, contextx.GetTx(txCtx), domain.UserCreatedEventType, fmt.Sprintf("%d", user.ID), event)
	}); err != nil {
		return nil, err
	}

	if s.userReadRepo != nil {
		_ = s.userReadRepo.Save(ctx, user)
	}
	if s.searchRepo != nil {
		_ = s.searchRepo.Index(ctx, user)
	}

	s.logger.InfoContext(ctx, "user registered", "user_id", user.ID)
	return user, nil
}

// Login 处理登录命令
func (s *UserCommandService) Login(ctx context.Context, cmd LoginCommand, ip string) (*LoginResponse, error) {
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
func (s *UserCommandService) UpdateProfile(ctx context.Context, cmd UpdateUserCommand) error {
	if cmd.ID == 0 {
		return errors.New("user id is required")
	}

	var updated *domain.User
	if err := s.userRepo.WithTx(ctx, func(txCtx context.Context) error {
		user, err := s.userRepo.FindByID(txCtx, cmd.ID)
		if err != nil {
			return err
		}
		if user == nil {
			return errors.New("user not found")
		}

		user.UpdateProfile(cmd.Nickname, cmd.Avatar, cmd.Gender, cmd.Birthday)

		if err := s.userRepo.Update(txCtx, user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		updated = user

		if s.publisher == nil {
			return nil
		}
		event := domain.UserUpdatedEvent{
			UserID:     user.ID,
			UpdatedAt:  time.Now().UnixMilli(),
			OccurredOn: time.Now(),
		}
		return s.publisher.PublishInTx(txCtx, contextx.GetTx(txCtx), domain.UserUpdatedEventType, fmt.Sprintf("%d", user.ID), event)
	}); err != nil {
		return err
	}

	if updated != nil {
		if s.userReadRepo != nil {
			_ = s.userReadRepo.Save(ctx, updated)
		}
		if s.searchRepo != nil {
			_ = s.searchRepo.Index(ctx, updated)
		}
	}

	return nil
}

// AddAddress 处理添加地址命令
func (s *UserCommandService) AddAddress(ctx context.Context, cmd AddAddressCommand) (*domain.Address, error) {
	var addr *domain.Address
	err := s.addressRepo.WithTx(ctx, func(txCtx context.Context) error {
		addr = domain.NewAddress(
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

		if err := s.addressRepo.Save(txCtx, addr); err != nil {
			return fmt.Errorf("failed to save address: %w", err)
		}

		if cmd.IsDefault {
			if err := s.addressRepo.SetDefault(txCtx, cmd.UserID, addr.ID); err != nil {
				s.logger.ErrorContext(ctx, "failed to set default address", "user_id", cmd.UserID, "address_id", addr.ID, "error", err)
			}
		}

		if s.publisher == nil {
			return nil
		}

		addedEvent := domain.UserAddressAddedEvent{
			UserID:     cmd.UserID,
			AddressID:  addr.ID,
			IsDefault:  cmd.IsDefault,
			OccurredOn: time.Now(),
		}
		if err := s.publisher.PublishInTx(txCtx, contextx.GetTx(txCtx), domain.UserAddressAddedEventType, fmt.Sprintf("%d", cmd.UserID), addedEvent); err != nil {
			return err
		}

		if cmd.IsDefault {
			defaultEvent := domain.UserAddressDefaultEvent{
				UserID:     cmd.UserID,
				AddressID:  addr.ID,
				OccurredOn: time.Now(),
			}
			return s.publisher.PublishInTx(txCtx, contextx.GetTx(txCtx), domain.UserAddressDefaultEventType, fmt.Sprintf("%d", cmd.UserID), defaultEvent)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.refreshAddressCache(ctx, cmd.UserID)
	return addr, nil
}

// UpdateAddress 处理更新地址命令
func (s *UserCommandService) UpdateAddress(ctx context.Context, cmd UpdateAddressCommand) error {
	return s.addressRepo.WithTx(ctx, func(txCtx context.Context) error {
		addr, err := s.addressRepo.FindByID(txCtx, cmd.ID)
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
		addr.IsDefault = cmd.IsDefault

		if err := s.addressRepo.Update(txCtx, addr); err != nil {
			return fmt.Errorf("failed to update address: %w", err)
		}

		if cmd.IsDefault {
			if err := s.addressRepo.SetDefault(txCtx, cmd.UserID, addr.ID); err != nil {
				s.logger.ErrorContext(ctx, "failed to set default address", "user_id", cmd.UserID, "address_id", addr.ID, "error", err)
			}
		}

		if s.publisher != nil {
			event := domain.UserAddressUpdatedEvent{
				UserID:     cmd.UserID,
				AddressID:  addr.ID,
				IsDefault:  cmd.IsDefault,
				OccurredOn: time.Now(),
			}
			if err := s.publisher.PublishInTx(txCtx, contextx.GetTx(txCtx), domain.UserAddressUpdatedEventType, fmt.Sprintf("%d", cmd.UserID), event); err != nil {
				return err
			}
			if cmd.IsDefault {
				defaultEvent := domain.UserAddressDefaultEvent{
					UserID:     cmd.UserID,
					AddressID:  addr.ID,
					OccurredOn: time.Now(),
				}
				if err := s.publisher.PublishInTx(txCtx, contextx.GetTx(txCtx), domain.UserAddressDefaultEventType, fmt.Sprintf("%d", cmd.UserID), defaultEvent); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// DeleteAddress 处理删除地址命令
func (s *UserCommandService) DeleteAddress(ctx context.Context, userID, addressID uint) error {
	err := s.addressRepo.WithTx(ctx, func(txCtx context.Context) error {
		addr, err := s.addressRepo.FindByID(txCtx, addressID)
		if err != nil {
			return err
		}
		if addr == nil || addr.UserID != userID {
			return errors.New("address not found")
		}

		if err := s.addressRepo.Delete(txCtx, addressID); err != nil {
			return fmt.Errorf("failed to delete address: %w", err)
		}

		if s.publisher != nil {
			event := domain.UserAddressDeletedEvent{
				UserID:     userID,
				AddressID:  addressID,
				OccurredOn: time.Now(),
			}
			return s.publisher.PublishInTx(txCtx, contextx.GetTx(txCtx), domain.UserAddressDeletedEventType, fmt.Sprintf("%d", userID), event)
		}
		return nil
	})
	if err != nil {
		return err
	}

	s.refreshAddressCache(ctx, userID)
	return nil
}

func (s *UserCommandService) refreshAddressCache(ctx context.Context, userID uint) {
	if s.addressReadRepo == nil {
		return
	}
	addrs, err := s.addressRepo.FindByUserID(ctx, userID)
	if err != nil {
		return
	}
	_ = s.addressReadRepo.Save(ctx, userID, addrs)
}

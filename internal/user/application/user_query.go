package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/pkg/algos/infra"
)

// UserQueryService 处理用户模块的所有只读查询操作
type UserQueryService struct {
	userRepo        domain.UserRepository
	addressRepo     domain.AddressRepository
	userReadRepo    domain.UserReadRepository
	addressReadRepo domain.AddressReadRepository
	searchRepo      domain.UserSearchRepository
	antiBot         *infra.AntiBotDetector
	logger          *slog.Logger
}

// NewUserQueryService 初始化并返回一个新的用户查询服务。
func NewUserQueryService(
	userRepo domain.UserRepository,
	addressRepo domain.AddressRepository,
	userReadRepo domain.UserReadRepository,
	addressReadRepo domain.AddressReadRepository,
	searchRepo domain.UserSearchRepository,
	antiBot *infra.AntiBotDetector,
	logger *slog.Logger,
) *UserQueryService {
	return &UserQueryService{
		userRepo:        userRepo,
		addressRepo:     addressRepo,
		userReadRepo:    userReadRepo,
		addressReadRepo: addressReadRepo,
		searchRepo:      searchRepo,
		antiBot:         antiBot,
		logger:          logger,
	}
}

// GetUser 获取指定 ID 用户的完整资料（含关联地址列表）。
func (q *UserQueryService) GetUser(ctx context.Context, userID uint) (*UserDTO, error) {
	if q.userReadRepo != nil {
		if cached, err := q.userReadRepo.GetByID(ctx, userID); err == nil && cached != nil {
			return toUserDTO(cached), nil
		}
	}

	user, err := q.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, err
	}

	if q.userReadRepo != nil {
		_ = q.userReadRepo.Save(ctx, user)
	}

	return toUserDTO(user), nil
}

// CheckBot 基于用户当前行为与 IP 地址执行实时的机器人/爬虫风险判定。
func (q *UserQueryService) CheckBot(ctx context.Context, userID uint64, ip string) bool {
	behavior := &infra.UserBehavior{
		UserID:    userID,
		IP:        ip,
		Timestamp: time.Now(),
		Action:    "check",
	}
	isBot, reason := q.antiBot.IsBot(behavior)
	if isBot {
		q.logger.WarnContext(ctx, "potential bot detected", "user_id", userID, "ip", ip, "reason", reason)
	}
	return isBot
}

// ListAddresses 获取指定用户的所有有效收货地址。
func (q *UserQueryService) ListAddresses(ctx context.Context, userID uint) ([]*AddressDTO, error) {
	if q.addressReadRepo != nil {
		if cached, err := q.addressReadRepo.GetByUserID(ctx, userID); err == nil && len(cached) > 0 {
			return toAddressDTOs(cached), nil
		}
	}

	addrs, err := q.addressRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if q.addressReadRepo != nil {
		_ = q.addressReadRepo.Save(ctx, userID, addrs)
	}
	return toAddressDTOs(addrs), nil
}

// GetAddress 安全获取特定的收货地址详情。
func (q *UserQueryService) GetAddress(ctx context.Context, userID, addressID uint) (*AddressDTO, error) {
	if q.addressReadRepo != nil {
		if cached, err := q.addressReadRepo.GetByUserID(ctx, userID); err == nil && len(cached) > 0 {
			for _, addr := range cached {
				if addr != nil && addr.ID == addressID {
					return toAddressDTO(addr), nil
				}
			}
		}
	}

	addr, err := q.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return nil, err
	}
	if addr != nil && addr.UserID != userID {
		q.logger.WarnContext(ctx, "unauthorized address access attempt", "user_id", userID, "target_address_id", addressID)
		return nil, nil
	}
	if addr == nil {
		return nil, nil
	}
	return toAddressDTO(addr), nil
}

// SearchUsers 用户搜索
func (q *UserQueryService) SearchUsers(ctx context.Context, keyword string, limit, offset int) ([]*UserDTO, int64, error) {
	if q.searchRepo == nil {
		return nil, 0, nil
	}
	users, total, err := q.searchRepo.Search(ctx, keyword, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]*UserDTO, len(users))
	for i, u := range users {
		dtos[i] = toUserDTO(u)
	}
	return dtos, total, nil
}

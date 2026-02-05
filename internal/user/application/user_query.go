// Package application 提供了用户模块的业务逻辑处理。
package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/pkg/algorithm/infra"
	"github.com/wyfcoding/pkg/cache"
)

// UserQuery 处理用户模块的所有只读查询操作，集成了安全审计与行为分析逻辑。
type UserQuery struct {
	userRepo    domain.UserRepository    // 用户基础信息仓储
	addressRepo domain.AddressRepository // 用户地址信息仓储
	cache       cache.Cache              // Cache Injected
	antiBot     *infra.AntiBotDetector   // 机器人检测引擎
	logger      *slog.Logger             // 结构化日志记录器
}

// NewUserQuery 初始化并返回一个新的用户查询服务。
func NewUserQuery(
	userRepo domain.UserRepository,
	addressRepo domain.AddressRepository,
	cache cache.Cache,
	antiBot *infra.AntiBotDetector,
	logger *slog.Logger,
) *UserQuery {
	return &UserQuery{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		cache:       cache,
		antiBot:     antiBot,
		logger:      logger,
	}
}

// GetUser 获取指定 ID 用户的完整资料（含关联地址列表）。
// GetUser 获取指定 ID 用户的完整资料（含关联地址列表）。
func (q *UserQuery) GetUser(ctx context.Context, userID uint) (*UserDTO, error) {
	cacheKey := fmt.Sprintf("user:%d", userID)
	var dto UserDTO

	// 1. Try Cache
	if err := q.cache.Get(ctx, cacheKey, &dto); err == nil {
		return &dto, nil
	}

	// 2. DB Fallback
	user, err := q.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	res := toUserDTO(user)

	// 3. Set Cache
	if err := q.cache.Set(ctx, cacheKey, res, time.Hour); err != nil {
		q.logger.WarnContext(ctx, "failed to set user cache", "key", cacheKey, "error", err)
	}

	return res, nil
}

// CheckBot 基于用户当前行为与 IP 地址执行实时的机器人/爬虫风险判定。
func (q *UserQuery) CheckBot(ctx context.Context, userID uint64, ip string) bool {
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
func (q *UserQuery) ListAddresses(ctx context.Context, userID uint) ([]*AddressDTO, error) {
	addrs, err := q.addressRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	dtos := make([]*AddressDTO, len(addrs))
	for i, addr := range addrs {
		dtos[i] = toAddressDTO(addr)
	}
	return dtos, nil
}

// GetAddress 安全获取特定的收货地址详情。
func (q *UserQuery) GetAddress(ctx context.Context, userID, addressID uint) (*AddressDTO, error) {
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

func toAddressDTO(addr *domain.Address) *AddressDTO {
	return &AddressDTO{
		ID:              addr.ID,
		UserID:          addr.UserID,
		RecipientName:   addr.RecipientName,
		PhoneNumber:     addr.PhoneNumber,
		Province:        addr.Province,
		City:            addr.City,
		District:        addr.District,
		DetailedAddress: addr.DetailedAddress,
		PostalCode:      addr.PostalCode,
		IsDefault:       addr.IsDefault,
		CreatedAt:       addr.CreatedAt,
		UpdatedAt:       addr.UpdatedAt,
	}
}

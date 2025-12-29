package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// UserQuery 处理所有读操作
type UserQuery struct {
	userRepo    domain.UserRepository
	addressRepo domain.AddressRepository
	antiBot     *algorithm.AntiBotDetector
	logger      *slog.Logger
}

func NewUserQuery(
	userRepo domain.UserRepository,
	addressRepo domain.AddressRepository,
	antiBot *algorithm.AntiBotDetector,
	logger *slog.Logger,
) *UserQuery {
	return &UserQuery{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		antiBot:     antiBot,
		logger:      logger,
	}
}

// GetUser 获取用户
func (q *UserQuery) GetUser(ctx context.Context, userID uint) (*domain.User, error) {
	return q.userRepo.FindByID(ctx, userID)
}

// CheckBot 检查
func (q *UserQuery) CheckBot(ctx context.Context, userID uint64, ip string) bool {
	behavior := algorithm.UserBehavior{
		UserID:    userID,
		IP:        ip,
		Timestamp: time.Now(),
		Action:    "check",
	}
	isBot, _ := q.antiBot.IsBot(behavior)
	return isBot
}

// ListAddresses 获取地址列表
func (q *UserQuery) ListAddresses(ctx context.Context, userID uint) ([]*domain.Address, error) {
	return q.addressRepo.FindByUserID(ctx, userID)
}

// GetAddress 获取单个地址
func (q *UserQuery) GetAddress(ctx context.Context, userID, addressID uint) (*domain.Address, error) {
	addr, err := q.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return nil, err
	}
	if addr != nil && addr.UserID != userID {
		return nil, nil // Hide others address
	}
	return addr, nil
}

package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
)

// User 用户服务门面（Facade）
// 聚合了 Auth, Profile, Address 等子服务
type User struct {
	Auth    *Auth
	Profile *ProfileService
	Address *AddressService
	logger  *slog.Logger
}

// NewUser 定义了 NewUser 相关的服务逻辑。
func NewUser(
	auth *Auth,
	profile *ProfileService,
	address *AddressService,
	logger *slog.Logger,
) *User {
	return &User{
		Auth:    auth,
		Profile: profile,
		Address: address,
		logger:  logger,
	}
}

// --- 委托给子服务的方法 ---
// 这里的委托方法是为了兼容旧的接口调用，或者是为了提供一个统一的入口。
// 如果 gRPC Server 还是直接调用 application.Register，那么我们可以在这里保留同名方法。

func (s *User) Register(ctx context.Context, username, password, email, phone string) (uint64, error) {
	return s.Auth.Register(ctx, username, password, email, phone)
}

func (s *User) Login(ctx context.Context, username, password, ip string) (string, int64, error) {
	return s.Auth.Login(ctx, username, password, ip)
}

func (s *User) CheckBot(ctx context.Context, userID uint64, ip string) bool {
	return s.Auth.CheckBot(ctx, userID, ip)
}

func (s *User) GetUser(ctx context.Context, userID uint64) (*domain.User, error) {
	return s.Profile.GetUser(ctx, userID)
}

func (s *User) UpdateProfile(ctx context.Context, userID uint64, nickname, avatar string, gender int8, birthday *time.Time) (*domain.User, error) {
	return s.Profile.UpdateProfile(ctx, userID, nickname, avatar, gender, birthday)
}

func (s *User) AddAddress(ctx context.Context, userID uint64, name, phone, province, city, district, detail string, isDefault bool) (*domain.Address, error) {
	return s.Address.AddAddress(ctx, userID, name, phone, province, city, district, detail, isDefault)
}

func (s *User) ListAddresses(ctx context.Context, userID uint64) ([]*domain.Address, error) {
	return s.Address.ListAddresses(ctx, userID)
}

func (s *User) UpdateAddress(ctx context.Context, userID, addressID uint64, name, phone, province, city, district, detail string, isDefault bool) (*domain.Address, error) {
	return s.Address.UpdateAddress(ctx, userID, addressID, name, phone, province, city, district, detail, isDefault)
}

func (s *User) DeleteAddress(ctx context.Context, userID, addressID uint64) error {
	return s.Address.DeleteAddress(ctx, userID, addressID)
}

func (s *User) GetAddress(ctx context.Context, userID, addressID uint64) (*domain.Address, error) {
	return s.Address.GetAddress(ctx, userID, addressID)
}

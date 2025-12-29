package application

import (
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// UserService 用户服务门面
type UserService struct {
	Manager *UserManager
	Query   *UserQuery
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo domain.UserRepository,
	addressRepo domain.AddressRepository,
	jwtSecret string,
	jwtIssuer string,
	jwtExpiry time.Duration,
	logger *slog.Logger,
) *UserService {
	// AntiBot needs instantiation
	antiBot := algorithm.NewAntiBotDetector()

	return &UserService{
		Manager: NewUserManager(userRepo, addressRepo, jwtSecret, jwtIssuer, jwtExpiry, antiBot, logger),
		Query:   NewUserQuery(userRepo, addressRepo, antiBot, logger),
	}
}

// --- DTOs ---

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   int8   `json:"gender"`
	Birthday string `json:"birthday"` // "2006-01-02"
}

type AddressDTO struct {
	RecipientName   string `json:"recipient_name" binding:"required"`
	PhoneNumber     string `json:"phone_number" binding:"required"`
	Province        string `json:"province" binding:"required"`
	City            string `json:"city" binding:"required"`
	District        string `json:"district" binding:"required"`
	DetailedAddress string `json:"detailed_address" binding:"required"`
	PostalCode      string `json:"postal_code"`
	IsDefault       bool   `json:"is_default"`
}

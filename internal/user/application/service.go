package application

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/ecommerce/pkg/algorithm"
	"github.com/wyfcoding/ecommerce/pkg/hash"
	"github.com/wyfcoding/ecommerce/pkg/idgen"
	"github.com/wyfcoding/ecommerce/pkg/jwt"
)

// UserApplicationService defines the application service for user operations.
type UserApplicationService struct {
	userRepo    domain.UserRepository
	addressRepo domain.AddressRepository
	jwtSecret   string
	jwtIssuer   string
	jwtExpiry   time.Duration
	antiBot     *algorithm.AntiBotDetector
}

// NewUserApplicationService creates a new UserApplicationService.
func NewUserApplicationService(
	userRepo domain.UserRepository,
	addressRepo domain.AddressRepository,
	jwtSecret string,
	jwtIssuer string,
	jwtExpiry time.Duration,
) *UserApplicationService {
	return &UserApplicationService{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		jwtSecret:   jwtSecret,
		jwtIssuer:   jwtIssuer,
		jwtExpiry:   jwtExpiry,
		antiBot:     algorithm.NewAntiBotDetector(),
	}
}

// Register registers a new user.
func (s *UserApplicationService) Register(ctx context.Context, username, password, email, phone string) (uint64, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return 0, err
	}
	if existingUser != nil {
		return 0, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return 0, err
	}

	// Create user entity
	user, err := domain.NewUser(username, email, hashedPassword, phone)
	if err != nil {
		return 0, err
	}

	// Generate ID
	// Note: gorm.Model IDs are uint, idgen returns uint64.
	user.ID = uint(idgen.GenID())

	// Save user
	if err := s.userRepo.Save(ctx, user); err != nil {
		return 0, err
	}

	return uint64(user.ID), nil
}

// Login logs in a user and returns a JWT token.
func (s *UserApplicationService) Login(ctx context.Context, username, password, ip string) (string, int64, error) {
	// Check for bot behavior
	behavior := algorithm.UserBehavior{
		UserID:    0, // Unknown user ID at this point
		IP:        ip,
		Timestamp: time.Now(),
		Action:    "login",
	}
	if isBot, _ := s.antiBot.IsBot(behavior); isBot {
		return "", 0, errors.New("bot detected")
	}
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", 0, err
	}
	if user == nil {
		return "", 0, errors.New("invalid credentials")
	}

	if !hash.CheckPasswordHash(password, user.Password) {
		return "", 0, errors.New("invalid credentials")
	}

	token, err := jwt.GenerateToken(uint64(user.ID), user.Username, s.jwtSecret, s.jwtIssuer, s.jwtExpiry, nil)
	if err != nil {
		return "", 0, err
	}

	return token, time.Now().Add(s.jwtExpiry).Unix(), nil
}

// CheckBot checks if the request is from a bot.
func (s *UserApplicationService) CheckBot(ctx context.Context, userID uint64, ip string) bool {
	behavior := algorithm.UserBehavior{
		UserID:    userID,
		IP:        ip,
		Timestamp: time.Now(),
		Action:    "check",
	}
	isBot, _ := s.antiBot.IsBot(behavior)
	return isBot
}

// GetUser gets a user by ID.
func (s *UserApplicationService) GetUser(ctx context.Context, userID uint64) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, uint(userID))
}

// UpdateProfile updates a user's profile.
func (s *UserApplicationService) UpdateProfile(ctx context.Context, userID uint64, nickname, avatar string, gender int8, birthday *time.Time) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	user.UpdateProfile(nickname, avatar, gender, birthday)

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// AddAddress adds a new address for a user.
func (s *UserApplicationService) AddAddress(ctx context.Context, userID uint64, name, phone, province, city, district, detail string, isDefault bool) (*domain.Address, error) {
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	address := domain.NewAddress(uint(userID), name, phone, province, city, district, detail, "", isDefault)
	address.ID = uint(idgen.GenID())

	// If isDefault is true, handle default logic
	if isDefault {
		// We need to save first or use domain logic.
		// Since SetDefault in repo updates DB, we should save first.
	}

	if err := s.addressRepo.Save(ctx, address); err != nil {
		return nil, err
	}

	if isDefault {
		// Now that it's saved, we can ensure it's the default
		if err := s.addressRepo.SetDefault(ctx, uint(userID), address.ID); err != nil {
			return nil, err
		}
		address.IsDefault = true
	}

	return address, nil
}

// ListAddresses lists all addresses for a user.
func (s *UserApplicationService) ListAddresses(ctx context.Context, userID uint64) ([]*domain.Address, error) {
	return s.addressRepo.FindByUserID(ctx, uint(userID))
}

// UpdateAddress updates an address.
func (s *UserApplicationService) UpdateAddress(ctx context.Context, userID, addressID uint64, name, phone, province, city, district, detail string, isDefault bool) (*domain.Address, error) {
	address, err := s.addressRepo.FindByID(ctx, uint(addressID))
	if err != nil {
		return nil, err
	}
	if address == nil || address.UserID != uint(userID) {
		return nil, errors.New("address not found")
	}

	if name != "" {
		address.RecipientName = name
	}
	if phone != "" {
		address.PhoneNumber = phone
	}
	if province != "" {
		address.Province = province
	}
	if city != "" {
		address.City = city
	}
	if district != "" {
		address.District = district
	}
	if detail != "" {
		address.DetailedAddress = detail
	}

	// address.UpdatedAt is handled by GORM or we can set it explicitly if needed,
	// but gorm.Model handles it on Save/Update.
	// However, domain logic might want to set it.
	// address.UpdatedAt = time.Now() // gorm.Model handles this

	if err := s.addressRepo.Update(ctx, address); err != nil {
		return nil, err
	}

	if isDefault {
		if err := s.addressRepo.SetDefault(ctx, uint(userID), uint(addressID)); err != nil {
			return nil, err
		}
		address.IsDefault = true
	}

	return address, nil
}

// DeleteAddress deletes an address.
func (s *UserApplicationService) DeleteAddress(ctx context.Context, userID, addressID uint64) error {
	address, err := s.addressRepo.FindByID(ctx, uint(addressID))
	if err != nil {
		return err
	}
	if address == nil || address.UserID != uint(userID) {
		return errors.New("address not found")
	}

	return s.addressRepo.Delete(ctx, uint(addressID))
}

// GetAddress gets an address by ID.
func (s *UserApplicationService) GetAddress(ctx context.Context, userID, addressID uint64) (*domain.Address, error) {
	address, err := s.addressRepo.FindByID(ctx, uint(addressID))
	if err != nil {
		return nil, err
	}
	if address == nil || address.UserID != uint(userID) {
		return nil, errors.New("address not found")
	}
	return address, nil
}

package application

import (
	"context"
	"errors" // 导入标准错误处理库。
	"time"   // 导入时间库。

	"log/slog" // 导入结构化日志库。

	"github.com/wyfcoding/ecommerce/internal/user/domain" // 导入用户领域的领域接口和实体。
	"github.com/wyfcoding/ecommerce/pkg/algorithm"        // 导入算法包，用于反机器人检测。
	"github.com/wyfcoding/ecommerce/pkg/hash"             // 导入哈希工具包，用于密码处理。
	"github.com/wyfcoding/ecommerce/pkg/idgen"            // 导入ID生成器。
	"github.com/wyfcoding/ecommerce/pkg/jwt"              // 导入JWT工具包。
)

// UserApplicationService 定义了用户操作的应用服务。
// 它协调领域层和基础设施层，处理用户注册、登录、资料更新、地址管理等业务逻辑。
type UserApplicationService struct {
	userRepo    domain.UserRepository      // 用户仓储接口。
	addressRepo domain.AddressRepository   // 地址仓储接口。
	jwtSecret   string                     // JWT签名密钥。
	jwtIssuer   string                     // JWT签发者。
	jwtExpiry   time.Duration              // JWT有效期。
	antiBot     *algorithm.AntiBotDetector // 反机器人检测器。
	logger      *slog.Logger               // 日志记录器。
}

// NewUserApplicationService 创建一个新的 UserApplicationService 实例。
func NewUserApplicationService(
	userRepo domain.UserRepository,
	addressRepo domain.AddressRepository,
	jwtSecret string,
	jwtIssuer string,
	jwtExpiry time.Duration,
	logger *slog.Logger,
) *UserApplicationService {
	return &UserApplicationService{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		jwtSecret:   jwtSecret,
		jwtIssuer:   jwtIssuer,
		jwtExpiry:   jwtExpiry,
		antiBot:     algorithm.NewAntiBotDetector(), // 初始化反机器人检测器。
		logger:      logger,
	}
}

// Register 注册一个新用户。
// ctx: 上下文。
// username: 用户名。
// password: 密码。
// email: 邮箱。
// phone: 手机号。
// 返回新用户的ID和可能发生的错误。
func (s *UserApplicationService) Register(ctx context.Context, username, password, email, phone string) (uint64, error) {
	// 1. 检查用户名是否已存在。
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check existing user", "username", username, "error", err)
		return 0, err
	}
	if existingUser != nil {
		return 0, errors.New("username already exists")
	}

	// 2. 对密码进行哈希处理。
	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to hash password", "username", username, "error", err)
		return 0, err
	}

	// 3. 创建用户实体。
	user, err := domain.NewUser(username, email, hashedPassword, phone)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create user entity", "username", username, "error", err)
		return 0, err
	}

	// 4. 生成用户ID。
	// 注意：gorm.Model 中的 ID 字段类型为 uint，idgen.GenID() 返回 uint64。
	user.ID = uint(idgen.GenID())

	// 5. 保存用户。
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.ErrorContext(ctx, "failed to save user", "username", username, "error", err)
		return 0, err
	}
	s.logger.InfoContext(ctx, "user registered successfully", "user_id", user.ID, "username", username)

	return uint64(user.ID), nil
}

// Login 用户登录，并返回JWT token。
// ctx: 上下文。
// username: 用户名。
// password: 密码。
// ip: 用户IP地址。
// 返回JWT token、过期时间戳和可能发生的错误。
func (s *UserApplicationService) Login(ctx context.Context, username, password, ip string) (string, int64, error) {
	// 1. 进行机器人行为检测。
	behavior := algorithm.UserBehavior{
		UserID:    0, // 在登录阶段，用户ID可能未知。
		IP:        ip,
		Timestamp: time.Now(),
		Action:    "login",
	}
	if isBot, _ := s.antiBot.IsBot(behavior); isBot {
		return "", 0, errors.New("bot detected")
	}

	// 2. 查找用户。
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to find user by username", "username", username, "error", err)
		return "", 0, err
	}
	if user == nil {
		return "", 0, errors.New("invalid credentials")
	}

	// 3. 验证密码。
	if !hash.CheckPasswordHash(password, user.Password) {
		return "", 0, errors.New("invalid credentials")
	}

	// 4. 生成JWT token。
	token, err := jwt.GenerateToken(uint64(user.ID), user.Username, s.jwtSecret, s.jwtIssuer, s.jwtExpiry, nil)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate token", "user_id", user.ID, "error", err)
		return "", 0, err
	}
	s.logger.InfoContext(ctx, "user logged in successfully", "user_id", user.ID, "username", username)

	return token, time.Now().Add(s.jwtExpiry).Unix(), nil
}

// CheckBot 检查请求是否来自机器人。
// ctx: 上下文。
// userID: 用户ID。
// ip: IP地址。
// 返回布尔值（是否为机器人）和可能发生的错误。
func (s *UserApplicationService) CheckBot(ctx context.Context, userID uint64, ip string) bool {
	behavior := algorithm.UserBehavior{
		UserID:    userID,
		IP:        ip,
		Timestamp: time.Now(),
		Action:    "check",
	}
	isBot, _ := s.antiBot.IsBot(behavior) // 忽略错误，返回结果。
	return isBot
}

// GetUser 根据用户ID获取用户信息。
// ctx: 上下文。
// userID: 用户ID。
// 返回User实体和可能发生的错误。
func (s *UserApplicationService) GetUser(ctx context.Context, userID uint64) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, uint(userID))
}

// UpdateProfile 更新用户个人资料。
// ctx: 上下文。
// userID: 用户ID。
// nickname: 昵称。
// avatar: 头像URL。
// gender: 性别。
// birthday: 生日。
// 返回更新后的User实体和可能发生的错误。
func (s *UserApplicationService) UpdateProfile(ctx context.Context, userID uint64, nickname, avatar string, gender int8, birthday *time.Time) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 调用用户实体方法更新个人资料。
	user.UpdateProfile(nickname, avatar, gender, birthday)

	// 保存更新后的用户。
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.ErrorContext(ctx, "failed to update profile", "user_id", userID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "user profile updated successfully", "user_id", userID)

	return user, nil
}

// AddAddress 为用户添加一个新地址。
// ctx: 上下文。
// userID: 用户ID。
// name: 收货人姓名。
// phone: 手机号。
// province, city, district, detail: 地址详情。
// isDefault: 是否设为默认地址。
// 返回新添加的Address实体和可能发生的错误。
func (s *UserApplicationService) AddAddress(ctx context.Context, userID uint64, name, phone, province, city, district, detail string, isDefault bool) (*domain.Address, error) {
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 创建地址实体。
	address := domain.NewAddress(uint(userID), name, phone, province, city, district, detail, "", isDefault) // postalCode为空。
	address.ID = uint(idgen.GenID())                                                                         // 生成地址ID。

	// 保存地址。
	if err := s.addressRepo.Save(ctx, address); err != nil {
		s.logger.ErrorContext(ctx, "failed to add address", "user_id", userID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "address added successfully", "user_id", userID, "address_id", address.ID)

	// 如果设置为默认地址，则更新用户的默认地址。
	if isDefault {
		// 在保存新地址后，调用SetDefault方法将其设置为默认地址，并取消其他地址的默认状态。
		if err := s.addressRepo.SetDefault(ctx, uint(userID), address.ID); err != nil {
			return nil, err
		}
		address.IsDefault = true // 更新地址实体的默认状态。
	}

	return address, nil
}

// ListAddresses 列出指定用户的所有地址。
// ctx: 上下文。
// userID: 用户ID。
// 返回地址实体列表和可能发生的错误。
func (s *UserApplicationService) ListAddresses(ctx context.Context, userID uint64) ([]*domain.Address, error) {
	return s.addressRepo.FindByUserID(ctx, uint(userID))
}

// UpdateAddress 更新用户地址。
// ctx: 上下文。
// userID: 用户ID。
// addressID: 地址ID。
// name, phone, province, city, district, detail: 新的地址信息。
// isDefault: 是否设为默认地址。
// 返回更新后的Address实体和可能发生的错误。
func (s *UserApplicationService) UpdateAddress(ctx context.Context, userID, addressID uint64, name, phone, province, city, district, detail string, isDefault bool) (*domain.Address, error) {
	address, err := s.addressRepo.FindByID(ctx, uint(addressID))
	if err != nil {
		return nil, err
	}
	// 验证地址是否存在且属于该用户。
	if address == nil || address.UserID != uint(userID) {
		return nil, errors.New("address not found or not owned by user")
	}

	// 根据传入的非空值更新地址字段。
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
	// 备注：gorm.Model 会自动处理 UpdatedAt 字段。

	// 保存更新后的地址。
	if err := s.addressRepo.Update(ctx, address); err != nil {
		s.logger.ErrorContext(ctx, "failed to update address", "user_id", userID, "address_id", addressID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "address updated successfully", "user_id", userID, "address_id", addressID)

	// 如果设置为默认地址，则更新用户的默认地址。
	if isDefault {
		if err := s.addressRepo.SetDefault(ctx, uint(userID), uint(addressID)); err != nil {
			return nil, err
		}
		address.IsDefault = true // 更新地址实体的默认状态。
	}

	return address, nil
}

// DeleteAddress 删除用户地址。
// ctx: 上下文。
// userID: 用户ID。
// addressID: 地址ID。
// 返回可能发生的错误。
func (s *UserApplicationService) DeleteAddress(ctx context.Context, userID, addressID uint64) error {
	address, err := s.addressRepo.FindByID(ctx, uint(addressID))
	if err != nil {
		return err
	}
	// 验证地址是否存在且属于该用户。
	if address == nil || address.UserID != uint(userID) {
		return errors.New("address not found or not owned by user")
	}

	// 调用仓储接口删除地址。
	if err := s.addressRepo.Delete(ctx, uint(addressID)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete address", "user_id", userID, "address_id", addressID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "address deleted successfully", "user_id", userID, "address_id", addressID)
	return nil
}

// GetAddress 根据ID获取用户地址。
// ctx: 上下文。
// userID: 用户ID。
// addressID: 地址ID。
// 返回地址实体和可能发生的错误。
func (s *UserApplicationService) GetAddress(ctx context.Context, userID, addressID uint64) (*domain.Address, error) {
	address, err := s.addressRepo.FindByID(ctx, uint(addressID))
	if err != nil {
		return nil, err
	}
	// 验证地址是否存在且属于该用户。
	if address == nil || address.UserID != uint(userID) {
		return nil, errors.New("address not found or not owned by user")
	}
	return address, nil
}

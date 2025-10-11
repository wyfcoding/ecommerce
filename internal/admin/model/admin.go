package model

import (
	"context"
	"errors"
	"time"

	"ecommerce/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrAdminUserNotFound      = errors.New("admin user not found")
	ErrAdminPasswordIncorrect = errors.New("incorrect admin password")
)

// AdminUser 是管理员用户的业务领域模型。
type AdminUser struct {
	ID       uint32
	Username string
	Password string
	Name     string
	Status   int32
}

// AuthRepo 定义了认证数据仓库的接口。
type AuthRepo interface {
	GetAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error)
	GetAdminUserByID(ctx context.Context, id uint32) (*AdminUser, error)
}

// AuthUsecase 封装了认证相关的业务逻辑。
type AuthUsecase struct {
	repo      AuthRepo
	jwtSecret string
	jwtIssuer string
	jwtExpire time.Duration
}

// NewAuthUsecase 是 AuthUsecase 的构造函数。
func NewAuthUsecase(repo AuthRepo, jwtSecret, jwtIssuer string, jwtExpire time.Duration) *AuthUsecase {
	return &AuthUsecase{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtIssuer: jwtIssuer,
		jwtExpire: jwtExpire,
	}
}

// AdminLogin 负责管理员登录的业务逻辑。
func (uc *AuthUsecase) AdminLogin(ctx context.Context, username, password string) (string, error) {
	user, err := uc.repo.GetAdminUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrAdminUserNotFound
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", ErrAdminPasswordIncorrect
	}

	// 检查用户状态
	if user.Status != 1 { // 假设 1 为正常状态
		return "", errors.New("admin account is disabled or abnormal")
	}

	claims := jwt.CustomClaims{
		UserID:   uint64(user.ID), // AdminUser ID 是 uint32 类型，转换为 uint64 以用于 CustomClaims
		Username: user.Username,
	}
	claims.ExpiresAt = time.Now().Add(uc.jwtExpire).Unix()
	claims.Issuer = uc.jwtIssuer

	token, err := jwt.GenerateToken(uint64(user.ID), user.Username, uc.jwtSecret, uc.jwtIssuer, uc.jwtExpire, jwt.SigningMethodHS256)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetJwtSecret 返回 JWT 密钥。
func (uc *AuthUsecase) GetJwtSecret() string {
	return uc.jwtSecret
}

// GetAdminUserByID 负责根据ID获取管理员用户信息的业务逻辑。
func (uc *AuthUsecase) GetAdminUserByID(ctx context.Context, id uint32) (*AdminUser, error) {
	user, err := uc.repo.GetAdminUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// Transaction 定义了事务管理器接口。
type Transaction interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// AdminRepo 定义了管理员数据仓库需要实现的接口。
type AdminRepo interface {
	CreateAdminUser(ctx context.Context, user *AdminUser) (*AdminUser, error)
	GetAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error)
}

// Spu 是商品标准产品单元的业务领域模型。
type Spu struct {
	ID            uint64
	CategoryID    uint64
	BrandID       uint64
	Title         string
	SubTitle      string
	MainImage     string
	GalleryImages []string
	DetailHTML    string
	Status        int32
}

// Sku 是商品库存单位的业务领域模型。
type Sku struct {
	ID            uint64
	SpuID         uint64
	Title         string
	Price         uint64
	OriginalPrice uint64
	Stock         uint32
	Image         string
	Specs         map[string]string
	Status        int32
}

// ProductClient 定义了管理员服务依赖的商品服务客户端接口。
type ProductClient interface {
	CreateProduct(ctx context.Context, spu *Spu, skus []*Sku) (*Spu, []*Sku, error)
	UpdateProduct(ctx context.Context, spu *Spu, skus []*Sku) (*Spu, []*Sku, error)
	GetSpuDetail(ctx context.Context, spuID uint64) (*Spu, []*Sku, error)
}

package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ecommerce/internal/user/model"
	"ecommerce/pkg/hash"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// --- 接口定义 ---

type UserRepo interface {
	CreateUser(ctx context.Context, u *model.User) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByUserID(ctx context.Context, userID uint64) (*model.User, error)
	UpdateUser(ctx context.Context, u *model.User) (*model.User, error)
	VerifyPassword(ctx context.Context, username, password string) (*model.User, error)
}

type AddressRepo interface {
	CreateAddress(ctx context.Context, addr *model.Address) (*model.Address, error)
	UpdateAddress(ctx context.Context, addr *model.Address) (*model.Address, error)
	DeleteAddress(ctx context.Context, userID, addrID uint64) error
	GetAddress(ctx context.Context, userID, addrID uint64) (*model.Address, error)
	ListAddresses(ctx context.Context, userID uint64) ([]*model.Address, error)
	SetDefaultAddress(ctx context.Context, userID, addrID uint64) error
}

// --- 数据库模型 ---

type DBUser struct {
	gorm.Model
	Username string    `gorm:"uniqueIndex;not null;type:varchar(64);comment:用户名"`
	Password string    `gorm:"not null;type:varchar(255);comment:密码"`
	Nickname string    `gorm:"type:varchar(64);comment:昵称"`
	Avatar   string    `gorm:"type:varchar(255);comment:头像URL"`
	Gender   int32     `gorm:"type:tinyint;comment:性别 0:未知 1:男 2:女"`
	Birthday time.Time `gorm:"comment:生日"`
	Phone    string    `gorm:"uniqueIndex;type:varchar(20);comment:手机号"`
	Email    string    `gorm:"uniqueIndex;type:varchar(100);comment:邮箱"`
}

func (DBUser) TableName() string {
	return "users"
}

type DBAddress struct {
	gorm.Model
	UserID          uint64 `gorm:"index;comment:用户ID"`
	Name            string `gorm:"type:varchar(64);not null;comment:收货人姓名"`
	Phone           string `gorm:"type:varchar(20);not null;comment:手机号"`
	Province        string `gorm:"type:varchar(32);comment:省份"`
	City            string `gorm:"type:varchar(32);comment:城市"`
	District        string `gorm:"type:varchar(32);comment:区县"`
	DetailedAddress string `gorm:"type:varchar(255);comment:详细地址"`
	IsDefault       bool   `gorm:"type:tinyint(1);default:0;comment:是否为默认地址"`
}

func (DBAddress) TableName() string {
	return "addresses"
}

// --- 数据层核心 ---

type Data struct {
	db *gorm.DB
}

func NewData(db *gorm.DB) (*Data, func(), error) {
	d := &Data{
		db: db,
	}
	zap.S().Info("running database migrations...")
	if err := db.AutoMigrate(&DBUser{}, &DBAddress{}); err != nil {
		zap.S().Errorf("failed to migrate database: %v", err)
		return nil, nil, err
	}

	cleanup := func() {
		zap.S().Info("closing data layer...")
	}

	return d, cleanup, nil
}

// --- UserRepo 实现 ---

type userRepository struct {
	*Data
}

func NewUserRepo(data *Data) UserRepo {
	return &userRepository{data}
}

func (r *userRepository) CreateUser(ctx context.Context, u *model.User) (*model.User, error) {
	hashedPassword, err := hash.HashPassword(u.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	u.Password = hashedPassword

	dbUser := fromBizUser(u)
	if err := r.db.WithContext(ctx).Create(dbUser).Error; err != nil {
		zap.S().Errorf("failed to create user in db: %v", err)
		return nil, err
	}
	zap.S().Infof("created user in db: %d", dbUser.ID)
	return toBizUser(dbUser), nil
}

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var dbUser DBUser
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		zap.S().Errorf("failed to get user by username %s from db: %v", username, err)
		return nil, err
	}
	return toBizUser(&dbUser), nil
}

func (r *userRepository) GetUserByUserID(ctx context.Context, userID uint64) (*model.User, error) {
	var dbUser DBUser
	if err := r.db.WithContext(ctx).First(&dbUser, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		zap.S().Errorf("failed to get user by id %d from db: %v", userID, err)
		return nil, err
	}
	return toBizUser(&dbUser), nil
}

func (r *userRepository) UpdateUser(ctx context.Context, u *model.User) (*model.User, error) {
	dbUser := fromBizUser(u)
	if err := r.db.WithContext(ctx).Model(&DBUser{}).Where("id = ?", u.ID).Updates(dbUser).Error; err != nil {
		zap.S().Errorf("failed to update user %d in db: %v", u.ID, err)
		return nil, err
	}
	return r.GetUserByUserID(ctx, u.ID)
}

func (r *userRepository) VerifyPassword(ctx context.Context, username, password string) (*model.User, error) {
	user, err := r.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err // error already logged in GetUserByUsername
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if !hash.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid password")
	}
	return user, nil
}

// --- AddressRepo 实现 ---

type addressRepository struct {
	*Data
}

func NewAddressRepo(data *Data) AddressRepo {
	return &addressRepository{data}
}

func (r *addressRepository) CreateAddress(ctx context.Context, addr *model.Address) (*model.Address, error) {
	dbAddr := fromBizAddress(addr)
	if err := r.db.WithContext(ctx).Create(dbAddr).Error; err != nil {
		zap.S().Errorf("failed to create address in db: %v", err)
		return nil, err
	}
	zap.S().Infof("created address in db: %d", dbAddr.ID)
	return toBizAddress(dbAddr), nil
}

func (r *addressRepository) UpdateAddress(ctx context.Context, addr *model.Address) (*model.Address, error) {
	dbAddr := fromBizAddress(addr)
	if err := r.db.WithContext(ctx).Model(&DBAddress{}).Where("id = ?", addr.ID).Updates(dbAddr).Error; err != nil {
		zap.S().Errorf("failed to update address %d in db: %v", addr.ID, err)
		return nil, err
	}
	return r.GetAddress(ctx, addr.UserID, addr.ID)
}

func (r *addressRepository) DeleteAddress(ctx context.Context, userID, addrID uint64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&DBAddress{}, addrID).Error; err != nil {
		zap.S().Errorf("failed to delete address %d for user %d: %v", addrID, userID, err)
		return err
	}
	return nil
}

func (r *addressRepository) GetAddress(ctx context.Context, userID, addrID uint64) (*model.Address, error) {
	var dbAddr DBAddress
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&dbAddr, addrID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		zap.S().Errorf("failed to get address %d for user %d: %v", addrID, userID, err)
		return nil, err
	}
	return toBizAddress(&dbAddr), nil
}

func (r *addressRepository) ListAddresses(ctx context.Context, userID uint64) ([]*model.Address, error) {
	var dbAddrs []*DBAddress
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&dbAddrs).Error; err != nil {
		zap.S().Errorf("failed to list addresses for user %d: %v", userID, err)
		return nil, err
	}
	bizAddrs := make([]*model.Address, len(dbAddrs))
	for i, dbAddr := range dbAddrs {
		bizAddrs[i] = toBizAddress(dbAddr)
	}
	return bizAddrs, nil
}

func (r *addressRepository) SetDefaultAddress(ctx context.Context, userID, addrID uint64) error {
	// 使用事务确保操作的原子性
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 将该用户的所有地址都设为非默认
		if err := tx.Model(&DBAddress{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			zap.S().Errorf("failed to unset default addresses for user %d: %v", userID, err)
			return err
		}
		// 2. 将指定的地址设为默认
		if err := tx.Model(&DBAddress{}).Where("id = ? AND user_id = ?", addrID, userID).Update("is_default", true).Error; err != nil {
			zap.S().Errorf("failed to set default address %d for user %d: %v", addrID, userID, err)
			return err
		}
		return nil
	})
}

// --- 模型转换辅助函数 ---

func toBizUser(dbUser *DBUser) *model.User {
	if dbUser == nil {
		return nil
	}
	return &model.User{
		ID:        uint64(dbUser.ID),
		Username:  dbUser.Username,
		Password:  dbUser.Password,
		Nickname:  dbUser.Nickname,
		Avatar:    dbUser.Avatar,
		Gender:    dbUser.Gender,
		Birthday:  dbUser.Birthday,
		Phone:     dbUser.Phone,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
}

func fromBizUser(bizUser *model.User) *DBUser {
	if bizUser == nil {
		return nil
	}
	return &DBUser{
		Model:    gorm.Model{ID: uint(bizUser.ID), CreatedAt: bizUser.CreatedAt, UpdatedAt: bizUser.UpdatedAt},
		Username: bizUser.Username,
		Password: bizUser.Password,
		Nickname: bizUser.Nickname,
		Avatar:   bizUser.Avatar,
		Gender:   bizUser.Gender,
		Birthday: bizUser.Birthday,
		Phone:    bizUser.Phone,
		Email:    bizUser.Email,
	}
}

func toBizAddress(dbAddr *DBAddress) *model.Address {
	if dbAddr == nil {
		return nil
	}
	return &model.Address{
		ID:              uint64(dbAddr.ID),
		UserID:          dbAddr.UserID,
		Name:            &dbAddr.Name,
		Phone:           &dbAddr.Phone,
		Province:        &dbAddr.Province,
		City:            &dbAddr.City,
		District:        &dbAddr.District,
		DetailedAddress: &dbAddr.DetailedAddress,
		IsDefault:       &dbAddr.IsDefault,
	}
}

func fromBizAddress(bizAddr *model.Address) *DBAddress {
	if bizAddr == nil {
		return nil
	}
	addr := &DBAddress{
		Model:  gorm.Model{ID: uint(bizAddr.ID)},
		UserID: bizAddr.UserID,
	}
	if bizAddr.Name != nil {
		addr.Name = *bizAddr.Name
	}
	if bizAddr.Phone != nil {
		addr.Phone = *bizAddr.Phone
	}
	if bizAddr.Province != nil {
		addr.Province = *bizAddr.Province
	}
	if bizAddr.City != nil {
		addr.City = *bizAddr.City
	}
	if bizAddr.District != nil {
		addr.District = *bizAddr.District
	}
	if bizAddr.DetailedAddress != nil {
		addr.DetailedAddress = *bizAddr.DetailedAddress
	}
	if bizAddr.IsDefault != nil {
		addr.IsDefault = *bizAddr.IsDefault
	}
	return addr
}

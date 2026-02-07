package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/pkg/contextx"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// tx helpers
func (r *userRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *userRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *userRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *userRepository) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txCtx := contextx.WithTx(ctx, tx)
		return fn(txCtx)
	})
}

func (r *userRepository) Save(ctx context.Context, user *domain.User) error {
	model := toUserModel(user)
	if model == nil {
		return nil
	}
	db := r.getDB(ctx).WithContext(ctx)
	if err := db.Save(model).Error; err != nil {
		return err
	}
	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var model UserModel
	err := r.getDB(ctx).WithContext(ctx).Preload("Addresses").First(&model, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toUser(&model), nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var model UserModel
	err := r.getDB(ctx).WithContext(ctx).Where("username = ?", username).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toUser(&model), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel
	err := r.getDB(ctx).WithContext(ctx).Where("email = ?", email).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toUser(&model), nil
}

func (r *userRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var model UserModel
	err := r.getDB(ctx).WithContext(ctx).Where("phone = ?", phone).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toUser(&model), nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	model := toUserModel(user)
	if model == nil {
		return nil
	}
	db := r.getDB(ctx).WithContext(ctx)
	if err := db.Model(&UserModel{}).Where("id = ?", model.ID).Updates(map[string]any{
		"full_name":  model.FullName,
		"nickname":   model.Nickname,
		"avatar":     model.Avatar,
		"gender":     model.Gender,
		"birthday":   model.Birthday,
		"status":     model.Status,
		"phone":      model.Phone,
		"email":      model.Email,
		"username":   model.Username,
		"password":   model.Password,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.getDB(ctx).WithContext(ctx).Delete(&UserModel{}, id).Error
}

func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, int64, error) {
	var models []*UserModel
	var count int64
	db := r.getDB(ctx).WithContext(ctx)
	if err := db.Model(&UserModel{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Preload("Addresses").Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	users := make([]*domain.User, len(models))
	for i, m := range models {
		users[i] = toUser(m)
	}
	return users, count, nil
}

func (r *userRepository) getDB(ctx context.Context) *gorm.DB {
	if tx := contextx.GetTx(ctx); tx != nil {
		if gormTx, ok := tx.(*gorm.DB); ok {
			return gormTx
		}
	}
	return r.db
}

type addressRepository struct {
	db *gorm.DB
}

// NewAddressRepository 创建地址仓储实例
func NewAddressRepository(db *gorm.DB) domain.AddressRepository {
	return &addressRepository{db: db}
}

// tx helpers
func (r *addressRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *addressRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *addressRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *addressRepository) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txCtx := contextx.WithTx(ctx, tx)
		return fn(txCtx)
	})
}

func (r *addressRepository) Save(ctx context.Context, address *domain.Address) error {
	model := toAddressModel(address)
	if model == nil {
		return nil
	}
	db := r.getDB(ctx).WithContext(ctx)
	if err := db.Save(model).Error; err != nil {
		return err
	}
	address.ID = model.ID
	address.CreatedAt = model.CreatedAt
	address.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *addressRepository) FindByID(ctx context.Context, id uint) (*domain.Address, error) {
	var model AddressModel
	err := r.getDB(ctx).WithContext(ctx).First(&model, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toAddress(&model), nil
}

func (r *addressRepository) FindDefaultByUserID(ctx context.Context, userID uint) (*domain.Address, error) {
	var model AddressModel
	err := r.getDB(ctx).WithContext(ctx).Where("user_id = ? AND is_default = ?", userID, true).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toAddress(&model), nil
}

func (r *addressRepository) FindByUserID(ctx context.Context, userID uint) ([]*domain.Address, error) {
	var models []*AddressModel
	err := r.getDB(ctx).WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	addrs := make([]*domain.Address, len(models))
	for i, m := range models {
		addrs[i] = toAddress(m)
	}
	return addrs, nil
}

func (r *addressRepository) Update(ctx context.Context, address *domain.Address) error {
	model := toAddressModel(address)
	if model == nil {
		return nil
	}
	db := r.getDB(ctx).WithContext(ctx)
	return db.Model(&AddressModel{}).Where("id = ?", model.ID).Updates(map[string]any{
		"recipient_name":   model.RecipientName,
		"phone_number":     model.PhoneNumber,
		"province":         model.Province,
		"city":             model.City,
		"district":         model.District,
		"detailed_address": model.DetailedAddress,
		"postal_code":      model.PostalCode,
		"is_default":       model.IsDefault,
	}).Error
}

func (r *addressRepository) Delete(ctx context.Context, id uint) error {
	return r.getDB(ctx).WithContext(ctx).Delete(&AddressModel{}, id).Error
}

func (r *addressRepository) SetDefault(ctx context.Context, userID, addressID uint) error {
	db := r.getDB(ctx).WithContext(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&AddressModel{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&AddressModel{}).Where("id = ?", addressID).Update("is_default", true).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *addressRepository) getDB(ctx context.Context) *gorm.DB {
	if tx := contextx.GetTx(ctx); tx != nil {
		if gormTx, ok := tx.(*gorm.DB); ok {
			return gormTx
		}
	}
	return r.db
}

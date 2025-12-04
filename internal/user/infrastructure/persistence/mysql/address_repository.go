package mysql

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/user/domain" // 导入用户领域的领域接口和实体。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// AddressRepository 结构体是 domain.AddressRepository 接口的MySQL实现。
type AddressRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewAddressRepository 创建并返回一个新的 AddressRepository 实例。
func NewAddressRepository(db *gorm.DB) *AddressRepository {
	return &AddressRepository{db: db}
}

// Save 将地址实体保存到数据库。
// 如果是新地址，则创建；如果地址ID已存在，则更新。
func (r *AddressRepository) Save(ctx context.Context, address *domain.Address) error {
	// 使用Create来插入新地址。
	// GORM的Create在没有主键值时插入。
	return r.db.WithContext(ctx).Create(address).Error
}

// FindByID 根据ID从数据库获取地址记录。
func (r *AddressRepository) FindByID(ctx context.Context, id uint) (*domain.Address, error) {
	var address domain.Address
	if err := r.db.WithContext(ctx).First(&address, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &address, nil
}

// FindDefaultByUserID 获取指定用户的默认地址记录。
// 如果记录未找到，则返回nil。
func (r *AddressRepository) FindDefaultByUserID(ctx context.Context, userID uint) (*domain.Address, error) {
	var address domain.Address
	// 查询 user_id 匹配且 is_default 为 true 的地址。
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_default = ?", userID, true).First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &address, nil
}

// FindByUserID 获取指定用户ID的所有地址记录。
func (r *AddressRepository) FindByUserID(ctx context.Context, userID uint) ([]*domain.Address, error) {
	var addresses []*domain.Address
	// 查询 user_id 匹配的所有地址。
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
		return nil, err
	}
	return addresses, nil
}

// Update 更新地址实体。
func (r *AddressRepository) Update(ctx context.Context, address *domain.Address) error {
	// 使用Save来更新实体，GORM会根据主键判断是插入还是更新。
	return r.db.WithContext(ctx).Save(address).Error
}

// Delete 根据ID从数据库删除地址记录。
// GORM默认进行软删除。
func (r *AddressRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Address{}, id).Error
}

// SetDefault 设置指定用户ID的默认地址。
// 此操作在一个事务中执行，以确保原子性：
// 1. 将该用户所有地址的is_default字段设置为false。
// 2. 将指定地址的is_default字段设置为true。
func (r *AddressRepository) SetDefault(ctx context.Context, userID, addressID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 将该用户所有地址的 is_default 字段更新为 false。
		if err := tx.Model(&domain.Address{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}

		// 2. 将指定地址的 is_default 字段更新为 true。
		if err := tx.Model(&domain.Address{}).Where("id = ? AND user_id = ?", addressID, userID).Update("is_default", true).Error; err != nil {
			return err
		}

		return nil
	})
}

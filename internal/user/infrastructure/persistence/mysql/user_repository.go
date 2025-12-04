package mysql

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/user/domain" // 导入用户领域的领域接口和实体。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// UserRepository 结构体是 domain.UserRepository 接口的MySQL实现。
type UserRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewUserRepository 创建并返回一个新的 UserRepository 实例。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Save 将用户实体保存到数据库。
// 如果是新用户，则创建；如果用户ID已存在，则更新。
// 对于User聚合根，第一次Save通常是Create。
func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	// 使用Create来插入新用户。
	// GORM的Create在没有主键值时插入。
	// TODO: 如果User聚合根包含Addresses，这里可能需要事务来确保User和Addresses的原子性保存。
	// 目前，Addresses通过Preload加载，Save操作只针对User主实体。
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByID 根据ID从数据库获取用户记录，并预加载其关联的地址。
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	// Preload "Addresses" 确保在获取用户时，同时加载所有关联的地址。
	if err := r.db.WithContext(ctx).Preload("Addresses").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &user, nil
}

// FindByUsername 根据用户名从数据库获取用户记录，并预加载其关联的地址。
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	// Preload "Addresses" 确保在获取用户时，同时加载所有关联的地址。
	if err := r.db.WithContext(ctx).Preload("Addresses").Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &user, nil
}

// FindByEmail 根据邮箱从数据库获取用户记录，并预加载其关联的地址。
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	// Preload "Addresses" 确保在获取用户时，同时加载所有关联的地址。
	if err := r.db.WithContext(ctx).Preload("Addresses").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &user, nil
}

// FindByPhone 根据手机号从数据库获取用户记录，并预加载其关联的地址。
func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var user domain.User
	// Preload "Addresses" 确保在获取用户时，同时加载所有关联的地址。
	if err := r.db.WithContext(ctx).Preload("Addresses").Where("phone = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &user, nil
}

// Update 更新用户实体。
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	// 使用Save来更新实体，GORM会根据主键判断是插入还是更新。
	// 这里更新User主实体，如果Addresses有更改，需要单独处理或通过其他逻辑。
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 根据ID从数据库删除用户记录。
// GORM默认进行软删除。
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	// 软删除用户，通常不会级联删除关联的Addresses，除非配置了级联删除。
	return r.db.WithContext(ctx).Delete(&domain.User{}, id).Error
}

// List 从数据库列出所有用户记录，支持分页，并预加载其关联的地址。
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&total).Error; err != nil { // 统计总记录数。
		return nil, 0, err
	}

	// Preload "Addresses" 确保在获取用户列表时，同时加载所有关联的地址。
	if err := r.db.WithContext(ctx).Preload("Addresses").Offset(offset).Limit(limit).Find(&users).Error; err != nil { // 应用分页和查找。
		return nil, 0, err
	}

	return users, total, nil
}

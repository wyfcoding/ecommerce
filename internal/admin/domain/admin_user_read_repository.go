// 生成摘要：定义管理员读模型仓储接口（Redis）。
package domain

import "context"

// AdminUserReadRepository 读取管理员缓存接口。
type AdminUserReadRepository interface {
	Save(ctx context.Context, user *AdminUser) error
	GetByID(ctx context.Context, id uint) (*AdminUser, error)
	Delete(ctx context.Context, id uint) error
}

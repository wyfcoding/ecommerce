// 生成摘要：定义广告位读模型仓储接口（Redis）。
package domain

import "context"

// BannerReadRepository 读取广告位详情的缓存接口。
type BannerReadRepository interface {
	Save(ctx context.Context, banner *Banner) error
	GetByID(ctx context.Context, id uint64) (*Banner, error)
	Delete(ctx context.Context, id uint64) error
}

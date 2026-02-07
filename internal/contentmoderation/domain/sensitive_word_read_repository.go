package domain

import "context"

// SensitiveWordReadRepository 定义敏感词读模型仓储接口（Redis）。
type SensitiveWordReadRepository interface {
	Save(ctx context.Context, word *SensitiveWord) error
	GetByID(ctx context.Context, id uint64) (*SensitiveWord, error)
	Delete(ctx context.Context, id uint64) error
}

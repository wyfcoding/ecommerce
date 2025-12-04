package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain/entity" // 导入内容审核领域的实体定义。
)

// ModerationRepository 是内容审核模块的仓储接口。
// 它定义了对内容审核记录和敏感词实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type ModerationRepository interface {
	// --- ModerationRecord methods ---

	// CreateRecord 在数据存储中创建一个新的内容审核记录。
	// ctx: 上下文。
	// record: 待创建的审核记录实体。
	CreateRecord(ctx context.Context, record *entity.ModerationRecord) error
	// GetRecord 根据ID获取内容审核记录实体。
	GetRecord(ctx context.Context, id uint64) (*entity.ModerationRecord, error)
	// UpdateRecord 更新内容审核记录实体的信息。
	UpdateRecord(ctx context.Context, record *entity.ModerationRecord) error
	// ListRecords 列出所有内容审核记录，支持通过状态过滤和分页。
	ListRecords(ctx context.Context, status entity.ModerationStatus, offset, limit int) ([]*entity.ModerationRecord, int64, error)

	// --- SensitiveWord methods ---

	// CreateWord 在数据存储中创建一个新的敏感词实体。
	CreateWord(ctx context.Context, word *entity.SensitiveWord) error
	// GetWord 根据ID获取敏感词实体。
	GetWord(ctx context.Context, id uint64) (*entity.SensitiveWord, error)
	// UpdateWord 更新敏感词实体的信息。
	UpdateWord(ctx context.Context, word *entity.SensitiveWord) error
	// DeleteWord 根据ID删除敏感词实体。
	DeleteWord(ctx context.Context, id uint64) error
	// ListWords 列出所有敏感词实体，支持分页。
	ListWords(ctx context.Context, offset, limit int) ([]*entity.SensitiveWord, int64, error)
	// FindWord 根据敏感词字符串查找敏感词实体。
	FindWord(ctx context.Context, word string) (*entity.SensitiveWord, error)
}

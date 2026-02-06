// 生成摘要：定义结算读模型仓储接口（Redis），用于高频查询。
// 假设：结算单号与结算ID为全局唯一，读模型为最终一致性。
package domain

import "context"

// SettlementReadRepository 定义结算读模型的高性能访问接口。
type SettlementReadRepository interface {
	// Save 保存或更新结算读模型。
	Save(ctx context.Context, settlement *Settlement) error
	// GetByID 根据结算ID获取读模型。
	GetByID(ctx context.Context, settlementID uint64) (*Settlement, error)
	// GetByNo 根据结算单号获取读模型。
	GetByNo(ctx context.Context, settlementNo string) (*Settlement, error)
	// Delete 删除读模型数据（用于清理）。
	Delete(ctx context.Context, settlementID uint64, settlementNo string) error
}

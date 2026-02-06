// 生成摘要：定义结算搜索仓储接口（Elasticsearch），用于分页与过滤查询。
// 假设：ES 索引字段与 domain.Settlement 的 JSON 映射一致，CreatedAt 可用于排序。
package domain

import (
	"context"
	"time"
)

// SettlementSearchRepository 定义结算搜索仓储接口。
type SettlementSearchRepository interface {
	// Index 将结算单写入搜索索引。
	Index(ctx context.Context, settlement *Settlement) error
	// Delete 从索引中删除结算单。
	Delete(ctx context.Context, settlementID uint64) error
	// Search 按条件检索结算单（支持商户与状态过滤、分页）。
	Search(ctx context.Context, merchantID *uint64, status *SettlementStatus, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*Settlement, int64, error)
	// FindByNo 通过结算单号检索结算单。
	FindByNo(ctx context.Context, settlementNo string) (*Settlement, error)
}

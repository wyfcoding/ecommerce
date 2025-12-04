package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/entity" // 导入售后领域的实体定义。
)

// AfterSalesRepository 是售后模块的仓储接口。
// 它定义了对售后申请实体及其关联日志进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type AfterSalesRepository interface {
	// --- AfterSales methods ---

	// Create 在数据存储中创建一个新的售后申请实体。
	// ctx: 上下文。
	// afterSales: 待创建的售后申请实体。
	Create(ctx context.Context, afterSales *entity.AfterSales) error
	// GetByID 根据ID获取售后申请实体，并预加载其关联的商品和日志。
	GetByID(ctx context.Context, id uint64) (*entity.AfterSales, error)
	// GetByNo 根据售后单号获取售后申请实体，并预加载其关联的商品和日志。
	GetByNo(ctx context.Context, no string) (*entity.AfterSales, error)
	// Update 更新售后申请实体的信息。
	Update(ctx context.Context, afterSales *entity.AfterSales) error
	// List 列出所有售后申请实体，支持通过查询条件进行过滤和分页。
	List(ctx context.Context, query *AfterSalesQuery) ([]*entity.AfterSales, int64, error)

	// --- Log methods ---

	// CreateLog 在数据存储中创建一个新的售后操作日志记录。
	CreateLog(ctx context.Context, log *entity.AfterSalesLog) error
	// ListLogs 列出指定售后申请的所有操作日志。
	ListLogs(ctx context.Context, afterSalesID uint64) ([]*entity.AfterSalesLog, error)
}

// AfterSalesQuery 结构体定义了查询售后申请的条件。
// 它用于在仓储层进行数据过滤和分页。
type AfterSalesQuery struct {
	UserID       uint64                  // 根据用户ID过滤。
	OrderID      uint64                  // 根据订单ID过滤。
	Type         entity.AfterSalesType   // 根据售后类型过滤。
	Status       entity.AfterSalesStatus // 根据售后状态过滤。
	AfterSalesNo string                  // 根据售后单号过滤。
	Page         int                     // 页码，用于分页。
	PageSize     int                     // 每页数量，用于分页。
}

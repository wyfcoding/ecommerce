package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/ai_model/domain/entity" // 导入AI模型领域的实体定义。
	"time"                                                           // 导入时间包，用于查询条件。
)

// AIModelRepository 是AI模型模块的仓储接口。
// 它定义了对AI模型、训练日志和预测记录进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type AIModelRepository interface {
	// --- AIModel methods ---

	// Create 在数据存储中创建一个新的AI模型实体。
	// ctx: 上下文。
	// model: 待创建的AI模型实体。
	Create(ctx context.Context, model *entity.AIModel) error
	// GetByID 根据ID获取AI模型实体。
	GetByID(ctx context.Context, id uint64) (*entity.AIModel, error)
	// GetByNo 根据模型编号获取AI模型实体。
	GetByNo(ctx context.Context, no string) (*entity.AIModel, error)
	// Update 更新AI模型实体的信息。
	Update(ctx context.Context, model *entity.AIModel) error
	// List 列出所有AI模型实体，支持通过查询条件进行过滤和分页。
	List(ctx context.Context, query *ModelQuery) ([]*entity.AIModel, int64, error)
	// Delete 根据ID删除AI模型实体。
	Delete(ctx context.Context, id uint64) error

	// --- Training Log methods ---

	// CreateTrainingLog 在数据存储中创建一个新的模型训练日志记录。
	CreateTrainingLog(ctx context.Context, log *entity.ModelTrainingLog) error
	// ListTrainingLogs 列出指定模型的所有训练日志。
	ListTrainingLogs(ctx context.Context, modelID uint64) ([]*entity.ModelTrainingLog, error)

	// --- Prediction methods ---

	// CreatePrediction 在数据存储中创建一个新的模型预测记录。
	CreatePrediction(ctx context.Context, prediction *entity.ModelPrediction) error
	// ListPredictions 列出指定模型的所有预测记录，支持时间范围过滤和分页。
	ListPredictions(ctx context.Context, modelID uint64, startTime, endTime time.Time, page, pageSize int) ([]*entity.ModelPrediction, int64, error)
}

// ModelQuery 结构体定义了查询AI模型列表的条件。
// 它用于在仓储层进行数据过滤和分页。
type ModelQuery struct {
	Status    entity.ModelStatus // 根据模型状态过滤。
	Type      string             // 根据模型类型过滤。
	Algorithm string             // 根据使用的算法过滤。
	CreatorID uint64             // 根据创建人ID过滤。
	Page      int                // 页码，用于分页。
	PageSize  int                // 每页数量，用于分页。
}

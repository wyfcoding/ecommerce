package domain

import (
	"context"
	"time"
)

// AIModelRepository 是AI模型模块的仓储接口。
// 它定义了对AI模型、训练日志和预测记录进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type AIModelRepository interface {
	// 事务支持
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// --- AIModel methods ---
	SaveModel(ctx context.Context, model *AIModel) error
	SaveModelInTx(ctx context.Context, tx any, model *AIModel) error
	GetModel(ctx context.Context, id uint64) (*AIModel, error)
	GetModelByNo(ctx context.Context, no string) (*AIModel, error)
	ListModels(ctx context.Context, query *ModelQuery) ([]*AIModel, int64, error)
	DeleteModel(ctx context.Context, id uint64) error

	// --- Training Log methods ---
	SaveTrainingLog(ctx context.Context, log *ModelTrainingLog) error
	SaveTrainingLogInTx(ctx context.Context, tx any, log *ModelTrainingLog) error
	ListTrainingLogs(ctx context.Context, modelID uint64) ([]*ModelTrainingLog, error)

	// --- Prediction methods ---
	SavePrediction(ctx context.Context, prediction *ModelPrediction) error
	SavePredictionInTx(ctx context.Context, tx any, prediction *ModelPrediction) error
	ListPredictions(ctx context.Context, modelID uint64, startTime, endTime time.Time, page, pageSize int) ([]*ModelPrediction, int64, error)
}

// ModelQuery 结构体定义了查询AI模型列表的条件。
// 它用于在仓储层进行数据过滤和分页。
type ModelQuery struct {
	Status    ModelStatus // 根据模型状态过滤。
	Type      string      // 根据模型类型过滤。
	Algorithm string      // 根据使用的算法过滤。
	CreatorID uint64      // 根据创建人ID过滤。
	Page      int         // 页码，用于分页。
	PageSize  int         // 每页数量，用于分页。
}

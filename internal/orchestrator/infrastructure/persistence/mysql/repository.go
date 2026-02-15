// 生成摘要：实现编排器服务的 MySQL 仓储层，基于 GORM。
// 变更说明：从旧的 infrastructure 目录迁移至 persistence/mysql，并完善步骤更新逻辑。

package mysql

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/orchestrator/domain"
	"gorm.io/gorm"
)

type orchestratorRepository struct {
	db *gorm.DB
}

// NewRepository 创建编排器仓储
func NewRepository(db *gorm.DB) domain.OrchestratorRepository {
	return &orchestratorRepository{db: db}
}

func (r *orchestratorRepository) SaveInstance(ctx context.Context, instance *domain.SagaInstance) error {
	return r.db.WithContext(ctx).Save(instance).Error
}

func (r *orchestratorRepository) FindInstance(ctx context.Context, id string) (*domain.SagaInstance, error) {
	var instance domain.SagaInstance
	if err := r.db.WithContext(ctx).Preload("Steps").Where("saga_id = ?", id).First(&instance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *orchestratorRepository) FindInstances(ctx context.Context, filter map[string]any) ([]*domain.SagaInstance, error) {
	var instances []*domain.SagaInstance
	query := r.db.WithContext(ctx).Preload("Steps")
	for k, v := range filter {
		query = query.Where(k+" = ?", v)
	}
	if err := query.Find(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}

func (r *orchestratorRepository) RegisterDefinition(ctx context.Context, def *domain.SagaDefinition) error {
	// 目前定义暂存于内存或配置中，若需持久化可在此实现
	return nil
}

func (r *orchestratorRepository) GetDefinition(ctx context.Context, sagaType string) (*domain.SagaDefinition, error) {
	// 同上
	return nil, nil
}

// OrchestratorRepository 接口补充实现 (为了向前兼容旧代码调用，也可在 domain 定义中统一)
func (r *orchestratorRepository) FindInstanceByID(ctx context.Context, sagaID string) (*domain.SagaInstance, error) {
	return r.FindInstance(ctx, sagaID)
}

func (r *orchestratorRepository) UpdateStep(ctx context.Context, sagaID string, step *domain.SagaStep) error {
	return r.db.WithContext(ctx).Save(step).Error
}

// 显式实现旧接口以防编译失败 (OrchestratorRepository vs SagaRepository)
var _ domain.OrchestratorRepository = (*orchestratorRepository)(nil)
var _ domain.SagaRepository = (*orchestratorRepository)(nil)

// 生成摘要：实现AI模型读模型 Redis 仓储，提供按模型ID/编号的快速读取。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/aimodel/domain"
)

const (
	modelDetailPrefix = "aimodel:model:detail:"
	modelNoPrefix     = "aimodel:model:no:"
)

type aiModelReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewAIModelReadRepository 创建AI模型读模型仓储。
func NewAIModelReadRepository(client redis.UniversalClient, ttl time.Duration) domain.AIModelReadRepository {
	return &aiModelReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *aiModelReadRepository) Save(ctx context.Context, model *domain.AIModel) error {
	if model == nil {
		return nil
	}

	data, err := json.Marshal(model)
	if err != nil {
		return err
	}

	modelIDKey := r.modelIDKey(uint64(model.ID))
	modelNoKey := r.modelNoKey(model.ModelNo)

	pipe := r.client.Pipeline()
	pipe.Set(ctx, modelIDKey, data, r.ttl)
	if model.ModelNo != "" {
		pipe.Set(ctx, modelNoKey, fmt.Sprintf("%d", model.ID), r.ttl)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (r *aiModelReadRepository) GetByID(ctx context.Context, id uint64) (*domain.AIModel, error) {
	data, err := r.client.Get(ctx, r.modelIDKey(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var model domain.AIModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *aiModelReadRepository) GetByNo(ctx context.Context, modelNo string) (*domain.AIModel, error) {
	if modelNo == "" {
		return nil, nil
	}
	idStr, err := r.client.Get(ctx, r.modelNoKey(modelNo)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var modelID uint64
	if _, err := fmt.Sscanf(idStr, "%d", &modelID); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, modelID)
}

func (r *aiModelReadRepository) Delete(ctx context.Context, id uint64, modelNo string) error {
	keys := []string{r.modelIDKey(id)}
	if modelNo != "" {
		keys = append(keys, r.modelNoKey(modelNo))
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *aiModelReadRepository) modelIDKey(id uint64) string {
	return fmt.Sprintf("%s%d", modelDetailPrefix, id)
}

func (r *aiModelReadRepository) modelNoKey(modelNo string) string {
	return fmt.Sprintf("%s%s", modelNoPrefix, modelNo)
}

// 生成摘要：实现风险分析结果读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/risksecurity/domain"
)

const riskAnalysisLatestPrefix = "risk:analysis:latest:"

type riskAnalysisReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewRiskAnalysisReadRepository 创建风险分析读模型仓储。
func NewRiskAnalysisReadRepository(client redis.UniversalClient, ttl time.Duration) domain.RiskAnalysisReadRepository {
	return &riskAnalysisReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *riskAnalysisReadRepository) SaveLatest(ctx context.Context, userID uint64, result *domain.RiskAnalysisResult) error {
	if result == nil {
		return nil
	}
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(userID), data, r.ttl).Err()
}

func (r *riskAnalysisReadRepository) GetLatestByUser(ctx context.Context, userID uint64) (*domain.RiskAnalysisResult, error) {
	data, err := r.client.Get(ctx, r.key(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result domain.RiskAnalysisResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *riskAnalysisReadRepository) DeleteLatestByUser(ctx context.Context, userID uint64) error {
	return r.client.Del(ctx, r.key(userID)).Err()
}

func (r *riskAnalysisReadRepository) key(userID uint64) string {
	return fmt.Sprintf("%s%d", riskAnalysisLatestPrefix, userID)
}

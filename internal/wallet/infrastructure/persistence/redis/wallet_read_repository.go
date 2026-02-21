// Package redis 钱包读模型 Redis 仓储实现
// 生成摘要：
// 1) 实现 WalletReadRepository 接口，使用 Redis 作为读模型存储
// 2) 采用 Hash 结构存储钱包信息，支持 TTL 自动过期和 LRU 淘汰
// 3) 通过事件投影保持与 MySQL 写模型的最终一致性
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/wallet/domain"
)

// WalletReadRepositoryImpl Redis 钱包读模型仓储实现
type WalletReadRepositoryImpl struct {
	redisClient redis.UniversalClient
	logger      *slog.Logger
}

// NewWalletReadRepository 创建 Redis 钱包读模型仓储实例
func NewWalletReadRepository(redisClient redis.UniversalClient, logger *slog.Logger) domain.WalletReadRepository {
	return &WalletReadRepositoryImpl{
		redisClient: redisClient,
		logger:      logger.With("module", "wallet_read_repository"),
	}
}

// walletKey 生成钱包读模型的 Redis Key
func (r *WalletReadRepositoryImpl) walletKey(walletID uint64) string {
	return fmt.Sprintf("wallet:read:%d", walletID)
}

// userWalletsKey 生成用户所有钱包列表的 Redis Key
func (r *WalletReadRepositoryImpl) userWalletsKey(userID uint64) string {
	return fmt.Sprintf("user:wallets:%d", userID)
}

// GetByWalletID 根据钱包ID获取读模型
func (r *WalletReadRepositoryImpl) GetByWalletID(ctx context.Context, walletID uint64) (*domain.WalletReadModel, error) {
	start := time.Now()
	key := r.walletKey(walletID)

	data, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		r.logger.ErrorContext(ctx, "failed to get wallet from redis",
			"wallet_id", walletID, "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var model domain.WalletReadModel
	if err := json.Unmarshal([]byte(data), &model); err != nil {
		r.logger.ErrorContext(ctx, "failed to unmarshal wallet data",
			"wallet_id", walletID, "error", err)
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	r.logger.DebugContext(ctx, "wallet read model retrieved from redis",
		"wallet_id", walletID, "duration", time.Since(start))
	return &model, nil
}

// GetByUserID 根据用户ID和币种获取读模型
func (r *WalletReadRepositoryImpl) GetByUserID(ctx context.Context, userID uint64, currency string) (*domain.WalletReadModel, error) {
	start := time.Now()

	// 先获取用户所有钱包ID列表
	walletsKey := r.userWalletsKey(userID)
	walletIDs, err := r.redisClient.SMembers(ctx, walletsKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		r.logger.ErrorContext(ctx, "failed to get user wallets from redis",
			"user_id", userID, "error", err)
		return nil, fmt.Errorf("get user wallets: %w", err)
	}

	// 遍历所有钱包，找到匹配币种的
	for _, walletIDStr := range walletIDs {
		var walletID uint64
		_, err := fmt.Sscanf(walletIDStr, "%d", &walletID)
		if err != nil {
			continue
		}

		model, err := r.GetByWalletID(ctx, walletID)
		if err != nil || model == nil {
			continue
		}

		if model.Currency == currency {
			r.logger.DebugContext(ctx, "wallet found by user and currency",
				"user_id", userID, "currency", currency, "duration", time.Since(start))
			return model, nil
		}
	}

	return nil, nil
}

// GetAllByUserID 获取用户所有钱包的读模型
func (r *WalletReadRepositoryImpl) GetAllByUserID(ctx context.Context, userID uint64) ([]*domain.WalletReadModel, error) {
	start := time.Now()

	walletsKey := r.userWalletsKey(userID)
	walletIDs, err := r.redisClient.SMembers(ctx, walletsKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []*domain.WalletReadModel{}, nil
		}
		r.logger.ErrorContext(ctx, "failed to get user wallets from redis",
			"user_id", userID, "error", err)
		return nil, fmt.Errorf("get user wallets: %w", err)
	}

	var models []*domain.WalletReadModel
	for _, walletIDStr := range walletIDs {
		var walletID uint64
		_, err := fmt.Sscanf(walletIDStr, "%d", &walletID)
		if err != nil {
			continue
		}

		model, err := r.GetByWalletID(ctx, walletID)
		if err != nil || model == nil {
			continue
		}

		models = append(models, model)
	}

	r.logger.DebugContext(ctx, "user wallets retrieved from redis",
		"user_id", userID, "count", len(models), "duration", time.Since(start))
	return models, nil
}

// Save 保存或更新读模型
func (r *WalletReadRepositoryImpl) Save(ctx context.Context, model *domain.WalletReadModel) error {
	start := time.Now()
	key := r.walletKey(model.WalletID)

	data, err := json.Marshal(model)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to marshal wallet model",
			"wallet_id", model.WalletID, "error", err)
		return fmt.Errorf("marshal: %w", err)
	}

	// 设置 TTL 为 24 小时
	err = r.redisClient.Set(ctx, key, string(data), 24*time.Hour).Err()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to save wallet to redis",
			"wallet_id", model.WalletID, "error", err)
		return fmt.Errorf("redis set: %w", err)
	}

	// 更新用户钱包列表
	walletsKey := r.userWalletsKey(model.UserID)
	walletIDStr := fmt.Sprintf("%d", model.WalletID)
	if err := r.redisClient.SAdd(ctx, walletsKey, walletIDStr).Err(); err != nil {
		r.logger.WarnContext(ctx, "failed to add wallet to user set",
			"user_id", model.UserID, "wallet_id", model.WalletID, "error", err)
	}

	r.logger.DebugContext(ctx, "wallet read model saved to redis",
		"wallet_id", model.WalletID, "duration", time.Since(start))
	return nil
}

// Delete 删除读模型
func (r *WalletReadRepositoryImpl) Delete(ctx context.Context, walletID uint64) error {
	start := time.Now()
	key := r.walletKey(walletID)

	// 先获取钱包信息，以便从用户集合中删除
	model, err := r.GetByWalletID(ctx, walletID)
	if err != nil {
		return err
	}
	if model == nil {
		return nil // 不存在则直接返回
	}

	// 删除钱包数据
	if err := r.redisClient.Del(ctx, key).Err(); err != nil {
		r.logger.ErrorContext(ctx, "failed to delete wallet from redis",
			"wallet_id", walletID, "error", err)
		return fmt.Errorf("redis del: %w", err)
	}

	// 从用户钱包列表中删除
	walletsKey := r.userWalletsKey(model.UserID)
	walletIDStr := fmt.Sprintf("%d", walletID)
	if err := r.redisClient.SRem(ctx, walletsKey, walletIDStr).Err(); err != nil {
		r.logger.WarnContext(ctx, "failed to remove wallet from user set",
			"user_id", model.UserID, "wallet_id", walletID, "error", err)
	}

	r.logger.DebugContext(ctx, "wallet read model deleted from redis",
		"wallet_id", walletID, "duration", time.Since(start))
	return nil
}

// TransactionReadRepositoryImpl Redis 交易记录读模型仓储实现
type TransactionReadRepositoryImpl struct {
	redisClient redis.UniversalClient
	logger      *slog.Logger
}

// NewTransactionReadRepository 创建 Redis 交易记录读模型仓储实例
func NewTransactionReadRepository(redisClient redis.UniversalClient, logger *slog.Logger) domain.TransactionReadRepository {
	return &TransactionReadRepositoryImpl{
		redisClient: redisClient,
		logger:      logger.With("module", "transaction_read_repository"),
	}
}

// recentTransactionsKey 生成最近交易记录的 Redis Key
func (r *TransactionReadRepositoryImpl) recentTransactionsKey(walletID uint64) string {
	return fmt.Sprintf("wallet:transactions:recent:%d", walletID)
}

// GetRecent 获取最近的交易记录
func (r *TransactionReadRepositoryImpl) GetRecent(ctx context.Context, walletID uint64, limit int) ([]*domain.TransactionReadModel, error) {
	start := time.Now()
	key := r.recentTransactionsKey(walletID)

	// 使用 Redis List 存储最近交易记录（按时间倒序）
	data, err := r.redisClient.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []*domain.TransactionReadModel{}, nil
		}
		r.logger.ErrorContext(ctx, "failed to get recent transactions from redis",
			"wallet_id", walletID, "error", err)
		return nil, fmt.Errorf("redis lrange: %w", err)
	}

	var transactions []*domain.TransactionReadModel
	for _, item := range data {
		var tx domain.TransactionReadModel
		if err := json.Unmarshal([]byte(item), &tx); err != nil {
			r.logger.WarnContext(ctx, "failed to unmarshal transaction data",
				"wallet_id", walletID, "error", err)
			continue
		}
		transactions = append(transactions, &tx)
	}

	r.logger.DebugContext(ctx, "recent transactions retrieved from redis",
		"wallet_id", walletID, "count", len(transactions), "duration", time.Since(start))
	return transactions, nil
}

// Save 保存交易记录到读模型
func (r *TransactionReadRepositoryImpl) Save(ctx context.Context, model *domain.TransactionReadModel) error {
	start := time.Now()
	key := r.recentTransactionsKey(model.WalletID)

	data, err := json.Marshal(model)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to marshal transaction model",
			"tx_id", model.ID, "error", err)
		return fmt.Errorf("marshal: %w", err)
	}

	// 使用 LPUSH 将新交易添加到列表头部
	if err := r.redisClient.LPush(ctx, key, string(data)).Err(); err != nil {
		r.logger.ErrorContext(ctx, "failed to save transaction to redis",
			"tx_id", model.ID, "error", err)
		return fmt.Errorf("redis lpush: %w", err)
	}

	// 限制列表长度（保留最近 100 条记录）
	if err := r.redisClient.LTrim(ctx, key, 0, 99).Err(); err != nil {
		r.logger.WarnContext(ctx, "failed to trim transaction list",
			"wallet_id", model.WalletID, "error", err)
	}

	// 设置 TTL 为 7 天
	if err := r.redisClient.Expire(ctx, key, 7*24*time.Hour).Err(); err != nil {
		r.logger.WarnContext(ctx, "failed to set ttl for transaction list",
			"wallet_id", model.WalletID, "error", err)
	}

	r.logger.DebugContext(ctx, "transaction saved to redis",
		"tx_id", model.ID, "duration", time.Since(start))
	return nil
}

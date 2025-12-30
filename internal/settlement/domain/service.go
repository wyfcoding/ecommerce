package domain

import (
	"context"
	"fmt"
	"time"
)

// LedgerService 账务核心服务
type LedgerService struct {
	repo LedgerRepository
}

// NewLedgerService 定义了 NewLedger 相关的服务逻辑。
func NewLedgerService(repo LedgerRepository) *LedgerService {
	return &LedgerService{repo: repo}
}

// PostEntry 记账
// 原子操作：校验平衡 -> 扣减/增加余额 -> 保存凭证
func (s *LedgerService) PostEntry(ctx context.Context, entry *JournalEntry) error {
	// 1. 校验借贷平衡
	if err := entry.Validate(); err != nil {
		return err
	}

	// 2. 准备更新账户余额
	// 注意：实际生产中需要 DB 事务支持。这里假设 repo 层处理了事务。
	// 这里我们模拟这一过程，repo 在 SaveJournalEntry 时应该包含更新余额逻辑

	// 为简化，我们假设 DB 事务在 Infrastructure 层处理
	// 或在此处开启事务 (需要 TransactionManager 抽象)

	// 步骤: 查找并锁定相关账户，更新余额
	// Debit (借) Asset/Expense -> 增加
	// Debit (借) Liability/Equity/Income -> 减少
	// Credit (贷) 反之

	// 这里仅做模型校验和保存调用
	if entry.EntryNo == "" {
		entry.EntryNo = fmt.Sprintf("JE%d", time.Now().UnixNano())
	}
	if entry.PostingDate.IsZero() {
		entry.PostingDate = time.Now()
	}

	return s.repo.CreateJournalEntry(entry)
}

// CreateAccount 开户
func (s *LedgerService) CreateAccount(ctx context.Context, subjectCode, entityID string) (*Account, error) {
	// check existing
	var err error
	acc, err := s.repo.GetAccount(subjectCode, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing account: %w", err)
	}
	if acc != nil {
		return acc, nil
	}

	acc = &Account{
		SubjectCode: subjectCode,
		EntityID:    entityID,
		Balance:     0,
		Currency:    "CNY",
	}
	if err = s.repo.SaveAccount(acc); err != nil {
		return nil, err
	}
	return acc, nil
}

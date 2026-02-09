// Package application 提供了 KYC 跨系统桥接服务。
// 变更说明：实现电商 KYC 流程与交易系统合规 (Compliance) 服务的对接，统一风险评估。
package application

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/identity/domain"
	"github.com/wyfcoding/pkg/logging"
)

// KYCBridgeService KYC 跨项目桥接服务
type KYCBridgeService struct {
	repo             domain.KYCRepository
	complianceClient ComplianceClient // 抽象的交易系统合规客户端
}

// ComplianceClient 交易系统合规服务客户端接口
type ComplianceClient interface {
	SyncKYCStatus(ctx context.Context, userID string, level int, status string) error
	CheckAML(ctx context.Context, userID string, idNumber string) (bool, string, error)
}

func NewKYCBridgeService(repo domain.KYCRepository, client ComplianceClient) *KYCBridgeService {
	return &KYCBridgeService{
		repo:             repo,
		complianceClient: client,
	}
}

// CompleteKYCAndSync 完成电商 KYC 并同步到交易系统
func (s *KYCBridgeService) CompleteKYCAndSync(ctx context.Context, userID uint64, reviewer string) error {
	logger := logging.Default()

	// 1. 获取电商 KYC 记录
	record, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// 2. 本地审批通过
	if err := record.Approve(reviewer, "Approved by Ecommerce KYC Bridge"); err != nil {
		return err
	}
	if err := s.repo.Save(ctx, record); err != nil {
		return err
	}

	// 3. 同步到交易系统 (Synergy)
	err = s.complianceClient.SyncKYCStatus(ctx, fmt.Sprintf("%d", userID), int(record.Level), "APPROVED")
	if err != nil {
		logger.ErrorContext(ctx, "failed to sync kyc status to trading system", "user_id", userID, "error", err)
		// 实际上这里可能需要重试机制或补偿逻辑
		return fmt.Errorf("kyc approved but sync failed: %w", err)
	}

	logger.InfoContext(ctx, "kyc status synced to trading system", "user_id", userID)
	return nil
}

// PerformGlobalAML 执行全局反洗钱检查
func (s *KYCBridgeService) PerformGlobalAML(ctx context.Context, userID uint64, idNumber string) (bool, error) {
	// 调用交易系统的专业 AML 模块
	passed, reason, err := s.complianceClient.CheckAML(ctx, fmt.Sprintf("%d", userID), idNumber)
	if err != nil {
		return false, err
	}

	if !passed {
		logging.Default().WarnContext(ctx, "global aml check failed", "user_id", userID, "reason", reason)
	}

	return passed, nil
}

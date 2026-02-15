// Package application 风控联动应用服务
package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/riskfederation/domain"
)

// RiskFederationService 风控联动服务
type RiskFederationService struct {
	repo              domain.RiskRepository
	ecomRiskEngine    domain.EcommerceRiskEngine
	finRiskEngine     domain.FinancialRiskEngine
	blacklistSvc      domain.BlacklistService
	logger            *slog.Logger
}

func NewRiskFederationService(
	repo domain.RiskRepository,
	ecomRiskEngine domain.EcommerceRiskEngine,
	finRiskEngine domain.FinancialRiskEngine,
	blacklistSvc domain.BlacklistService,
	logger *slog.Logger,
) *RiskFederationService {
	return &RiskFederationService{
		repo:           repo,
		ecomRiskEngine: ecomRiskEngine,
		finRiskEngine:  finRiskEngine,
		blacklistSvc:   blacklistSvc,
		logger:         logger.With("module", "risk_federation_service"),
	}
}

// AssessEcommerceRisk 评估电商交易风险
func (s *RiskFederationService) AssessEcommerceRisk(ctx context.Context, req domain.EcommerceRiskRequest) (*domain.RiskAssessment, error) {
	// 1. 黑名单检查
	if isBlacklisted, _ := s.blacklistSvc.CheckUser(ctx, req.UserID); isBlacklisted {
		return s.createHighRiskAssessment(req.TransactionID, req.UserID, "user_in_blacklist"), nil
	}
	
	if isBlacklisted, _ := s.blacklistSvc.CheckIP(ctx, req.IP); isBlacklisted {
		return s.createHighRiskAssessment(req.TransactionID, req.UserID, "ip_in_blacklist"), nil
	}
	
	if isBlacklisted, _ := s.blacklistSvc.CheckDevice(ctx, req.DeviceID); isBlacklisted {
		return s.createHighRiskAssessment(req.TransactionID, req.UserID, "device_in_blacklist"), nil
	}

	// 2. 调用电商风控引擎
	ecomAssessment, err := s.ecomRiskEngine.EvaluateTransaction(ctx, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "ecommerce risk engine failed", 
			"transaction_id", req.TransactionID,
			"error", err)
		return nil, err
	}

	// 3. 保存评估结果
	if err := s.repo.SaveAssessment(ctx, ecomAssessment); err != nil {
		s.logger.WarnContext(ctx, "failed to save risk assessment", 
			"transaction_id", req.TransactionID,
			"error", err)
	}

	return ecomAssessment, nil
}

// AssessFinancialRisk 评估金融交易风险
func (s *RiskFederationService) AssessFinancialRisk(ctx context.Context, req domain.FinancialRiskRequest) (*domain.RiskAssessment, error) {
	// 调用金融风控引擎
	finAssessment, err := s.finRiskEngine.EvaluateTrade(ctx, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "financial risk engine failed", 
			"transaction_id", req.TransactionID,
			"error", err)
		return nil, err
	}

	// 保存评估结果
	if err := s.repo.SaveAssessment(ctx, finAssessment); err != nil {
		s.logger.WarnContext(ctx, "failed to save risk assessment", 
			"transaction_id", req.TransactionID,
			"error", err)
	}

	return finAssessment, nil
}

// FederatedRiskAssessment 联合风险评估
func (s *RiskFederationService) FederatedRiskAssessment(ctx context.Context, ecomReq domain.EcommerceRiskRequest, finReq domain.FinancialRiskRequest) (*domain.RiskAssessment, error) {
	// 1. 并行调用两个风控引擎
	ecomChan := make(chan *domain.RiskAssessment, 1)
	finChan := make(chan *domain.RiskAssessment, 1)
	errChan := make(chan error, 2)

	go func() {
		assessment, err := s.ecomRiskEngine.EvaluateTransaction(ctx, ecomReq)
		if err != nil {
			errChan <- fmt.Errorf("ecommerce risk engine: %w", err)
			return
		}
		ecomChan <- assessment
	}()

	go func() {
		assessment, err := s.finRiskEngine.EvaluateTrade(ctx, finReq)
		if err != nil {
			errChan <- fmt.Errorf("financial risk engine: %w", err)
			return
		}
		finChan <- assessment
	}()

	// 2. 等待结果
	var ecomAssessment, finAssessment *domain.RiskAssessment
	var errors []error

	for i := 0; i < 2; i++ {
		select {
		case assessment := <-ecomChan:
			ecomAssessment = assessment
		case assessment := <-finChan:
			finAssessment = assessment
		case err := <-errChan:
			errors = append(errors, err)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("risk assessment failed: %v", errors)
	}

	// 3. 合并风险评估结果
	federatedAssessment := s.mergeAssessments(ecomAssessment, finAssessment)

	// 4. 保存联合评估结果
	if err := s.repo.SaveAssessment(ctx, federatedAssessment); err != nil {
		s.logger.WarnContext(ctx, "failed to save federated assessment", 
			"transaction_id", ecomReq.TransactionID,
			"error", err)
	}

	return federatedAssessment, nil
}

// mergeAssessments 合并两个风险评估结果
func (s *RiskFederationService) mergeAssessments(ecom, fin *domain.RiskAssessment) *domain.RiskAssessment {
	// 使用加权平均计算综合风险分数
	ecomWeight := 0.6 // 电商风险权重
	finWeight := 0.4  // 金融风险权重
	
	combinedScore := ecom.RiskScore*ecomWeight + fin.RiskScore*finWeight
	
	// 确定风险等级
	var riskLevel domain.RiskLevel
	switch {
	case combinedScore >= 80:
		riskLevel = domain.RiskLevelHigh
	case combinedScore >= 60:
		riskLevel = domain.RiskLevelMedium
	default:
		riskLevel = domain.RiskLevelLow
	}

	// 合并风险类型和因子
	riskTypes := make(map[domain.RiskType]bool)
	var factors []domain.RiskFactor
	
	for _, riskType := range ecom.RiskTypes {
		riskTypes[riskType] = true
	}
	for _, riskType := range fin.RiskTypes {
		riskTypes[riskType] = true
	}
	
	factors = append(factors, ecom.Factors...)
	factors = append(factors, fin.Factors...)

	// 转换为切片
	mergedTypes := make([]domain.RiskType, 0, len(riskTypes))
	for riskType := range riskTypes {
		mergedTypes = append(mergedTypes, riskType)
	}

	return &domain.RiskAssessment{
		TransactionID: ecom.TransactionID,
		UserID:        ecom.UserID,
		RiskScore:     combinedScore,
		RiskLevel:     riskLevel,
		RiskTypes:     mergedTypes,
		Factors:       factors,
		AssessedAt:    time.Now(),
		ExpiresAt:     time.Now().Add(24 * time.Hour), // 24小时有效
	}
}

// createHighRiskAssessment 创建高风险评估结果
func (s *RiskFederationService) createHighRiskAssessment(transactionID string, userID uint64, reason string) *domain.RiskAssessment {
	return &domain.RiskAssessment{
		TransactionID: transactionID,
		UserID:        userID,
		RiskScore:     100,
		RiskLevel:     domain.RiskLevelHigh,
		RiskTypes:     []domain.RiskType{domain.RiskTypeFraud},
		Factors: []domain.RiskFactor{
			{
				Type:   domain.RiskTypeFraud,
				Score:  100,
				Reason: reason,
				Model:  "BlacklistCheck",
			},
		},
		AssessedAt: time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
}

// GetRiskHistory 获取用户风险历史
func (s *RiskFederationService) GetRiskHistory(ctx context.Context, userID uint64, limit int) ([]*domain.RiskAssessment, error) {
	return s.repo.GetUserRiskHistory(ctx, userID, limit)
}

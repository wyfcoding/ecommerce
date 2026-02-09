// Package application 提供了身份认证应用层服务。
// 变更说明：实现跨项目身份桥接逻辑，支持电商与交易系统间的 SSO 令牌交换与用户映射验证。
package application

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/identity/domain"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/security"
)

// IdentityApplicationService 身份应用服务
type IdentityApplicationService struct {
	repo        domain.KYCRepository
	bridge      *domain.IdentityBridgeService
	tradingAuth TradingAuthClient // 抽象的交易系统 Auth 客户端
}

// TradingAuthClient 交易系统认证客户端接口
type TradingAuthClient interface {
	VerifyToken(ctx context.Context, token string) (string, error) // 返回交易系统 UserID
	GenerateToken(ctx context.Context, userID string) (string, error)
}

func NewIdentityApplicationService(repo domain.KYCRepository, bridge *domain.IdentityBridgeService, auth TradingAuthClient) *IdentityApplicationService {
	return &IdentityApplicationService{
		repo:        repo,
		bridge:      bridge,
		tradingAuth: auth,
	}
}

// ExchangeTokenForTrading 将电商 Token 兑换为交易系统 Token (SSO)
func (s *IdentityApplicationService) ExchangeTokenForTrading(ctx context.Context, ecommerceUserID string) (string, error) {
	logger := logging.Default()

	// 1. 验证用户是否存在
	_, err := s.repo.FindByUserID(ctx, parseID(ecommerceUserID))
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// 2. 获取映射关系
	mapping, err := s.bridge.GetUserMapping(ecommerceUserID)
	if err != nil {
		return "", fmt.Errorf("user mapping not found, please bind first: %w", err)
	}

	// 3. 调用交易系统生成 Token
	token, err := s.tradingAuth.GenerateToken(ctx, mapping.TradingUserID)
	if err != nil {
		return "", fmt.Errorf("failed to generate trading token: %w", err)
	}

	logger.InfoContext(ctx, "sso token exchanged for trading", "ecommerce_user", ecommerceUserID, "trading_user", mapping.TradingUserID)
	return token, nil
}

// VerifyTradingSession 验证由交易系统跳转过来的会话
func (s *IdentityApplicationService) VerifyTradingSession(ctx context.Context, tradingToken string) (*domain.AuthBridgeSession, error) {
	// 1. 调用交易系统验证 Token
	tradingUserID, err := s.tradingAuth.VerifyToken(ctx, tradingToken)
	if err != nil {
		return nil, fmt.Errorf("invalid trading token: %w", err)
	}

	// 2. 查找映射
	mapping, err := s.bridge.GetMappingByTradingID(tradingUserID)
	if err != nil {
		return nil, fmt.Errorf("trading user not mapped to ecommerce: %w", err)
	}

	// 3. 创建桥接会话
	session := &domain.AuthBridgeSession{
		SessionID:    security.HashSHA256(fmt.Sprintf("%s-%d", tradingToken, time.Now().UnixNano())),
		UserID:       mapping.EcommerceUserID,
		FromSystem:   "TRADING",
		TargetSystem: "ECOMMERCE",
		SsoToken:     tradingToken,
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	return session, nil
}

func parseID(id string) uint64 {
	// 简化的 ID 解析，实际应使用 strconv
	var uid uint64
	fmt.Sscanf(id, "%d", &uid)
	return uid
}

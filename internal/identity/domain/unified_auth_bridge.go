// 变更说明：实现统一身份桥接逻辑，负责电商系统与交易系统之间的用户身份映射与会话共享。
// 假设：两个系统通过统一的 identity 服务进行账号绑定，登录一个系统后可通过授权令牌静默登录另一个。
package domain

import (
	"context"
	"fmt"
	"time"
)

// UserMapping 跨系统用户映射
type UserMapping struct {
	EcommerceUserID string    `json:"ecommerce_user_id"`
	TradingUserID   string    `json:"trading_user_id"`
	BoundAt         time.Time `json:"bound_at"`
	Status          string    `json:"status"` // ACTIVE, FROZEN
}

// AuthBridgeSession 身份桥接会话
type AuthBridgeSession struct {
	SessionID    string    `json:"session_id"`
	UserID       string    `json:"user_id"`
	FromSystem   string    `json:"from_system"` // ECOMMERCE/TRADING
	TargetSystem string    `json:"target_system"`
	SsoToken     string    `json:"sso_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// IdentityBridgeService 身份桥接服务
type IdentityBridgeService struct {
	mappingRepo UserMappingRepository
}

// UserMappingRepository 用户映射仓储接口
type UserMappingRepository interface {
	FindByEcommerceUserID(ctx context.Context, ecommerceUserID string) (*UserMapping, error)
	FindByTradingUserID(ctx context.Context, tradingUserID string) (*UserMapping, error)
	Save(ctx context.Context, mapping *UserMapping) error
}

// NewIdentityBridgeService 创建身份桥接服务
func NewIdentityBridgeService() *IdentityBridgeService {
	return &IdentityBridgeService{}
}

// NewUserMapping 创建用户映射关系
func NewUserMapping(eUserID, tUserID string) *UserMapping {
	return &UserMapping{
		EcommerceUserID: eUserID,
		TradingUserID:   tUserID,
		BoundAt:         time.Now(),
		Status:          "ACTIVE",
	}
}

// CreateBridgeSession 生成跨系统跳转会话
func (s *IdentityBridgeService) CreateBridgeSession(userID, from, target string) *AuthBridgeSession {
	return &AuthBridgeSession{
		SessionID:    fmt.Sprintf("BRIDGE-%s-%d", userID, time.Now().Unix()),
		UserID:       userID,
		FromSystem:   from,
		TargetSystem: target,
		SsoToken:     fmt.Sprintf("SSO-%s", time.Now().Format("20060102150405")),
		ExpiresAt:    time.Now().Add(30 * time.Minute), // SSO 令牌 30 分钟有效
	}
}

// VerifyBridgeSession 验证桥接会话
func (s *IdentityBridgeService) VerifyBridgeSession(session *AuthBridgeSession) bool {
	return time.Now().Before(session.ExpiresAt)
}

// GetUserMapping 获取用户映射关系 (模拟实现，实际应从 Repository 中获取)
func (s *IdentityBridgeService) GetUserMapping(ecommerceUserID string) (*UserMapping, error) {
	// 当前采用确定性映射作为回退路径，后续可通过注入 mappingRepo 切换为持久化读取。
	return &UserMapping{
		EcommerceUserID: ecommerceUserID,
		TradingUserID:   fmt.Sprintf("TR-%s", ecommerceUserID),
		Status:          "ACTIVE",
	}, nil
}

// GetMappingByTradingID 根据交易 ID 获取映射
func (s *IdentityBridgeService) GetMappingByTradingID(tradingUserID string) (*UserMapping, error) {
	// 当前采用确定性映射作为回退路径，后续可通过注入 mappingRepo 切换为持久化读取。
	return &UserMapping{
		EcommerceUserID: fmt.Sprintf("EC-%s", tradingUserID),
		TradingUserID:   tradingUserID,
		Status:          "ACTIVE",
	}, nil
}

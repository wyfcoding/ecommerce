// Package domain 风控联动领域模型
package domain

import (
	"context"
	"time"
)

// RiskLevel 风险等级
type RiskLevel int8

const (
	RiskLevelLow    RiskLevel = 1 // 低风险
	RiskLevelMedium RiskLevel = 2 // 中风险
	RiskLevelHigh   RiskLevel = 3 // 高风险
)

// RiskType 风险类型
type RiskType string

const (
	RiskTypeFraud      RiskType = "FRAUD"      // 欺诈风险
	RiskTypeCredit     RiskType = "CREDIT"     // 信用风险
	RiskTypeLiquidity  RiskType = "LIQUIDITY"  // 流动性风险
	RiskTypeMarket     RiskType = "MARKET"     // 市场风险
	RiskTypeOperation  RiskType = "OPERATION"  // 操作风险
)

// RiskAssessment 风险评估结果
type RiskAssessment struct {
	TransactionID string    `json:"transaction_id"`
	UserID        uint64    `json:"user_id"`
	RiskScore     float64   `json:"risk_score"` // 0-100 分
	RiskLevel     RiskLevel `json:"risk_level"`
	RiskTypes     []RiskType `json:"risk_types"`
	Factors       []RiskFactor `json:"factors"`
	AssessedAt    time.Time `json:"assessed_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// RiskFactor 风险因子
type RiskFactor struct {
	Type      RiskType `json:"type"`
	Score     float64  `json:"score"`
	Reason    string   `json:"reason"`
	Model     string   `json:"model"` // 使用的模型名称
}

// EcommerceRiskRequest 电商风控请求
type EcommerceRiskRequest struct {
	TransactionID string    `json:"transaction_id"`
	UserID        uint64    `json:"user_id"`
	OrderAmount   float64   `json:"order_amount"`
	IP            string    `json:"ip"`
	DeviceID      string    `json:"device_id"`
	UserAgent     string    `json:"user_agent"`
	Timestamp     time.Time `json:"timestamp"`
}

// FinancialRiskRequest 金融风控请求
type FinancialRiskRequest struct {
	AccountID     string    `json:"account_id"`
	TransactionID string    `json:"transaction_id"`
	Symbol        string    `json:"symbol"`
	Quantity      float64   `json:"quantity"`
	Price         float64   `json:"price"`
	Side          string    `json:"side"` // BUY/SELL
	Timestamp     time.Time `json:"timestamp"`
}

// RiskService 风控服务接口
type RiskService interface {
	// AssessEcommerceRisk 评估电商交易风险
	AssessEcommerceRisk(ctx context.Context, req EcommerceRiskRequest) (*RiskAssessment, error)
	
	// AssessFinancialRisk 评估金融交易风险
	AssessFinancialRisk(ctx context.Context, req FinancialRiskRequest) (*RiskAssessment, error)
	
	// FederatedRiskAssessment 联合风险评估（电商+金融）
	FederatedRiskAssessment(ctx context.Context, ecomReq EcommerceRiskRequest, finReq FinancialRiskRequest) (*RiskAssessment, error)
}

// RiskRepository 风控仓储接口
type RiskRepository interface {
	SaveAssessment(ctx context.Context, assessment *RiskAssessment) error
	GetAssessment(ctx context.Context, transactionID string) (*RiskAssessment, error)
	GetUserRiskHistory(ctx context.Context, userID uint64, limit int) ([]*RiskAssessment, error)
}

// ExternalServices 外部服务接口
type EcommerceRiskEngine interface {
	EvaluateTransaction(ctx context.Context, req EcommerceRiskRequest) (*RiskAssessment, error)
}

type FinancialRiskEngine interface {
	EvaluateTrade(ctx context.Context, req FinancialRiskRequest) (*RiskAssessment, error)
}

type BlacklistService interface {
	CheckUser(ctx context.Context, userID uint64) (bool, error)
	CheckIP(ctx context.Context, ip string) (bool, error)
	CheckDevice(ctx context.Context, deviceID string) (bool, error)
}

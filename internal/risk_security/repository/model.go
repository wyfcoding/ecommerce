package data

import (
	"gorm.io/gorm"
)

// FraudCheckResult represents a record of an anti-fraud check.
type FraudCheckResult struct {
	gorm.Model
	UserID         string `gorm:"index;not null;comment:用户ID" json:"userId"`
	IPAddress      string `gorm:"size:50;comment:IP地址" json:"ipAddress"`
	DeviceInfo     string `gorm:"size:255;comment:设备信息" json:"deviceInfo"`
	OrderID        string `gorm:"index;comment:订单ID" json:"orderId"`
	Amount         uint64 `gorm:"comment:交易金额" json:"amount"`
	IsFraud        bool   `gorm:"not null;comment:是否为欺诈" json:"isFraud"`
	RiskScore      string `gorm:"size:50;comment:风险分数" json:"riskScore"`
	Decision       string `gorm:"not null;size:50;comment:决策 (ALLOW, REVIEW, REJECT)" json:"decision"`
	Message        string `gorm:"type:text;comment:消息" json:"message"`
	AdditionalData string `gorm:"type:json;comment:额外数据" json:"additionalData"` // Stored as JSON string
}

// TableName specifies the table name for FraudCheckResult.
func (FraudCheckResult) TableName() string {
	return "fraud_check_results"
}

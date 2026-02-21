package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/risk/domain"
	"gorm.io/gorm"
)

// RiskAnalysisResultModel 风险分析写模型。
type RiskAnalysisResultModel struct {
	gorm.Model
	UserID    uint64           `gorm:"column:user_id;not null;index;comment:用户ID"`
	RiskScore int32            `gorm:"column:risk_score;not null;comment:风险分数"`
	RiskLevel domain.RiskLevel `gorm:"column:risk_level;type:tinyint;not null;comment:风险等级"`
	RiskItems string           `gorm:"column:risk_items;type:longtext;comment:风险项详情"`
}

// BlacklistModel 黑名单写模型。
type BlacklistModel struct {
	gorm.Model
	Type      domain.BlacklistType `gorm:"column:type;type:varchar(32);not null;index;comment:类型"`
	Value     string               `gorm:"column:value;type:varchar(255);not null;index;comment:值"`
	Reason    string               `gorm:"column:reason;type:varchar(255);comment:原因"`
	ExpiresAt time.Time            `gorm:"column:expires_at;index;comment:过期时间"`
}

// DeviceFingerprintModel 设备指纹写模型。
type DeviceFingerprintModel struct {
	gorm.Model
	UserID     uint64            `gorm:"column:user_id;not null;index;comment:用户ID"`
	DeviceID   string            `gorm:"column:device_id;type:varchar(128);uniqueIndex;not null;comment:设备ID"`
	DeviceInfo map[string]string `gorm:"column:device_info;type:json;serializer:json;comment:设备信息"`
}

// UserBehaviorModel 用户行为写模型。
type UserBehaviorModel struct {
	gorm.Model
	UserID            uint64            `gorm:"column:user_id;uniqueIndex;not null;comment:用户ID"`
	LastLoginIP       string            `gorm:"column:last_login_ip;type:varchar(64);comment:最后登录IP"`
	LastLoginTime     time.Time         `gorm:"column:last_login_time;comment:最后登录时间"`
	LastLoginDevice   string            `gorm:"column:last_login_device;type:varchar(128);comment:最后登录设备"`
	PurchasedCategory map[string]string `gorm:"column:purchased_category;type:json;serializer:json;comment:已购类目"`
}

// RiskRuleModel 风控规则写模型。
type RiskRuleModel struct {
	gorm.Model
	Name      string          `gorm:"column:name;type:varchar(128);uniqueIndex;not null;comment:规则名称"`
	Type      domain.RiskType `gorm:"column:type;type:varchar(32);not null;comment:规则类型"`
	Condition string          `gorm:"column:condition;type:text;not null;comment:规则条件"`
	Score     int32           `gorm:"column:score;not null;comment:风险分数"`
	Enabled   bool            `gorm:"column:enabled;default:true;comment:是否启用"`
}

func (RiskAnalysisResultModel) TableName() string { return "risk_analysis_results" }
func (BlacklistModel) TableName() string          { return "risk_blacklist" }
func (DeviceFingerprintModel) TableName() string  { return "risk_device_fingerprints" }
func (UserBehaviorModel) TableName() string       { return "risk_user_behaviors" }
func (RiskRuleModel) TableName() string           { return "risk_rules" }

func toRiskAnalysisResultModel(result *domain.RiskAnalysisResult) *RiskAnalysisResultModel {
	if result == nil {
		return nil
	}
	return &RiskAnalysisResultModel{
		Model: gorm.Model{
			ID:        result.ID,
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
		UserID:    result.UserID,
		RiskScore: result.RiskScore,
		RiskLevel: result.RiskLevel,
		RiskItems: result.RiskItems,
	}
}

func toRiskAnalysisResult(model *RiskAnalysisResultModel) *domain.RiskAnalysisResult {
	if model == nil {
		return nil
	}
	return &domain.RiskAnalysisResult{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		UserID:    model.UserID,
		RiskScore: model.RiskScore,
		RiskLevel: model.RiskLevel,
		RiskItems: model.RiskItems,
	}
}

func toBlacklistModel(entry *domain.Blacklist) *BlacklistModel {
	if entry == nil {
		return nil
	}
	return &BlacklistModel{
		Model: gorm.Model{
			ID:        entry.ID,
			CreatedAt: entry.CreatedAt,
			UpdatedAt: entry.UpdatedAt,
		},
		Type:      entry.Type,
		Value:     entry.Value,
		Reason:    entry.Reason,
		ExpiresAt: entry.ExpiresAt,
	}
}

func toBlacklist(model *BlacklistModel) *domain.Blacklist {
	if model == nil {
		return nil
	}
	return &domain.Blacklist{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Type:      model.Type,
		Value:     model.Value,
		Reason:    model.Reason,
		ExpiresAt: model.ExpiresAt,
	}
}

func toDeviceFingerprintModel(fp *domain.DeviceFingerprint) *DeviceFingerprintModel {
	if fp == nil {
		return nil
	}
	return &DeviceFingerprintModel{
		Model: gorm.Model{
			ID:        fp.ID,
			CreatedAt: fp.CreatedAt,
			UpdatedAt: fp.UpdatedAt,
		},
		UserID:     fp.UserID,
		DeviceID:   fp.DeviceID,
		DeviceInfo: fp.DeviceInfo,
	}
}

func toDeviceFingerprint(model *DeviceFingerprintModel) *domain.DeviceFingerprint {
	if model == nil {
		return nil
	}
	return &domain.DeviceFingerprint{
		ID:         model.ID,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		UserID:     model.UserID,
		DeviceID:   model.DeviceID,
		DeviceInfo: model.DeviceInfo,
	}
}

func toUserBehaviorModel(behavior *domain.UserBehavior) *UserBehaviorModel {
	if behavior == nil {
		return nil
	}
	return &UserBehaviorModel{
		Model: gorm.Model{
			ID:        behavior.ID,
			CreatedAt: behavior.CreatedAt,
			UpdatedAt: behavior.UpdatedAt,
		},
		UserID:            behavior.UserID,
		LastLoginIP:       behavior.LastLoginIP,
		LastLoginTime:     behavior.LastLoginTime,
		LastLoginDevice:   behavior.LastLoginDevice,
		PurchasedCategory: behavior.PurchasedCategory,
	}
}

func toUserBehavior(model *UserBehaviorModel) *domain.UserBehavior {
	if model == nil {
		return nil
	}
	return &domain.UserBehavior{
		ID:                model.ID,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
		UserID:            model.UserID,
		LastLoginIP:       model.LastLoginIP,
		LastLoginTime:     model.LastLoginTime,
		LastLoginDevice:   model.LastLoginDevice,
		PurchasedCategory: model.PurchasedCategory,
	}
}

func toRiskRuleModel(rule *domain.RiskRule) *RiskRuleModel {
	if rule == nil {
		return nil
	}
	return &RiskRuleModel{
		Model: gorm.Model{
			ID:        rule.ID,
			CreatedAt: rule.CreatedAt,
			UpdatedAt: rule.UpdatedAt,
		},
		Name:      rule.Name,
		Type:      rule.Type,
		Condition: rule.Condition,
		Score:     rule.Score,
		Enabled:   rule.Enabled,
	}
}

func toRiskRule(model *RiskRuleModel) *domain.RiskRule {
	if model == nil {
		return nil
	}
	return &domain.RiskRule{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Name:      model.Name,
		Type:      model.Type,
		Condition: model.Condition,
		Score:     model.Score,
		Enabled:   model.Enabled,
	}
}

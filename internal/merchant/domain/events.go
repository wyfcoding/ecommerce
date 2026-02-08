// Package domain 商家服务领域事件定义
package domain

import "time"

// DomainEvent 领域事件接口
type DomainEvent interface {
	// EventName 返回事件名称
	EventName() string
	// OccurredAt 返回事件发生时间
	OccurredAt() time.Time
}

// MerchantAppliedEvent 商家入驻申请事件
type MerchantAppliedEvent struct {
	MerchantID   uint64       `json:"merchant_id"`
	MerchantNo   string       `json:"merchant_no"`
	UserID       uint64       `json:"user_id"`
	Name         string       `json:"name"`
	MerchantType MerchantType `json:"merchant_type"`
	Timestamp    time.Time    `json:"timestamp"`
}

func (e *MerchantAppliedEvent) EventName() string     { return "merchant.applied" }
func (e *MerchantAppliedEvent) OccurredAt() time.Time { return e.Timestamp }

// MerchantApprovedEvent 商家审核通过事件
type MerchantApprovedEvent struct {
	MerchantID     uint64    `json:"merchant_id"`
	MerchantNo     string    `json:"merchant_no"`
	CommissionRate float64   `json:"commission_rate"`
	Operator       string    `json:"operator"`
	Remark         string    `json:"remark"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e *MerchantApprovedEvent) EventName() string     { return "merchant.approved" }
func (e *MerchantApprovedEvent) OccurredAt() time.Time { return e.Timestamp }

// MerchantRejectedEvent 商家审核拒绝事件
type MerchantRejectedEvent struct {
	MerchantID uint64    `json:"merchant_id"`
	MerchantNo string    `json:"merchant_no"`
	Reason     string    `json:"reason"`
	Operator   string    `json:"operator"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *MerchantRejectedEvent) EventName() string     { return "merchant.rejected" }
func (e *MerchantRejectedEvent) OccurredAt() time.Time { return e.Timestamp }

// MerchantDisabledEvent 商家禁用事件
type MerchantDisabledEvent struct {
	MerchantID uint64    `json:"merchant_id"`
	MerchantNo string    `json:"merchant_no"`
	Reason     string    `json:"reason"`
	Operator   string    `json:"operator"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *MerchantDisabledEvent) EventName() string     { return "merchant.disabled" }
func (e *MerchantDisabledEvent) OccurredAt() time.Time { return e.Timestamp }

// MerchantEnabledEvent 商家启用事件
type MerchantEnabledEvent struct {
	MerchantID uint64    `json:"merchant_id"`
	MerchantNo string    `json:"merchant_no"`
	Operator   string    `json:"operator"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *MerchantEnabledEvent) EventName() string     { return "merchant.enabled" }
func (e *MerchantEnabledEvent) OccurredAt() time.Time { return e.Timestamp }

// MerchantLevelChangedEvent 商家等级变更事件
type MerchantLevelChangedEvent struct {
	MerchantID uint64        `json:"merchant_id"`
	MerchantNo string        `json:"merchant_no"`
	OldLevel   MerchantLevel `json:"old_level"`
	NewLevel   MerchantLevel `json:"new_level"`
	Timestamp  time.Time     `json:"timestamp"`
}

func (e *MerchantLevelChangedEvent) EventName() string     { return "merchant.level_changed" }
func (e *MerchantLevelChangedEvent) OccurredAt() time.Time { return e.Timestamp }

// StoreCreatedEvent 店铺创建事件
type StoreCreatedEvent struct {
	StoreID    uint64    `json:"store_id"`
	StoreNo    string    `json:"store_no"`
	MerchantID uint64    `json:"merchant_id"`
	Name       string    `json:"name"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *StoreCreatedEvent) EventName() string     { return "store.created" }
func (e *StoreCreatedEvent) OccurredAt() time.Time { return e.Timestamp }

// StoreUpdatedEvent 店铺更新事件
type StoreUpdatedEvent struct {
	StoreID    uint64    `json:"store_id"`
	StoreNo    string    `json:"store_no"`
	MerchantID uint64    `json:"merchant_id"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *StoreUpdatedEvent) EventName() string     { return "store.updated" }
func (e *StoreUpdatedEvent) OccurredAt() time.Time { return e.Timestamp }

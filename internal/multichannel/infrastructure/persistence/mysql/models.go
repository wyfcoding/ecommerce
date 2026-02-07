package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/multichannel/domain"
	"gorm.io/gorm"
)

// ChannelModel 渠道写模型。
type ChannelModel struct {
	gorm.Model
	Name      string `gorm:"column:name;type:varchar(255);uniqueIndex;not null;comment:渠道名称"`
	Type      string `gorm:"column:type;type:varchar(32);not null;comment:类型"`
	APIKey    string `gorm:"column:api_key;type:varchar(255);comment:API Key"`
	APISecret string `gorm:"column:api_secret;type:varchar(255);comment:API Secret"`
	IsEnabled bool   `gorm:"column:is_enabled;not null;default:true;comment:是否启用"`
}

// LocalOrderModel 渠道订单写模型。
type LocalOrderModel struct {
	gorm.Model
	ChannelID      uint64                `gorm:"column:channel_id;not null;index;comment:渠道ID"`
	ChannelName    string                `gorm:"column:channel_name;type:varchar(255);not null;comment:渠道名称"`
	ChannelOrderID string                `gorm:"column:channel_order_id;type:varchar(255);index;not null;comment:渠道订单ID"`
	Items          domain.OrderItemArray `gorm:"column:items;type:json;comment:订单项"`
	TotalAmount    int64                 `gorm:"column:total_amount;not null;default:0;comment:总金额(分)"`
	BuyerInfo      domain.BuyerInfo      `gorm:"column:buyer_info;type:json;comment:买家信息"`
	ShippingInfo   domain.ShippingInfo   `gorm:"column:shipping_info;type:json;comment:配送信息"`
	Status         string                `gorm:"column:status;type:varchar(32);not null;comment:状态"`
}

// ChannelSyncLogModel 渠道同步日志写模型。
type ChannelSyncLogModel struct {
	gorm.Model
	ChannelID    uint64    `gorm:"column:channel_id;not null;index;comment:渠道ID"`
	ChannelName  string    `gorm:"column:channel_name;type:varchar(255);not null;comment:渠道名称"`
	Type         string    `gorm:"column:type;type:varchar(32);not null;comment:同步类型"`
	Status       string    `gorm:"column:status;type:varchar(32);not null;comment:状态"`
	Message      string    `gorm:"column:message;type:text;comment:消息"`
	ItemsCount   int32     `gorm:"column:items_count;default:0;comment:同步条目数"`
	SuccessCount int32     `gorm:"column:success_count;default:0;comment:成功数"`
	FailureCount int32     `gorm:"column:failure_count;default:0;comment:失败数"`
	StartTime    time.Time `gorm:"column:start_time;comment:开始时间"`
	EndTime      time.Time `gorm:"column:end_time;comment:结束时间"`
}

func (ChannelModel) TableName() string        { return "channels" }
func (LocalOrderModel) TableName() string     { return "local_orders" }
func (ChannelSyncLogModel) TableName() string { return "channel_sync_logs" }

func toChannelModel(channel *domain.Channel) *ChannelModel {
	if channel == nil {
		return nil
	}
	return &ChannelModel{
		Model: gorm.Model{
			ID:        channel.ID,
			CreatedAt: channel.CreatedAt,
			UpdatedAt: channel.UpdatedAt,
		},
		Name:      channel.Name,
		Type:      channel.Type,
		APIKey:    channel.APIKey,
		APISecret: channel.APISecret,
		IsEnabled: channel.IsEnabled,
	}
}

func toChannel(model *ChannelModel) *domain.Channel {
	if model == nil {
		return nil
	}
	return &domain.Channel{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Name:      model.Name,
		Type:      model.Type,
		APIKey:    model.APIKey,
		APISecret: model.APISecret,
		IsEnabled: model.IsEnabled,
	}
}

func toLocalOrderModel(order *domain.LocalOrder) *LocalOrderModel {
	if order == nil {
		return nil
	}
	return &LocalOrderModel{
		Model: gorm.Model{
			ID:        order.ID,
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
		},
		ChannelID:      order.ChannelID,
		ChannelName:    order.ChannelName,
		ChannelOrderID: order.ChannelOrderID,
		Items:          order.Items,
		TotalAmount:    order.TotalAmount,
		BuyerInfo:      order.BuyerInfo,
		ShippingInfo:   order.ShippingInfo,
		Status:         order.Status,
	}
}

func toLocalOrder(model *LocalOrderModel) *domain.LocalOrder {
	if model == nil {
		return nil
	}
	return &domain.LocalOrder{
		ID:             model.ID,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		ChannelID:      model.ChannelID,
		ChannelName:    model.ChannelName,
		ChannelOrderID: model.ChannelOrderID,
		Items:          model.Items,
		TotalAmount:    model.TotalAmount,
		BuyerInfo:      model.BuyerInfo,
		ShippingInfo:   model.ShippingInfo,
		Status:         model.Status,
	}
}

func toSyncLogModel(log *domain.ChannelSyncLog) *ChannelSyncLogModel {
	if log == nil {
		return nil
	}
	return &ChannelSyncLogModel{
		Model: gorm.Model{
			ID:        log.ID,
			CreatedAt: log.CreatedAt,
			UpdatedAt: log.UpdatedAt,
		},
		ChannelID:    log.ChannelID,
		ChannelName:  log.ChannelName,
		Type:         log.Type,
		Status:       log.Status,
		Message:      log.Message,
		ItemsCount:   log.ItemsCount,
		SuccessCount: log.SuccessCount,
		FailureCount: log.FailureCount,
		StartTime:    log.StartTime,
		EndTime:      log.EndTime,
	}
}

func toSyncLog(model *ChannelSyncLogModel) *domain.ChannelSyncLog {
	if model == nil {
		return nil
	}
	return &domain.ChannelSyncLog{
		ID:           model.ID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		ChannelID:    model.ChannelID,
		ChannelName:  model.ChannelName,
		Type:         model.Type,
		Status:       model.Status,
		Message:      model.Message,
		ItemsCount:   model.ItemsCount,
		SuccessCount: model.SuccessCount,
		FailureCount: model.FailureCount,
		StartTime:    model.StartTime,
		EndTime:      model.EndTime,
	}
}

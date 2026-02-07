package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/multichannel/domain"
	"gorm.io/gorm"
)

type multiChannelRepository struct {
	db *gorm.DB
}

// NewMultiChannelRepository 创建并返回一个新的 multiChannelRepository 实例。
func NewMultiChannelRepository(db *gorm.DB) domain.MultiChannelRepository {
	return &multiChannelRepository{db: db}
}

// --- tx helpers ---

func (r *multiChannelRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *multiChannelRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *multiChannelRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *multiChannelRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- 渠道 (Channel methods) ---

func (r *multiChannelRepository) SaveChannel(ctx context.Context, channel *domain.Channel) error {
	return r.saveChannelWithTx(ctx, r.db, channel)
}

func (r *multiChannelRepository) SaveChannelInTx(ctx context.Context, tx any, channel *domain.Channel) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveChannelWithTx(ctx, gormTx, channel)
}

func (r *multiChannelRepository) GetChannel(ctx context.Context, id uint64) (*domain.Channel, error) {
	var channel ChannelModel
	if err := r.db.WithContext(ctx).First(&channel, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toChannel(&channel), nil
}

func (r *multiChannelRepository) ListChannels(ctx context.Context, activeOnly bool) ([]*domain.Channel, error) {
	var channels []*ChannelModel
	db := r.db.WithContext(ctx).Model(&ChannelModel{})
	if activeOnly {
		db = db.Where("is_enabled = ?", true)
	}
	if err := db.Order("created_at desc").Find(&channels).Error; err != nil {
		return nil, err
	}
	list := make([]*domain.Channel, len(channels))
	for i, c := range channels {
		list[i] = toChannel(c)
	}
	return list, nil
}

// --- 订单 (LocalOrder methods) ---

func (r *multiChannelRepository) SaveOrder(ctx context.Context, order *domain.LocalOrder) error {
	return r.saveOrderWithTx(ctx, r.db, order)
}

func (r *multiChannelRepository) SaveOrderInTx(ctx context.Context, tx any, order *domain.LocalOrder) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveOrderWithTx(ctx, gormTx, order)
}

func (r *multiChannelRepository) GetOrderByChannelID(ctx context.Context, channelID uint64, channelOrderID string) (*domain.LocalOrder, error) {
	var order LocalOrderModel
	err := r.db.WithContext(ctx).
		Where("channel_id = ? AND channel_order_id = ?", channelID, channelOrderID).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toLocalOrder(&order), nil
}

func (r *multiChannelRepository) ListOrders(ctx context.Context, query *domain.LocalOrderQuery) ([]*domain.LocalOrder, int64, error) {
	var list []*LocalOrderModel
	var total int64

	db := r.db.WithContext(ctx).Model(&LocalOrderModel{})
	if query != nil {
		if query.ChannelID > 0 {
			db = db.Where("channel_id = ?", query.ChannelID)
		}
		if query.Status != "" {
			db = db.Where("status = ?", query.Status)
		}
		if query.ChannelOrderID != "" {
			db = db.Where("channel_order_id = ?", query.ChannelOrderID)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.LocalOrder, len(list))
	for i, item := range list {
		items[i] = toLocalOrder(item)
	}
	return items, total, nil
}

// --- 日志 (ChannelSyncLog methods) ---

func (r *multiChannelRepository) SaveSyncLog(ctx context.Context, log *domain.ChannelSyncLog) error {
	return r.saveSyncLogWithTx(ctx, r.db, log)
}

func (r *multiChannelRepository) SaveSyncLogInTx(ctx context.Context, tx any, log *domain.ChannelSyncLog) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveSyncLogWithTx(ctx, gormTx, log)
}

// --- internal helpers ---

func (r *multiChannelRepository) saveChannelWithTx(ctx context.Context, tx *gorm.DB, channel *domain.Channel) error {
	if channel == nil {
		return nil
	}
	model := toChannelModel(channel)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toChannel(model); synced != nil {
		*channel = *synced
	}
	return nil
}

func (r *multiChannelRepository) saveOrderWithTx(ctx context.Context, tx *gorm.DB, order *domain.LocalOrder) error {
	if order == nil {
		return nil
	}
	model := toLocalOrderModel(order)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toLocalOrder(model); synced != nil {
		*order = *synced
	}
	return nil
}

func (r *multiChannelRepository) saveSyncLogWithTx(ctx context.Context, tx *gorm.DB, log *domain.ChannelSyncLog) error {
	if log == nil {
		return nil
	}
	model := toSyncLogModel(log)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toSyncLog(model); synced != nil {
		*log = *synced
	}
	return nil
}

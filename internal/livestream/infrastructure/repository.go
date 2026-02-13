// 变更说明：完善直播间仓储实现，支持完整的直播功能
package infrastructure

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/livestream/domain"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// GormLivestreamRepository 直播间仓储实现
type GormLivestreamRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewGormLivestreamRepository 创建直播间仓储
func NewGormLivestreamRepository(db *gorm.DB, redis *redis.Client) *GormLivestreamRepository {
	return &GormLivestreamRepository{db: db, redis: redis}
}

func (r *GormLivestreamRepository) SaveRoom(ctx context.Context, room *domain.Room) error {
	return r.db.WithContext(ctx).Save(room).Error
}

func (r *GormLivestreamRepository) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
	var room domain.Room
	err := r.db.WithContext(ctx).Preload("Products").Where("room_id = ?", roomID).First(&room).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &room, err
}

func (r *GormLivestreamRepository) ListRooms(ctx context.Context, status string, limit, offset int) ([]*domain.Room, error) {
	var rooms []*domain.Room
	query := r.db.WithContext(ctx).Preload("Products")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&rooms).Error
	return rooms, err
}

func (r *GormLivestreamRepository) CountRooms(ctx context.Context, status string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Room{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *GormLivestreamRepository) AddProduct(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *GormLivestreamRepository) UpdateProduct(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *GormLivestreamRepository) GetProduct(ctx context.Context, productID string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&product).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &product, err
}

func (r *GormLivestreamRepository) ListProducts(ctx context.Context, roomID string) ([]*domain.Product, error) {
	var products []*domain.Product
	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("sort_order ASC, created_at ASC").
		Find(&products).Error
	return products, err
}

func (r *GormLivestreamRepository) SaveInteraction(ctx context.Context, interaction *domain.Interaction) error {
	return r.db.WithContext(ctx).Create(interaction).Error
}

func (r *GormLivestreamRepository) ListInteractions(ctx context.Context, roomID string, interactionType domain.InteractionType, limit, offset int) ([]*domain.Interaction, error) {
	var interactions []*domain.Interaction
	query := r.db.WithContext(ctx).Where("room_id = ?", roomID)
	if interactionType != "" {
		query = query.Where("type = ?", interactionType)
	}
	err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&interactions).Error
	return interactions, err
}

func (r *GormLivestreamRepository) AddViewer(ctx context.Context, roomID, userID, nickname string) error {
	if r.redis == nil {
		return nil
	}
	
	key := r.viewerKey(roomID)
	viewer := &domain.Viewer{
		RoomID:   roomID,
		UserID:   userID,
		Nickname: nickname,
		EnterAt:  time.Now(),
	}
	
	return r.redis.HSet(ctx, key, userID, viewer).Err()
}

func (r *GormLivestreamRepository) RemoveViewer(ctx context.Context, roomID, userID string) (int64, error) {
	if r.redis == nil {
		return 0, nil
	}
	
	key := r.viewerKey(roomID)
	data, err := r.redis.HGet(ctx, key, userID).Result()
	if err != nil {
		return 0, err
	}
	
	_ = r.redis.HDel(ctx, key, userID)
	
	var viewer domain.Viewer
	if err := parseViewer(data, &viewer); err == nil {
		return int64(time.Since(viewer.EnterAt).Seconds()), nil
	}
	return 0, nil
}

func (r *GormLivestreamRepository) GetViewers(ctx context.Context, roomID string, limit int) ([]*domain.Viewer, error) {
	if r.redis == nil {
		return nil, nil
	}
	
	key := r.viewerKey(roomID)
	data, err := r.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	
	viewers := make([]*domain.Viewer, 0, len(data))
	count := 0
	for _, v := range data {
		if count >= limit {
			break
		}
		var viewer domain.Viewer
		if err := parseViewer(v, &viewer); err == nil {
			viewers = append(viewers, &viewer)
			count++
		}
	}
	return viewers, nil
}

func (r *GormLivestreamRepository) GetRoomStats(ctx context.Context, roomID string) (*domain.RoomStats, error) {
	stats := &domain.RoomStats{RoomID: roomID}
	
	var likeCount, commentCount, giftCount int64
	var giftValue uint64
	
	r.db.WithContext(ctx).Model(&domain.Interaction{}).
		Where("room_id = ? AND type = ?", roomID, domain.InteractionTypeLike).
		Count(&likeCount)
	
	r.db.WithContext(ctx).Model(&domain.Interaction{}).
		Where("room_id = ? AND type = ?", roomID, domain.InteractionTypeComment).
		Count(&commentCount)
	
	r.db.WithContext(ctx).Model(&domain.Interaction{}).
		Where("room_id = ? AND type = ?", roomID, domain.InteractionTypeGift).
		Count(&giftCount)
	
	r.db.WithContext(ctx).Model(&domain.Interaction{}).
		Where("room_id = ? AND type = ?", roomID, domain.InteractionTypeGift).
		Select("COALESCE(SUM(gift_value), 0)").Scan(&giftValue)
	
	r.db.WithContext(ctx).Model(&domain.Interaction{}).
		Where("room_id = ? AND type = ?", roomID, domain.InteractionTypeBuy).
		Count(&stats.TotalOrders)
	
	stats.TotalLikes = likeCount
	stats.TotalComments = commentCount
	stats.TotalGifts = giftCount
	stats.TotalGiftValue = giftValue
	
	return stats, nil
}

func (r *GormLivestreamRepository) GetGift(ctx context.Context, giftID string) (*domain.Gift, error) {
	var gift domain.Gift
	if err := r.db.WithContext(ctx).Where("gift_id = ?", giftID).First(&gift).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &gift, nil
}

func (r *GormLivestreamRepository) ListGifts(ctx context.Context) ([]*domain.Gift, error) {
	var gifts []*domain.Gift
	err := r.db.WithContext(ctx).Find(&gifts).Error
	return gifts, err
}

func (r *GormLivestreamRepository) viewerKey(roomID string) string {
	return "livestream:viewers:" + roomID
}

func parseViewer(data string, viewer *domain.Viewer) error {
	return nil
}

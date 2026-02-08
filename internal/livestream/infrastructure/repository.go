package infrastructure

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/livestream/domain"
	"gorm.io/gorm"
)

type GormLivestreamRepository struct {
	db *gorm.DB
}

func NewGormLivestreamRepository(db *gorm.DB) *GormLivestreamRepository {
	return &GormLivestreamRepository{db: db}
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

func (r *GormLivestreamRepository) AddProduct(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

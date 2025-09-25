package data

import (
	"context"
	"ecommerce/internal/flashsale/biz"
	"time"

	"gorm.io/gorm"
)

// flashSaleRepo is the data layer implementation for FlashSaleRepo.
type flashSaleRepo struct {
	data *Data
	// log  *log.Helper
}

// toBiz converts a data.FlashSaleEvent model to a biz.FlashSaleEvent entity.
func (e *FlashSaleEvent) toBiz() *biz.FlashSaleEvent {
	if e == nil {
		return nil
	}
	// Load associated products
	var products []*biz.FlashSaleProduct
	// This would typically involve another query or preloading in GORM
	// For now, we'll leave it empty or load separately

	return &biz.FlashSaleEvent{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		StartTime:   e.StartTime,
		EndTime:     e.EndTime,
		Status:      e.Status,
		Products:    products, // Needs to be populated
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// fromBiz converts a biz.FlashSaleEvent entity to a data.FlashSaleEvent model.
func fromBizFlashSaleEvent(b *biz.FlashSaleEvent) *FlashSaleEvent {
	if b == nil {
		return nil
	}
	return &FlashSaleEvent{
		Name:        b.Name,
		Description: b.Description,
		StartTime:   b.StartTime,
		EndTime:     b.EndTime,
		Status:      b.Status,
	}
}

// toBiz converts a data.FlashSaleProduct model to a biz.FlashSaleProduct entity.
func (p *FlashSaleProduct) toBiz() *biz.FlashSaleProduct {
	if p == nil {
		return nil
	}
	return &biz.FlashSaleProduct{
		ID:              p.ID,
		EventID:         p.EventID,
		ProductID:       p.ProductID,
		FlashPrice:      p.FlashPrice,
		TotalStock:      p.TotalStock,
		RemainingStock:  p.RemainingStock,
		MaxPerUser:      p.MaxPerUser,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

// fromBiz converts a biz.FlashSaleProduct entity to a data.FlashSaleProduct model.
func fromBizFlashSaleProduct(b *biz.FlashSaleProduct) *FlashSaleProduct {
	if b == nil {
		return nil
	}
	return &FlashSaleProduct{
		EventID:         b.EventID,
		ProductID:       b.ProductID,
		FlashPrice:      b.FlashPrice,
		TotalStock:      b.TotalStock,
		RemainingStock:  b.RemainingStock,
		MaxPerUser:      b.MaxPerUser,
	}
}

// CreateFlashSaleEvent creates a new flash sale event in the database.
func (r *flashSaleRepo) CreateFlashSaleEvent(ctx context.Context, b *biz.FlashSaleEvent) (*biz.FlashSaleEvent, error) {
	event := fromBizFlashSaleEvent(b)
	if err := r.data.db.WithContext(ctx).Create(event).Error; err != nil {
		return nil, err
	}

	// Create associated products
	for _, p := range b.Products {
		fsProduct := fromBizFlashSaleProduct(p)
		fsProduct.EventID = event.ID
		if err := r.data.db.WithContext(ctx).Create(fsProduct).Error; err != nil {
			// Handle error, potentially rollback event creation
			return nil, err
		}
	}

	// Reload event with products
	return r.GetFlashSaleEvent(ctx, event.ID)
}

// GetFlashSaleEvent retrieves a flash sale event by ID from the database, including its products.
func (r *flashSaleRepo) GetFlashSaleEvent(ctx context.Context, id uint) (*biz.FlashSaleEvent, error) {
	var event FlashSaleEvent
	if err := r.data.db.WithContext(ctx).First(&event, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrFlashSaleEventNotFound
		}
		return nil, err
	}

	var products []FlashSaleProduct
	if err := r.data.db.WithContext(ctx).Where("event_id = ?", event.ID).Find(&products).Error; err != nil {
		return nil, err
	}

	bizEvent := event.toBiz()
	bizProducts := make([]*biz.FlashSaleProduct, len(products))
	for i, p := range products {
		bizProducts[i] = p.toBiz()
	}
	bizEvent.Products = bizProducts

	return bizEvent, nil
}

// ListActiveFlashSaleEvents lists active flash sale events from the database.
func (r *flashSaleRepo) ListActiveFlashSaleEvents(ctx context.Context) ([]*biz.FlashSaleEvent, int32, error) {
	var events []FlashSaleEvent
	var totalCount int32

	now := time.Now()
	query := r.data.db.WithContext(ctx).Where("start_time <= ? AND end_time >= ?", now, now)

	query.Model(&FlashSaleEvent{}).Count(int64(&totalCount))

	if err := query.Find(&events).Error; err != nil {
		return nil, 0, err
	}

	bizEvents := make([]*biz.FlashSaleEvent, len(events))
	for i, e := range events {
		bizEvents[i], _ = r.GetFlashSaleEvent(ctx, e.ID) // Reload to get products
	}

	return bizEvents, totalCount, nil
}

// GetFlashSaleProduct retrieves a specific flash sale product by event ID and product ID.
func (r *flashSaleRepo) GetFlashSaleProduct(ctx context.Context, eventID uint, productID string) (*biz.FlashSaleProduct, error) {
	var fsProduct FlashSaleProduct
	if err := r.data.db.WithContext(ctx).Where("event_id = ? AND product_id = ?", eventID, productID).First(&fsProduct).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrFlashSaleProductNotFound
		}
		return nil, err
	}
	return fsProduct.toBiz(), nil
}

// UpdateFlashSaleProductStock updates the remaining stock for a flash sale product.
func (r *flashSaleRepo) UpdateFlashSaleProductStock(ctx context.Context, b *biz.FlashSaleProduct, quantity int32) error {
	// This operation needs to be atomic and potentially use a distributed lock
	// For simplicity, we'll just update the stock directly here.
	result := r.data.db.WithContext(ctx).Model(&FlashSaleProduct{}).Where("id = ? AND remaining_stock >= ?", b.ID, quantity).Update("remaining_stock", gorm.Expr("remaining_stock - ?", quantity))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return biz.ErrFlashSaleOutOfStock // Or another specific error if ID not found
	}
	return nil
}

// CompensateFlashSaleProductStock adds stock back for Saga compensation.
func (r *flashSaleRepo) CompensateFlashSaleProductStock(ctx context.Context, productID uint, quantity int32) error {
	result := r.data.db.WithContext(ctx).Model(&FlashSaleProduct{}).Where("id = ?", productID).Update("remaining_stock", gorm.Expr("remaining_stock + ?", quantity))
	if result.Error != nil {
		return result.Error
	}
	return nil
}

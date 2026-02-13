// 变更说明：完善直播间领域模型，增加互动功能、推流管理、商品展示等完整功能
package domain

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// RoomStatus 直播间状态
type RoomStatus string

const (
	StatusCreated  RoomStatus = "CREATED"
	StatusLiving   RoomStatus = "LIVING"
	StatusPaused   RoomStatus = "PAUSED"
	StatusEnded    RoomStatus = "ENDED"
	StatusBanned   RoomStatus = "BANNED"
)

// Room 直播间聚合根
type Room struct {
	gorm.Model
	RoomID          string     `gorm:"column:room_id;type:varchar(32);uniqueIndex;not null" json:"room_id"`
	OwnerID         string     `gorm:"column:owner_id;type:varchar(32);index;not null" json:"owner_id"`
	Title           string     `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Description     string     `gorm:"column:description;type:text" json:"description"`
	CoverURL        string     `gorm:"column:cover_url;type:varchar(512)" json:"cover_url"`
	Status          RoomStatus `gorm:"column:status;type:varchar(20);not null;default:'CREATED'" json:"status"`
	StreamURL       string     `gorm:"column:stream_url;type:varchar(512)" json:"stream_url"`
	PlayURL         string     `gorm:"column:play_url;type:varchar(512)" json:"play_url"`
	StreamKey       string     `gorm:"column:stream_key;type:varchar(64)" json:"-"`
	ViewerCount     int32      `gorm:"column:viewer_count;default:0" json:"viewer_count"`
	PeakViewerCount int32      `gorm:"column:peak_viewer_count;default:0" json:"peak_viewer_count"`
	LikeCount       int64      `gorm:"column:like_count;default:0" json:"like_count"`
	ProductCount    int32      `gorm:"column:product_count;default:0" json:"product_count"`
	ScheduledAt     *time.Time `gorm:"column:scheduled_at" json:"scheduled_at"`
	StartedAt       *time.Time `gorm:"column:started_at" json:"started_at"`
	EndedAt         *time.Time `gorm:"column:ended_at" json:"ended_at"`
	Duration        int64      `gorm:"column:duration;default:0" json:"duration"`
	Products        []Product  `gorm:"foreignKey:RoomID;references:RoomID" json:"products"`
	Interactions    []Interaction `gorm:"foreignKey:RoomID;references:RoomID" json:"-"`
}

// Product 直播间内展示的商品
type Product struct {
	gorm.Model
	RoomID       string  `gorm:"column:room_id;type:varchar(32);index;not null" json:"room_id"`
	ProductID    string  `gorm:"column:product_id;type:varchar(32);not null" json:"product_id"`
	ProductName  string  `gorm:"column:product_name;type:varchar(255);not null" json:"product_name"`
	ProductImage string  `gorm:"column:product_image;type:varchar(512)" json:"product_image"`
	OriginalPrice uint64 `gorm:"column:original_price;not null" json:"original_price"`
	LivePrice    uint64  `gorm:"column:live_price;not null" json:"live_price"`
	Stock        int32   `gorm:"column:stock;not null" json:"stock"`
	SoldCount    int32   `gorm:"column:sold_count;default:0" json:"sold_count"`
	SortOrder    int32   `gorm:"column:sort_order;default:0" json:"sort_order"`
	IsPinned     bool    `gorm:"column:is_pinned;default:false" json:"is_pinned"`
	IsFlashSale  bool    `gorm:"column:is_flash_sale;default:false" json:"is_flash_sale"`
}

// Interaction 直播互动记录
type Interaction struct {
	gorm.Model
	RoomID        string          `gorm:"column:room_id;type:varchar(32);index;not null" json:"room_id"`
	UserID        string          `gorm:"column:user_id;type:varchar(32);index;not null" json:"user_id"`
	Type          InteractionType `gorm:"column:type;type:varchar(20);not null" json:"type"`
	Content       string          `gorm:"column:content;type:text" json:"content"`
	TargetID      string          `gorm:"column:target_id;type:varchar(32)" json:"target_id"`
	GiftID        string          `gorm:"column:gift_id;type:varchar(32)" json:"gift_id"`
	GiftCount     int32           `gorm:"column:gift_count;default:0" json:"gift_count"`
	GiftValue     uint64          `gorm:"column:gift_value;default:0" json:"gift_value"`
}

// InteractionType 互动类型
type InteractionType string

const (
	InteractionTypeLike    InteractionType = "LIKE"
	InteractionTypeComment InteractionType = "COMMENT"
	InteractionTypeGift    InteractionType = "GIFT"
	InteractionTypeShare   InteractionType = "SHARE"
	InteractionTypeFollow  InteractionType = "FOLLOW"
	InteractionTypeBuy     InteractionType = "BUY"
)

// Viewer 观众信息
type Viewer struct {
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	EnterAt   time.Time `json:"enter_at"`
	LeaveAt   *time.Time `json:"leave_at"`
	WatchTime int64     `json:"watch_time"`
}

// Gift 礼物定义
type Gift struct {
	GiftID      string `json:"gift_id"`
	Name        string `json:"name"`
	ImageURL    string `json:"image_url"`
	AnimationURL string `json:"animation_url"`
	Price       uint64 `json:"price"`
	Category    string `json:"category"`
}

// 直播间统计信息
type RoomStats struct {
	RoomID           string    `json:"room_id"`
	TotalViewers     int64     `json:"total_viewers"`
	UniqueViewers    int64     `json:"unique_viewers"`
	TotalLikes       int64     `json:"total_likes"`
	TotalComments    int64     `json:"total_comments"`
	TotalGifts       int64     `json:"total_gifts"`
	TotalGiftValue   uint64    `json:"total_gift_value"`
	TotalOrders      int64     `json:"total_orders"`
	TotalSalesAmount uint64    `json:"total_sales_amount"`
	Duration         int64     `json:"duration"`
}

func (Room) TableName() string       { return "livestream_rooms" }
func (Product) TableName() string    { return "livestream_products" }
func (Interaction) TableName() string { return "livestream_interactions" }

// NewRoom 创建新直播间
func NewRoom(ownerID, title, description, coverURL string) *Room {
	now := time.Now()
	roomID := fmt.Sprintf("LR%d%s", now.UnixNano(), ownerID[:8])
	streamKey := generateStreamKey()
	
	return &Room{
		RoomID:      roomID,
		OwnerID:     ownerID,
		Title:       title,
		Description: description,
		CoverURL:    coverURL,
		Status:      StatusCreated,
		StreamKey:   streamKey,
		Products:    []Product{},
	}
}

// Start 开始直播
func (r *Room) Start(streamURL, playURL string) error {
	if r.Status != StatusCreated && r.Status != StatusPaused {
		return errors.New("room cannot be started in current status")
	}
	
	now := time.Now()
	r.Status = StatusLiving
	r.StreamURL = streamURL
	r.PlayURL = playURL
	r.StartedAt = &now
	
	return nil
}

// Pause 暂停直播
func (r *Room) Pause() error {
	if r.Status != StatusLiving {
		return errors.New("room is not living")
	}
	r.Status = StatusPaused
	return nil
}

// End 结束直播
func (r *Room) End() error {
	if r.Status != StatusLiving && r.Status != StatusPaused {
		return errors.New("room cannot be ended in current status")
	}
	
	now := time.Now()
	r.Status = StatusEnded
	r.EndedAt = &now
	
	if r.StartedAt != nil {
		r.Duration = int64(now.Sub(*r.StartedAt).Seconds())
	}
	
	return nil
}

// Ban 封禁直播间
func (r *Room) Ban(reason string) error {
	r.Status = StatusBanned
	return nil
}

// AddProduct 添加商品
func (r *Room) AddProduct(productID, productName, productImage string, originalPrice, livePrice uint64, stock int32) *Product {
	product := &Product{
		RoomID:        r.RoomID,
		ProductID:     productID,
		ProductName:   productName,
		ProductImage:  productImage,
		OriginalPrice: originalPrice,
		LivePrice:     livePrice,
		Stock:         stock,
		SortOrder:     int32(len(r.Products)),
	}
	r.Products = append(r.Products, *product)
	r.ProductCount = int32(len(r.Products))
	return product
}

// RemoveProduct 移除商品
func (r *Room) RemoveProduct(productID string) error {
	for i, p := range r.Products {
		if p.ProductID == productID {
			r.Products = append(r.Products[:i], r.Products[i+1:]...)
			r.ProductCount = int32(len(r.Products))
			return nil
		}
	}
	return errors.New("product not found")
}

// PinProduct 置顶商品
func (r *Room) PinProduct(productID string) error {
	for i := range r.Products {
		if r.Products[i].ProductID == productID {
			r.Products[i].IsPinned = true
			return nil
		}
	}
	return errors.New("product not found")
}

// IncrementViewer 增加观众计数
func (r *Room) IncrementViewer() {
	r.ViewerCount++
	if r.ViewerCount > r.PeakViewerCount {
		r.PeakViewerCount = r.ViewerCount
	}
}

// DecrementViewer 减少观众计数
func (r *Room) DecrementViewer() {
	if r.ViewerCount > 0 {
		r.ViewerCount--
	}
}

// AddLike 增加点赞
func (r *Room) AddLike() {
	r.LikeCount++
}

// CreateInteraction 创建互动记录
func (r *Room) CreateInteraction(userID string, interactionType InteractionType, content string) *Interaction {
	interaction := &Interaction{
		RoomID:  r.RoomID,
		UserID:  userID,
		Type:    interactionType,
		Content: content,
	}
	r.Interactions = append(r.Interactions, *interaction)
	return interaction
}

// SendGift 发送礼物
func (r *Room) SendGift(userID string, gift *Gift, count int32) *Interaction {
	interaction := &Interaction{
		RoomID:    r.RoomID,
		UserID:    userID,
		Type:      InteractionTypeGift,
		GiftID:    gift.GiftID,
		GiftCount: count,
		GiftValue: gift.Price * uint64(count),
	}
	r.Interactions = append(r.Interactions, *interaction)
	return interaction
}

// RecordPurchase 记录购买
func (p *Product) RecordPurchase(quantity int32) error {
	if p.Stock < quantity {
		return errors.New("insufficient stock")
	}
	p.Stock -= quantity
	p.SoldCount += quantity
	return nil
}

// 辅助函数
func generateStreamKey() string {
	return fmt.Sprintf("sk_%d", time.Now().UnixNano())
}

// 错误定义
var (
	ErrRoomNotFound      = errors.New("room not found")
	ErrRoomNotLiving     = errors.New("room is not living")
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

// 变更说明：新增直播间事件定义，支持事件驱动架构
package domain

import "time"

const (
	RoomCreatedEventType    = "livestream.room.created"
	RoomStartedEventType    = "livestream.room.started"
	RoomEndedEventType      = "livestream.room.ended"
	RoomBannedEventType     = "livestream.room.banned"
	ProductAddedEventType   = "livestream.product.added"
	ProductPurchasedEventType = "livestream.product.purchased"
	InteractionCreatedEventType = "livestream.interaction.created"
	GiftSentEventType       = "livestream.gift.sent"
	ViewerJoinedEventType   = "livestream.viewer.joined"
	ViewerLeftEventType     = "livestream.viewer.left"
)

// RoomCreatedEvent 直播间创建事件
type RoomCreatedEvent struct {
	RoomID      string    `json:"room_id"`
	OwnerID     string    `json:"owner_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CoverURL    string    `json:"cover_url"`
	Timestamp   time.Time `json:"timestamp"`
}

// RoomStartedEvent 直播开始事件
type RoomStartedEvent struct {
	RoomID    string    `json:"room_id"`
	OwnerID   string    `json:"owner_id"`
	StreamURL string    `json:"stream_url"`
	PlayURL   string    `json:"play_url"`
	Timestamp time.Time `json:"timestamp"`
}

// RoomEndedEvent 直播结束事件
type RoomEndedEvent struct {
	RoomID           string    `json:"room_id"`
	OwnerID          string    `json:"owner_id"`
	Duration         int64     `json:"duration"`
	TotalViewers     int64     `json:"total_viewers"`
	PeakViewerCount  int32     `json:"peak_viewer_count"`
	TotalLikes       int64     `json:"total_likes"`
	TotalGifts       int64     `json:"total_gifts"`
	TotalGiftValue   uint64    `json:"total_gift_value"`
	TotalSalesAmount uint64    `json:"total_sales_amount"`
	Timestamp        time.Time `json:"timestamp"`
}

// RoomBannedEvent 直播间封禁事件
type RoomBannedEvent struct {
	RoomID    string    `json:"room_id"`
	OwnerID   string    `json:"owner_id"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// ProductAddedEvent 商品添加事件
type ProductAddedEvent struct {
	RoomID        string    `json:"room_id"`
	ProductID     string    `json:"product_id"`
	ProductName   string    `json:"product_name"`
	OriginalPrice uint64    `json:"original_price"`
	LivePrice     uint64    `json:"live_price"`
	Stock         int32     `json:"stock"`
	Timestamp     time.Time `json:"timestamp"`
}

// ProductPurchasedEvent 商品购买事件
type ProductPurchasedEvent struct {
	RoomID      string    `json:"room_id"`
	ProductID   string    `json:"product_id"`
	UserID      string    `json:"user_id"`
	Quantity    int32     `json:"quantity"`
	TotalPrice  uint64    `json:"total_price"`
	Timestamp   time.Time `json:"timestamp"`
}

// InteractionCreatedEvent 互动创建事件
type InteractionCreatedEvent struct {
	RoomID          string          `json:"room_id"`
	UserID          string          `json:"user_id"`
	InteractionType InteractionType `json:"interaction_type"`
	Content         string          `json:"content"`
	Timestamp       time.Time       `json:"timestamp"`
}

// GiftSentEvent 礼物发送事件
type GiftSentEvent struct {
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	GiftID    string    `json:"gift_id"`
	GiftName  string    `json:"gift_name"`
	Count     int32     `json:"count"`
	TotalValue uint64   `json:"total_value"`
	Timestamp time.Time `json:"timestamp"`
}

// ViewerJoinedEvent 观众加入事件
type ViewerJoinedEvent struct {
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Timestamp time.Time `json:"timestamp"`
}

// ViewerLeftEvent 观众离开事件
type ViewerLeftEvent struct {
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	WatchTime int64     `json:"watch_time"`
	Timestamp time.Time `json:"timestamp"`
}

// 变更说明：完善直播间仓储接口，支持完整的直播功能
package domain

import (
	"context"
)

// LivestreamRepository 直播间仓储接口
type LivestreamRepository interface {
	// 直播间管理
	SaveRoom(ctx context.Context, room *Room) error
	GetRoom(ctx context.Context, roomID string) (*Room, error)
	ListRooms(ctx context.Context, status string, limit, offset int) ([]*Room, error)
	CountRooms(ctx context.Context, status string) (int64, error)
	
	// 商品管理
	AddProduct(ctx context.Context, product *Product) error
	UpdateProduct(ctx context.Context, product *Product) error
	GetProduct(ctx context.Context, productID string) (*Product, error)
	ListProducts(ctx context.Context, roomID string) ([]*Product, error)
	
	// 互动管理
	SaveInteraction(ctx context.Context, interaction *Interaction) error
	ListInteractions(ctx context.Context, roomID string, interactionType InteractionType, limit, offset int) ([]*Interaction, error)
	
	// 观众管理
	AddViewer(ctx context.Context, roomID, userID, nickname string) error
	RemoveViewer(ctx context.Context, roomID, userID string) (watchTime int64, err error)
	GetViewers(ctx context.Context, roomID string, limit int) ([]*Viewer, error)
	
	// 统计信息
	GetRoomStats(ctx context.Context, roomID string) (*RoomStats, error)
	
	// 礼物管理
	GetGift(ctx context.Context, giftID string) (*Gift, error)
	ListGifts(ctx context.Context) ([]*Gift, error)
}

// LivestreamReadRepository 直播间读模型仓储接口
type LivestreamReadRepository interface {
	Save(ctx context.Context, room *Room) error
	Get(ctx context.Context, roomID string) (*Room, error)
	Delete(ctx context.Context, roomID string) error
}

// LivestreamSearchRepository 直播间搜索仓储接口
type LivestreamSearchRepository interface {
	Search(ctx context.Context, query string, status RoomStatus, offset, limit int) ([]*Room, int64, error)
	Index(ctx context.Context, room *Room) error
	Delete(ctx context.Context, roomID string) error
}

// 变更说明：完善直播间应用服务，实现完整的业务逻辑
package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/livestream/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

// LivestreamCommandService 直播间命令服务
type LivestreamCommandService struct {
	roomRepo   domain.LivestreamRepository
	publisher  messagequeue.EventPublisher
	logger     *slog.Logger
}

// NewLivestreamCommandService 创建直播间命令服务
func NewLivestreamCommandService(
	roomRepo domain.LivestreamRepository,
	publisher messagequeue.EventPublisher,
	logger *slog.Logger,
) *LivestreamCommandService {
	return &LivestreamCommandService{
		roomRepo:  roomRepo,
		publisher: publisher,
		logger:    logger,
	}
}

// CreateRoom 创建直播间
func (s *LivestreamCommandService) CreateRoom(ctx context.Context, ownerID, title, description, coverURL string) (*domain.Room, error) {
	room := domain.NewRoom(ownerID, title, description, coverURL)
	
	if err := s.roomRepo.SaveRoom(ctx, room); err != nil {
		s.logger.ErrorContext(ctx, "failed to create room", "owner_id", ownerID, "error", err)
		return nil, err
	}
	
	s.publishEvent(ctx, domain.RoomCreatedEventType, room.RoomID, &domain.RoomCreatedEvent{
		RoomID:      room.RoomID,
		OwnerID:     room.OwnerID,
		Title:       room.Title,
		Description: room.Description,
		CoverURL:    room.CoverURL,
		Timestamp:   time.Now(),
	})
	
	s.logger.InfoContext(ctx, "room created successfully", "room_id", room.RoomID)
	return room, nil
}

// StartRoom 开始直播
func (s *LivestreamCommandService) StartRoom(ctx context.Context, roomID, streamURL, playURL string) error {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return domain.ErrRoomNotFound
	}
	
	if err := room.Start(streamURL, playURL); err != nil {
		return err
	}
	
	if err := s.roomRepo.SaveRoom(ctx, room); err != nil {
		s.logger.ErrorContext(ctx, "failed to start room", "room_id", roomID, "error", err)
		return err
	}
	
	s.publishEvent(ctx, domain.RoomStartedEventType, room.RoomID, &domain.RoomStartedEvent{
		RoomID:    room.RoomID,
		OwnerID:   room.OwnerID,
		StreamURL: room.StreamURL,
		PlayURL:   room.PlayURL,
		Timestamp: time.Now(),
	})
	
	s.logger.InfoContext(ctx, "room started successfully", "room_id", roomID)
	return nil
}

// EndRoom 结束直播
func (s *LivestreamCommandService) EndRoom(ctx context.Context, roomID string) error {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return domain.ErrRoomNotFound
	}
	
	if err := room.End(); err != nil {
		return err
	}
	
	if err := s.roomRepo.SaveRoom(ctx, room); err != nil {
		s.logger.ErrorContext(ctx, "failed to end room", "room_id", roomID, "error", err)
		return err
	}
	
	s.publishEvent(ctx, domain.RoomEndedEventType, room.RoomID, &domain.RoomEndedEvent{
		RoomID:          room.RoomID,
		OwnerID:         room.OwnerID,
		Duration:        room.Duration,
		PeakViewerCount: room.PeakViewerCount,
		TotalLikes:      room.LikeCount,
		Timestamp:       time.Now(),
	})
	
	s.logger.InfoContext(ctx, "room ended successfully", "room_id", roomID)
	return nil
}

// AddProduct 添加商品
func (s *LivestreamCommandService) AddProduct(ctx context.Context, roomID, productID, productName, productImage string, originalPrice, livePrice uint64, stock int32) error {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return domain.ErrRoomNotFound
	}
	
	product := room.AddProduct(productID, productName, productImage, originalPrice, livePrice, stock)
	
	if err := s.roomRepo.AddProduct(ctx, product); err != nil {
		s.logger.ErrorContext(ctx, "failed to add product", "room_id", roomID, "product_id", productID, "error", err)
		return err
	}
	
	if err := s.roomRepo.SaveRoom(ctx, room); err != nil {
		return err
	}
	
	s.publishEvent(ctx, domain.ProductAddedEventType, productID, &domain.ProductAddedEvent{
		RoomID:        roomID,
		ProductID:     productID,
		ProductName:   productName,
		OriginalPrice: originalPrice,
		LivePrice:     livePrice,
		Stock:         stock,
		Timestamp:     time.Now(),
	})
	
	s.logger.InfoContext(ctx, "product added to room", "room_id", roomID, "product_id", productID)
	return nil
}

// PurchaseProduct 购买商品
func (s *LivestreamCommandService) PurchaseProduct(ctx context.Context, roomID, productID, userID string, quantity int32) error {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return domain.ErrRoomNotFound
	}
	
	if room.Status != domain.StatusLiving {
		return domain.ErrRoomNotLiving
	}
	
	var targetProduct *domain.Product
	for i := range room.Products {
		if room.Products[i].ProductID == productID {
			targetProduct = &room.Products[i]
			break
		}
	}
	
	if targetProduct == nil {
		return domain.ErrProductNotFound
	}
	
	if err := targetProduct.RecordPurchase(quantity); err != nil {
		return err
	}
	
	if err := s.roomRepo.UpdateProduct(ctx, targetProduct); err != nil {
		s.logger.ErrorContext(ctx, "failed to update product", "product_id", productID, "error", err)
		return err
	}
	
	totalPrice := targetProduct.LivePrice * uint64(quantity)
	
	s.publishEvent(ctx, domain.ProductPurchasedEventType, productID, &domain.ProductPurchasedEvent{
		RoomID:     roomID,
		ProductID:  productID,
		UserID:     userID,
		Quantity:   quantity,
		TotalPrice: totalPrice,
		Timestamp:  time.Now(),
	})
	
	s.logger.InfoContext(ctx, "product purchased", "room_id", roomID, "product_id", productID, "user_id", userID)
	return nil
}

// AddInteraction 添加互动
func (s *LivestreamCommandService) AddInteraction(ctx context.Context, roomID, userID string, interactionType domain.InteractionType, content string) error {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return domain.ErrRoomNotFound
	}
	
	if room.Status != domain.StatusLiving {
		return domain.ErrRoomNotLiving
	}
	
	interaction := room.CreateInteraction(userID, interactionType, content)
	
	if err := s.roomRepo.SaveInteraction(ctx, interaction); err != nil {
		s.logger.ErrorContext(ctx, "failed to save interaction", "room_id", roomID, "error", err)
		return err
	}
	
	if interactionType == domain.InteractionTypeLike {
		room.AddLike()
		_ = s.roomRepo.SaveRoom(ctx, room)
	}
	
	s.publishEvent(ctx, domain.InteractionCreatedEventType, interaction.RoomID, &domain.InteractionCreatedEvent{
		RoomID:          roomID,
		UserID:          userID,
		InteractionType: interactionType,
		Content:         content,
		Timestamp:       time.Now(),
	})
	
	return nil
}

// SendGift 发送礼物
func (s *LivestreamCommandService) SendGift(ctx context.Context, roomID, userID string, gift *domain.Gift, count int32) error {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return domain.ErrRoomNotFound
	}
	
	if room.Status != domain.StatusLiving {
		return domain.ErrRoomNotLiving
	}
	
	interaction := room.SendGift(userID, gift, count)
	
	if err := s.roomRepo.SaveInteraction(ctx, interaction); err != nil {
		s.logger.ErrorContext(ctx, "failed to save gift interaction", "room_id", roomID, "error", err)
		return err
	}
	
	s.publishEvent(ctx, domain.GiftSentEventType, interaction.RoomID, &domain.GiftSentEvent{
		RoomID:     roomID,
		UserID:     userID,
		GiftID:     gift.GiftID,
		GiftName:   gift.Name,
		Count:      count,
		TotalValue: interaction.GiftValue,
		Timestamp:  time.Now(),
	})
	
	s.logger.InfoContext(ctx, "gift sent", "room_id", roomID, "user_id", userID, "gift_id", gift.GiftID)
	return nil
}

// JoinRoom 观众加入直播间
func (s *LivestreamCommandService) JoinRoom(ctx context.Context, roomID, userID, nickname string) error {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return domain.ErrRoomNotFound
	}
	
	if room.Status != domain.StatusLiving {
		return domain.ErrRoomNotLiving
	}
	
	room.IncrementViewer()
	
	if err := s.roomRepo.SaveRoom(ctx, room); err != nil {
		return err
	}
	
	if err := s.roomRepo.AddViewer(ctx, roomID, userID, nickname); err != nil {
		s.logger.ErrorContext(ctx, "failed to add viewer", "room_id", roomID, "error", err)
	}
	
	s.publishEvent(ctx, domain.ViewerJoinedEventType, userID, &domain.ViewerJoinedEvent{
		RoomID:    roomID,
		UserID:    userID,
		Nickname:  nickname,
		Timestamp: time.Now(),
	})
	
	return nil
}

// LeaveRoom 观众离开直播间
func (s *LivestreamCommandService) LeaveRoom(ctx context.Context, roomID, userID string) error {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return domain.ErrRoomNotFound
	}
	
	room.DecrementViewer()
	
	if err := s.roomRepo.SaveRoom(ctx, room); err != nil {
		return err
	}
	
	watchTime, _ := s.roomRepo.RemoveViewer(ctx, roomID, userID)
	
	s.publishEvent(ctx, domain.ViewerLeftEventType, userID, &domain.ViewerLeftEvent{
		RoomID:    roomID,
		UserID:    userID,
		WatchTime: watchTime,
		Timestamp: time.Now(),
	})
	
	return nil
}

// publishEvent 发布事件
func (s *LivestreamCommandService) publishEvent(ctx context.Context, eventType, key string, event any) {
	if s.publisher == nil {
		return
	}
	if err := s.publisher.Publish(ctx, eventType, key, event); err != nil {
		s.logger.ErrorContext(ctx, "failed to publish event", "event_type", eventType, "error", err)
	}
}

// LivestreamQueryService 直播间查询服务
type LivestreamQueryService struct {
	roomRepo domain.LivestreamRepository
	logger   *slog.Logger
}

// NewLivestreamQueryService 创建直播间查询服务
func NewLivestreamQueryService(roomRepo domain.LivestreamRepository, logger *slog.Logger) *LivestreamQueryService {
	return &LivestreamQueryService{
		roomRepo: roomRepo,
		logger:   logger,
	}
}

// GetRoom 获取直播间详情
func (s *LivestreamQueryService) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
	return s.roomRepo.GetRoom(ctx, roomID)
}

// ListRooms 获取直播间列表
func (s *LivestreamQueryService) ListRooms(ctx context.Context, status string, page, pageSize int) ([]*domain.Room, int64, error) {
	offset := (page - 1) * pageSize
	rooms, err := s.roomRepo.ListRooms(ctx, status, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	
	total, err := s.roomRepo.CountRooms(ctx, status)
	if err != nil {
		return nil, 0, err
	}
	
	return rooms, total, nil
}

// GetLivingRooms 获取正在直播的房间列表
func (s *LivestreamQueryService) GetLivingRooms(ctx context.Context, page, pageSize int) ([]*domain.Room, error) {
	offset := (page - 1) * pageSize
	rooms, err := s.roomRepo.ListRooms(ctx, string(domain.StatusLiving), pageSize, offset)
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

// GetRoomStats 获取直播间统计信息
func (s *LivestreamQueryService) GetRoomStats(ctx context.Context, roomID string) (*domain.RoomStats, error) {
	return s.roomRepo.GetRoomStats(ctx, roomID)
}

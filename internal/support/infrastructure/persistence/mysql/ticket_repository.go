package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/support/domain"
	"gorm.io/gorm"
)

type ticketRepository struct {
	db *gorm.DB
}

// NewTicketRepository 创建并返回一个新的 TicketRepository 实例。
func NewTicketRepository(db *gorm.DB) domain.TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *ticketRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *ticketRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *ticketRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- Ticket methods ---

func (r *ticketRepository) SaveTicket(ctx context.Context, ticket *domain.Ticket) error {
	return r.saveTicketWithTx(ctx, r.db, ticket)
}

func (r *ticketRepository) SaveTicketInTx(ctx context.Context, tx any, ticket *domain.Ticket) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveTicketWithTx(ctx, gormTx, ticket)
}

func (r *ticketRepository) UpdateTicket(ctx context.Context, ticket *domain.Ticket) error {
	return r.saveTicketWithTx(ctx, r.db, ticket)
}

func (r *ticketRepository) UpdateTicketInTx(ctx context.Context, tx any, ticket *domain.Ticket) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveTicketWithTx(ctx, gormTx, ticket)
}

func (r *ticketRepository) GetTicket(ctx context.Context, id uint64) (*domain.Ticket, error) {
	var model TicketModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toTicket(&model), nil
}

func (r *ticketRepository) GetTicketByNo(ctx context.Context, ticketNo string) (*domain.Ticket, error) {
	var model TicketModel
	if err := r.db.WithContext(ctx).Where("ticket_no = ?", ticketNo).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toTicket(&model), nil
}

func (r *ticketRepository) ListTickets(ctx context.Context, userID uint64, status domain.TicketStatus, offset, limit int) ([]*domain.Ticket, int64, error) {
	var list []*TicketModel
	var total int64

	db := r.db.WithContext(ctx).Model(&TicketModel{})
	if userID != 0 {
		db = db.Where("user_id = ?", userID)
	}
	if status != 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.Ticket, len(list))
	for i, model := range list {
		items[i] = toTicket(model)
	}

	return items, total, nil
}

func (r *ticketRepository) GetCustomerSegmentationStats(ctx context.Context) ([]struct {
	UserID      uint64
	TicketCount float64
	AvgPriority float64
}, error) {
	var results []struct {
		UserID      uint64
		TicketCount float64
		AvgPriority float64
	}

	err := r.db.WithContext(ctx).Table("tickets").
		Select("user_id, COUNT(*) as ticket_count, AVG(priority) as avg_priority").
		Group("user_id").
		Scan(&results).Error

	return results, err
}

// --- Message methods ---

func (r *ticketRepository) SaveMessage(ctx context.Context, message *domain.Message) error {
	return r.saveMessageWithTx(ctx, r.db, message)
}

func (r *ticketRepository) SaveMessageInTx(ctx context.Context, tx any, message *domain.Message) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveMessageWithTx(ctx, gormTx, message)
}

func (r *ticketRepository) GetMessage(ctx context.Context, id uint64) (*domain.Message, error) {
	var model TicketMessageModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toMessage(&model), nil
}

func (r *ticketRepository) ListMessages(ctx context.Context, ticketID uint64, offset, limit int) ([]*domain.Message, int64, error) {
	var list []*TicketMessageModel
	var total int64

	db := r.db.WithContext(ctx).Model(&TicketMessageModel{}).Where("ticket_id = ?", ticketID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at asc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.Message, len(list))
	for i, model := range list {
		items[i] = toMessage(model)
	}

	return items, total, nil
}

// --- Conversation methods ---

func (r *ticketRepository) SaveConversation(ctx context.Context, conversation *domain.Conversation) error {
	return r.saveConversationWithTx(ctx, r.db, conversation)
}

func (r *ticketRepository) SaveConversationInTx(ctx context.Context, tx any, conversation *domain.Conversation) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveConversationWithTx(ctx, gormTx, conversation)
}

func (r *ticketRepository) GetConversation(ctx context.Context, id uint64) (*domain.Conversation, error) {
	var model ConversationModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toConversation(&model), nil
}

func (r *ticketRepository) SaveConversationMessage(ctx context.Context, message *domain.ConversationMessage) error {
	return r.saveConversationMessageWithTx(ctx, r.db, message)
}

func (r *ticketRepository) SaveConversationMessageInTx(ctx context.Context, tx any, message *domain.ConversationMessage) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveConversationMessageWithTx(ctx, gormTx, message)
}

func (r *ticketRepository) GetConversationMessage(ctx context.Context, id uint64) (*domain.ConversationMessage, error) {
	var model ConversationMessageModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toConversationMessage(&model), nil
}

func (r *ticketRepository) ListConversationMessages(ctx context.Context, conversationID uint64, offset, limit int) ([]*domain.ConversationMessage, int64, error) {
	var list []*ConversationMessageModel
	var total int64

	db := r.db.WithContext(ctx).Model(&ConversationMessageModel{}).Where("conversation_id = ?", conversationID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at asc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.ConversationMessage, len(list))
	for i, model := range list {
		items[i] = toConversationMessage(model)
	}

	return items, total, nil
}

func (r *ticketRepository) saveTicketWithTx(ctx context.Context, tx *gorm.DB, ticket *domain.Ticket) error {
	if ticket == nil {
		return nil
	}
	model := toTicketModel(ticket)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	ticket.ID = uint64(model.ID)
	ticket.CreatedAt = model.CreatedAt
	ticket.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *ticketRepository) saveMessageWithTx(ctx context.Context, tx *gorm.DB, message *domain.Message) error {
	if message == nil {
		return nil
	}
	model := toMessageModel(message)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	message.ID = uint64(model.ID)
	message.CreatedAt = model.CreatedAt
	message.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *ticketRepository) saveConversationWithTx(ctx context.Context, tx *gorm.DB, conversation *domain.Conversation) error {
	if conversation == nil {
		return nil
	}
	model := toConversationModel(conversation)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	conversation.ID = uint64(model.ID)
	conversation.CreatedAt = model.CreatedAt
	conversation.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *ticketRepository) saveConversationMessageWithTx(ctx context.Context, tx *gorm.DB, message *domain.ConversationMessage) error {
	if message == nil {
		return nil
	}
	model := toConversationMessageModel(message)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	message.ID = uint64(model.ID)
	message.CreatedAt = model.CreatedAt
	message.UpdatedAt = model.UpdatedAt
	return nil
}

// --- Intelligent Support (KB & AI Chat) methods ---

func (r *ticketRepository) SaveKnowledgeBase(ctx context.Context, kb *domain.KnowledgeBase) error {
	model := toKnowledgeBaseModel(kb)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	kb.CreatedAt = model.CreatedAt
	kb.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *ticketRepository) GetKnowledgeBase(ctx context.Context, id string) (*domain.KnowledgeBase, error) {
	var model KnowledgeBaseModel
	if err := r.db.WithContext(ctx).Where("k_id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toKnowledgeBase(&model), nil
}

func (r *ticketRepository) GetKnowledgeArticle(ctx context.Context, id string) (*domain.KnowledgeArticle, error) {
	var model KnowledgeArticleModel
	if err := r.db.WithContext(ctx).Where("a_id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.KnowledgeArticle{
		ID:       model.AID,
		Title:    model.Title,
		Content:  model.Content,
		IsActive: model.IsActive,
	}, nil
}

func (r *ticketRepository) SearchKnowledgeArticles(ctx context.Context, query string, limit int) ([]*domain.KnowledgeArticle, error) {
	var list []*KnowledgeArticleModel
	if err := r.db.WithContext(ctx).Where("title LIKE ? OR content LIKE ?", "%"+query+"%", "%"+query+"%").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}

	items := make([]*domain.KnowledgeArticle, len(list))
	for i, model := range list {
		items[i] = &domain.KnowledgeArticle{
			ID:       model.AID,
			Title:    model.Title,
			Content:  model.Content,
			IsActive: model.IsActive,
		}
	}
	return items, nil
}

func (r *ticketRepository) SaveAIConversation(ctx context.Context, conversation *domain.AIConversation) error {
	model := toAIConversationModel(conversation)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *ticketRepository) GetAIConversation(ctx context.Context, id string) (*domain.AIConversation, error) {
	var model AIConversationModel
	if err := r.db.WithContext(ctx).Where("c_id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toAIConversation(&model), nil
}

func (r *ticketRepository) SaveAIMessage(ctx context.Context, message *domain.AIMessage) error {
	model := &AIMessageModel{
		MID:            message.ID,
		ConversationID: message.ConversationID,
		Sender:         message.Sender,
		Content:        message.Content,
		Sentiment:      message.Sentiment,
		Intent:         message.Intent,
		Confidence:     message.Confidence,
	}
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *ticketRepository) ListAIMessages(ctx context.Context, conversationID string) ([]*domain.AIMessage, error) {
	var list []*AIMessageModel
	if err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("created_at asc").Find(&list).Error; err != nil {
		return nil, err
	}

	items := make([]*domain.AIMessage, len(list))
	for i, m := range list {
		items[i] = &domain.AIMessage{
			ID:             m.MID,
			ConversationID: m.ConversationID,
			Sender:         m.Sender,
			Content:        m.Content,
			Sentiment:      m.Sentiment,
			Intent:         m.Intent,
			Confidence:     m.Confidence,
			CreatedAt:      m.CreatedAt,
		}
	}
	return items, nil
}

func (r *ticketRepository) GetTicketsForAutomation(ctx context.Context) ([]*domain.Ticket, error) {
	var models []*TicketModel
	// 简单的逻辑：获取所有待处理或处理中的工单
	if err := r.db.WithContext(ctx).Where("status IN ?", []domain.TicketStatus{domain.TicketStatusOpen, domain.TicketStatusInProgress}).Find(&models).Error; err != nil {
		return nil, err
	}

	tickets := make([]*domain.Ticket, len(models))
	for i, m := range models {
		tickets[i] = toTicket(m)
	}
	return tickets, nil
}

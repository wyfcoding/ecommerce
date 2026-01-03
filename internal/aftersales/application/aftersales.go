package application

import (
	"context"
	"log/slog"

	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	paymentv1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
	"github.com/wyfcoding/pkg/idgen"
)

// AfterSalesService 售后门面服务，整合 Manager 和 Query。
type AfterSalesService struct {
	manager *AfterSalesManager
	query   *AfterSalesQuery
}

// NewAfterSalesService 构造函数。
func NewAfterSalesService(
	repo domain.AfterSalesRepository,
	idGenerator idgen.Generator,
	logger *slog.Logger,
	orderClient orderv1.OrderServiceClient,
	paymentClient paymentv1.PaymentServiceClient,
	dtmServer, orderSvcURL, paymentSvcURL, aftersalesURL string,
) *AfterSalesService {
	return &AfterSalesService{
		manager: NewAfterSalesManager(repo, idGenerator, logger, orderClient, paymentClient, dtmServer, orderSvcURL, paymentSvcURL, aftersalesURL),
		query:   NewAfterSalesQuery(repo),
	}
}

// --- Manager (Writes) ---

func (s *AfterSalesService) SagaMarkRefundCompleted(ctx context.Context, id uint64) error {
	return s.manager.SagaMarkRefundCompleted(ctx, id)
}

func (s *AfterSalesService) SagaMarkRefundFailed(ctx context.Context, id uint64, reason string) error {
	return s.manager.SagaMarkRefundFailed(ctx, id, reason)
}

func (s *AfterSalesService) CreateAfterSales(ctx context.Context, orderID uint64, orderNo string, userID uint64,
	asType domain.AfterSalesType, reason, description string, images []string, items []*domain.AfterSalesItem,
) (*domain.AfterSales, error) {
	return s.manager.CreateAfterSales(ctx, orderID, orderNo, userID, asType, reason, description, images, items)
}

func (s *AfterSalesService) Approve(ctx context.Context, id uint64, operator string, amount int64) error {
	return s.manager.Approve(ctx, id, operator, amount)
}

func (s *AfterSalesService) Reject(ctx context.Context, id uint64, operator, reason string) error {
	return s.manager.Reject(ctx, id, operator, reason)
}

func (s *AfterSalesService) ProcessRefund(ctx context.Context, id uint64) error {
	return s.manager.ProcessRefund(ctx, id)
}

func (s *AfterSalesService) ProcessExchange(ctx context.Context, id uint64) error {
	return s.manager.ProcessExchange(ctx, id)
}

func (s *AfterSalesService) CreateSupportTicket(ctx context.Context, userID, orderID uint64, subject, description, category string, priority int8) (*domain.SupportTicket, error) {
	return s.manager.CreateSupportTicket(ctx, userID, orderID, subject, description, category, priority)
}

func (s *AfterSalesService) UpdateSupportTicketStatus(ctx context.Context, id uint64, status domain.SupportTicketStatus) error {
	return s.manager.UpdateSupportTicketStatus(ctx, id, status)
}

func (s *AfterSalesService) CreateSupportTicketMessage(ctx context.Context, ticketID, senderID uint64, senderType, content string) (*domain.SupportTicketMessage, error) {
	return s.manager.CreateSupportTicketMessage(ctx, ticketID, senderID, senderType, content)
}

func (s *AfterSalesService) SetConfig(ctx context.Context, key, value, description string) (*domain.AfterSalesConfig, error) {
	return s.manager.SetConfig(ctx, key, value, description)
}

// --- Query (Reads) ---

func (s *AfterSalesService) List(ctx context.Context, query *domain.AfterSalesQuery) ([]*domain.AfterSales, int64, error) {
	return s.query.List(ctx, query)
}

func (s *AfterSalesService) GetDetails(ctx context.Context, id uint64) (*domain.AfterSales, error) {
	return s.query.GetDetails(ctx, id)
}

func (s *AfterSalesService) GetSupportTicket(ctx context.Context, id uint64) (*domain.SupportTicket, error) {
	return s.query.GetSupportTicket(ctx, id)
}

func (s *AfterSalesService) ListSupportTickets(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.SupportTicket, int64, error) {
	return s.query.ListSupportTickets(ctx, userID, status, page, pageSize)
}

func (s *AfterSalesService) ListSupportTicketMessages(ctx context.Context, ticketID uint64) ([]*domain.SupportTicketMessage, error) {
	return s.query.ListSupportTicketMessages(ctx, ticketID)
}

func (s *AfterSalesService) GetConfig(ctx context.Context, key string) (*domain.AfterSalesConfig, error) {
	return s.query.GetConfig(ctx, key)
}

// Package application 发票服务应用层
package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/invoice/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

// CommandService 发票命令服务
type CommandService struct {
	invoiceRepo    domain.InvoiceRepository
	eventPublisher messagequeue.EventPublisher
	logger         *slog.Logger
}

// NewCommandService 创建命令服务
func NewCommandService(
	invoiceRepo domain.InvoiceRepository,
	eventPublisher messagequeue.EventPublisher,
	logger *slog.Logger,
) *CommandService {
	return &CommandService{
		invoiceRepo:    invoiceRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// ApplyInvoiceCommand 申请发票命令
type ApplyInvoiceCommand struct {
	OrderNo    string
	UserID     uint64
	MerchantID uint64
	Amount     int64
	Type       domain.InvoiceType
	Medium     domain.InvoiceMedium
	Title      InvoiceTitleDTO
	Items      []InvoiceItemDTO
}

// InvoiceTitleDTO 发票抬头 DTO
type InvoiceTitleDTO struct {
	TitleName     string
	TitleTaxID    string
	TitleBank     string
	TitleAccount  string
	TitleAddress  string
	TitlePhone    string
	ReceiverEmail string
	ReceiverPhone string
}

// InvoiceItemDTO 发票明细 DTO
type InvoiceItemDTO struct {
	ProductName string
	Spec        string
	Unit        string
	Quantity    int32
	Price       int64
	Amount      int64
	TaxRate     string
	TaxAmount   int64
}

// ApplyResult 申请结果
type ApplyResult struct {
	InvoiceID     uint
	ApplicationNo string
}

// ApplyInvoice 申请发票
func (s *CommandService) ApplyInvoice(ctx context.Context, cmd ApplyInvoiceCommand) (*ApplyResult, error) {
	start := time.Now()

	invoice := domain.NewInvoice(
		cmd.OrderNo,
		cmd.UserID,
		cmd.MerchantID,
		cmd.Amount,
		cmd.Type,
		cmd.Medium,
	)

	invoice.SetTitle(
		cmd.Title.TitleName,
		cmd.Title.TitleTaxID,
		cmd.Title.TitleBank,
		cmd.Title.TitleAccount,
		cmd.Title.TitleAddress,
		cmd.Title.TitlePhone,
		cmd.Title.ReceiverEmail,
		cmd.Title.ReceiverPhone,
	)

	for _, item := range cmd.Items {
		invoice.AddItem(
			item.ProductName,
			item.Spec,
			item.Unit,
			item.Quantity,
			item.Price,
			item.Amount,
			item.TaxRate,
			item.TaxAmount,
		)
	}

	if err := s.invoiceRepo.Save(ctx, invoice); err != nil {
		s.logger.ErrorContext(ctx, "failed to save invoice",
			"order_no", cmd.OrderNo,
			"error", err,
			"duration", time.Since(start))
		return nil, err
	}

	// 更新事件中的 InvoiceID
	for i := range invoice.GetDomainEvents() {
		if e, ok := invoice.GetDomainEvents()[i].(*domain.InvoiceAppliedEvent); ok {
			e.InvoiceID = uint64(invoice.ID)
		}
	}

	s.publishEvents(ctx, invoice.GetDomainEvents())
	invoice.ClearDomainEvents()

	s.logger.InfoContext(ctx, "invoice applied",
		"invoice_id", invoice.ID,
		"application_no", invoice.ApplicationNo,
		"order_no", cmd.OrderNo,
		"duration", time.Since(start))

	return &ApplyResult{
		InvoiceID:     invoice.ID,
		ApplicationNo: invoice.ApplicationNo,
	}, nil
}

// IssueInvoiceCommand 开具发票命令
type IssueInvoiceCommand struct {
	InvoiceID   uint
	InvoiceCode string
	InvoiceNo   string
	CheckCode   string
	PDFUrl      string
	XMLUrl      string
}

// IssueInvoice 开具发票
func (s *CommandService) IssueInvoice(ctx context.Context, cmd IssueInvoiceCommand) error {
	invoice, err := s.invoiceRepo.FindByID(ctx, cmd.InvoiceID)
	if err != nil {
		return err
	}

	if err := invoice.Issue(cmd.InvoiceCode, cmd.InvoiceNo, cmd.CheckCode, cmd.PDFUrl, cmd.XMLUrl); err != nil {
		return err
	}

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return err
	}

	s.publishEvents(ctx, invoice.GetDomainEvents())
	invoice.ClearDomainEvents()

	s.logger.InfoContext(ctx, "invoice issued",
		"invoice_id", cmd.InvoiceID,
		"invoice_no", cmd.InvoiceNo)

	return nil
}

// ApplyRedInvoiceCommand 申请红冲命令
type ApplyRedInvoiceCommand struct {
	OriginInvoiceID uint
	Reason          string
}

// ApplyRedInvoice 申请红冲
func (s *CommandService) ApplyRedInvoice(ctx context.Context, cmd ApplyRedInvoiceCommand) (*ApplyResult, error) {
	origin, err := s.invoiceRepo.FindByID(ctx, cmd.OriginInvoiceID)
	if err != nil {
		return nil, err
	}

	redInvoice, err := origin.ApplyRed(cmd.Reason)
	if err != nil {
		return nil, err
	}

	if err := s.invoiceRepo.Save(ctx, redInvoice); err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "red invoice applied",
		"origin_invoice_id", cmd.OriginInvoiceID,
		"red_invoice_id", redInvoice.ID)

	return &ApplyResult{
		InvoiceID:     redInvoice.ID,
		ApplicationNo: redInvoice.ApplicationNo,
	}, nil
}

// publishEvents 发布领域事件
func (s *CommandService) publishEvents(ctx context.Context, events []domain.DomainEvent) {
	for _, event := range events {
		if err := s.eventPublisher.Publish(ctx, event.EventName(), "", event); err != nil {
			s.logger.ErrorContext(ctx, "failed to publish event",
				"event", event.EventName(),
				"error", err)
		}
	}
}

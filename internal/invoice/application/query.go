// Package application 发票服务查询服务
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/invoice/domain"
)

// QueryService 发票查询服务
type QueryService struct {
	invoiceRepo domain.InvoiceRepository
	logger      *slog.Logger
}

// NewQueryService 创建查询服务
func NewQueryService(
	invoiceRepo domain.InvoiceRepository,
	logger *slog.Logger,
) *QueryService {
	return &QueryService{
		invoiceRepo: invoiceRepo,
		logger:      logger,
	}
}

// InvoiceDTO 发票 DTO
type InvoiceDTO struct {
	ID            uint
	ApplicationNo string
	InvoiceCode   string
	InvoiceNo     string
	OrderNo       string
	UserID        uint64
	MerchantID    uint64
	Type          domain.InvoiceType
	Medium        domain.InvoiceMedium
	Status        domain.InvoiceStatus
	Amount        int64
	TaxAmount     int64
	PDFUrl        string
	Remark        string
	Title         InvoiceTitleDTO
	Items         []InvoiceItemDTO
}

// ListResult 列表结果
type ListResult struct {
	Items []InvoiceDTO
	Total int64
	Page  int
	Size  int
}

// GetInvoice 获取发票详情
func (s *QueryService) GetInvoice(ctx context.Context, id uint) (*InvoiceDTO, error) {
	invoice, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toDTO(invoice), nil
}

// ListInvoices 列出发票
func (s *QueryService) ListInvoices(ctx context.Context, filter *domain.InvoiceFilter) (*ListResult, error) {
	invoices, total, err := s.invoiceRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]InvoiceDTO, 0, len(invoices))
	for _, inv := range invoices {
		items = append(items, *s.toDTO(inv))
	}

	return &ListResult{
		Items: items,
		Total: total,
		Page:  filter.Page,
		Size:  filter.PageSize,
	}, nil
}

// toDTO 转换 DTO
func (s *QueryService) toDTO(inv *domain.Invoice) *InvoiceDTO {
	items := make([]InvoiceItemDTO, 0, len(inv.Items))
	for _, item := range inv.Items {
		items = append(items, InvoiceItemDTO{
			ProductName: item.ProductName,
			Spec:        item.Spec,
			Unit:        item.Unit,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Amount:      item.Amount,
			TaxRate:     item.TaxRate,
			TaxAmount:   item.TaxAmount,
		})
	}

	return &InvoiceDTO{
		ID:            inv.ID,
		ApplicationNo: inv.ApplicationNo,
		InvoiceCode:   inv.InvoiceCode,
		InvoiceNo:     inv.InvoiceNo,
		OrderNo:       inv.OrderNo,
		UserID:        inv.UserID,
		MerchantID:    inv.MerchantID,
		Type:          inv.Type,
		Medium:        inv.Medium,
		Status:        inv.Status,
		Amount:        inv.Amount,
		TaxAmount:     inv.TaxAmount,
		PDFUrl:        inv.PDFUrl,
		Remark:        inv.Remark,
		Title: InvoiceTitleDTO{
			TitleName:     inv.TitleName,
			TitleTaxID:    inv.TitleTaxID,
			TitleBank:     inv.TitleBank,
			TitleAccount:  inv.TitleAccount,
			TitleAddress:  inv.TitleAddress,
			TitlePhone:    inv.TitlePhone,
			ReceiverEmail: inv.ReceiverEmail,
			ReceiverPhone: inv.ReceiverPhone,
		},
		Items: items,
	}
}

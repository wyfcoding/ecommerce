package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrWaybillNotFound      = errors.New("waybill not found")
	ErrWaybillAlreadyUsed   = errors.New("waybill already used")
	ErrWaybillAlreadyPrinted = errors.New("waybill already printed")
	ErrWaybillTemplateNotFound = errors.New("waybill template not found")
)

type WaybillStatus int8

const (
	WaybillStatusCreated   WaybillStatus = 1
	WaybillStatusPrinted   WaybillStatus = 2
	WaybillStatusUsed      WaybillStatus = 3
	WaybillStatusCancelled WaybillStatus = 4
	WaybillStatusExpired   WaybillStatus = 5
)

type ElectronicWaybill struct {
	ID             uint64        `json:"id"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	WaybillNo      string        `json:"waybill_no"`
	OrderID        uint64        `json:"order_id"`
	OrderNo        string        `json:"order_no"`
	ProviderCode   string        `json:"provider_code"`
	ProviderName   string        `json:"provider_name"`
	Status         WaybillStatus `json:"status"`
	TemplateID     string        `json:"template_id"`
	TemplateUrl    string        `json:"template_url"`
	PrintUrl       string        `json:"print_url"`
	PdfUrl         string        `json:"pdf_url"`
	PrintCount     int           `json:"print_count"`
	FirstPrintAt   *time.Time    `json:"first_print_at"`
	LastPrintAt    *time.Time    `json:"last_print_at"`
	UsedAt         *time.Time    `json:"used_at"`
	CancelledAt    *time.Time    `json:"cancelled_at"`
	CancelReason   string        `json:"cancel_reason"`
	ExpiredAt      *time.Time    `json:"expired_at"`
	Sender         *WaybillContact `json:"sender"`
	Receiver       *WaybillContact `json:"receiver"`
	Package        *WaybillPackage `json:"package"`
	Remark         string        `json:"remark"`
	CustomFields   map[string]string `json:"custom_fields"`
}

type WaybillTemplate struct {
	ID           uint64        `json:"id"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	TemplateID   string        `json:"template_id"`
	Name         string        `json:"name"`
	ProviderCode string        `json:"provider_code"`
	Width        int           `json:"width"`
	Height       int           `json:"height"`
	IsDefault    bool          `json:"is_default"`
	Enabled      bool          `json:"enabled"`
	Content      string        `json:"content"`
	PreviewUrl   string        `json:"preview_url"`
}

type WaybillPrintRequest struct {
	WaybillNo   string            `json:"waybill_no"`
	TemplateID  string            `json:"template_id"`
	PrintType   PrintType         `json:"print_type"`
	Copies      int               `json:"copies"`
	CustomData  map[string]string `json:"custom_data"`
}

type PrintType string

const (
	PrintTypePDF   PrintType = "pdf"
	PrintTypeImage PrintType = "image"
	PrintTypeZPL   PrintType = "zpl"
	PrintTypeTSPL  PrintType = "tspl"
)

type WaybillPrintResult struct {
	WaybillNo   string    `json:"waybill_no"`
	TemplateID  string    `json:"template_id"`
	PrintType   PrintType `json:"print_type"`
	Content     []byte    `json:"content"`
	Url         string    `json:"url"`
	PrintedAt   time.Time `json:"printed_at"`
}

type WaybillBatchResult struct {
	Total      int                      `json:"total"`
	Success    int                      `json:"success"`
	Failed     int                      `json:"failed"`
	Results    []*WaybillSingleResult   `json:"results"`
}

type WaybillSingleResult struct {
	WaybillNo string `json:"waybill_no"`
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	PrintUrl  string `json:"print_url"`
}

func NewElectronicWaybill(waybillNo string, orderID uint64, orderNo, providerCode, providerName, templateID string) *ElectronicWaybill {
	return &ElectronicWaybill{
		WaybillNo:    waybillNo,
		OrderID:      orderID,
		OrderNo:      orderNo,
		ProviderCode: providerCode,
		ProviderName: providerName,
		Status:       WaybillStatusCreated,
		TemplateID:   templateID,
		PrintCount:   0,
		CustomFields: make(map[string]string),
	}
}

func (w *ElectronicWaybill) SetSender(name, phone, province, city, district, address string) {
	w.Sender = &WaybillContact{
		Name:     name,
		Phone:    phone,
		Province: province,
		City:     city,
		District: district,
		Address:  address,
	}
}

func (w *ElectronicWaybill) SetReceiver(name, phone, province, city, district, address string) {
	w.Receiver = &WaybillContact{
		Name:     name,
		Phone:    phone,
		Province: province,
		City:     city,
		District: district,
		Address:  address,
	}
}

func (w *ElectronicWaybill) SetPackage(weight, volume float64, quantity int) {
	w.Package = &WaybillPackage{
		Weight:   weight,
		Volume:   volume,
		Quantity: quantity,
	}
}

func (w *ElectronicWaybill) MarkPrinted() error {
	if w.Status == WaybillStatusCancelled {
		return ErrWaybillAlreadyUsed
	}

	w.PrintCount++
	now := time.Now()
	if w.FirstPrintAt == nil {
		w.FirstPrintAt = &now
	}
	w.LastPrintAt = &now
	w.Status = WaybillStatusPrinted
	return nil
}

func (w *ElectronicWaybill) MarkUsed() error {
	if w.Status == WaybillStatusUsed {
		return ErrWaybillAlreadyUsed
	}
	if w.Status == WaybillStatusCancelled {
		return ErrWaybillAlreadyUsed
	}

	w.Status = WaybillStatusUsed
	now := time.Now()
	w.UsedAt = &now
	return nil
}

func (w *ElectronicWaybill) Cancel(reason string) error {
	if w.Status == WaybillStatusUsed {
		return ErrWaybillAlreadyUsed
	}

	w.Status = WaybillStatusCancelled
	w.CancelReason = reason
	now := time.Now()
	w.CancelledAt = &now
	return nil
}

func (w *ElectronicWaybill) IsExpired() bool {
	if w.ExpiredAt == nil {
		return false
	}
	return time.Now().After(*w.ExpiredAt)
}

func (w *ElectronicWaybill) CanPrint() bool {
	return w.Status == WaybillStatusCreated || w.Status == WaybillStatusPrinted
}

func (w *ElectronicWaybill) CanCancel() bool {
	return w.Status == WaybillStatusCreated || w.Status == WaybillStatusPrinted
}

func NewWaybillTemplate(templateID, name, providerCode string, width, height int) *WaybillTemplate {
	return &WaybillTemplate{
		TemplateID:   templateID,
		Name:         name,
		ProviderCode: providerCode,
		Width:        width,
		Height:       height,
		IsDefault:    false,
		Enabled:      true,
	}
}

func (t *WaybillTemplate) SetDefault(isDefault bool) {
	t.IsDefault = isDefault
}

func (t *WaybillTemplate) SetContent(content string) {
	t.Content = content
}

func (s WaybillStatus) String() string {
	switch s {
	case WaybillStatusCreated:
		return "CREATED"
	case WaybillStatusPrinted:
		return "PRINTED"
	case WaybillStatusUsed:
		return "USED"
	case WaybillStatusCancelled:
		return "CANCELLED"
	case WaybillStatusExpired:
		return "EXPIRED"
	default:
		return "UNKNOWN"
	}
}

type ElectronicWaybillRepository interface {
	Save(ctx context.Context, waybill *ElectronicWaybill) error
	FindByID(ctx context.Context, id uint64) (*ElectronicWaybill, error)
	FindByWaybillNo(ctx context.Context, waybillNo string) (*ElectronicWaybill, error)
	FindByOrderID(ctx context.Context, orderID uint64) (*ElectronicWaybill, error)
	FindByStatus(ctx context.Context, status WaybillStatus, limit, offset int) ([]*ElectronicWaybill, error)
	FindExpired(ctx context.Context) ([]*ElectronicWaybill, error)
	Update(ctx context.Context, waybill *ElectronicWaybill) error
}

type WaybillTemplateRepository interface {
	Save(ctx context.Context, template *WaybillTemplate) error
	FindByID(ctx context.Context, id uint64) (*WaybillTemplate, error)
	FindByTemplateID(ctx context.Context, templateID string) (*WaybillTemplate, error)
	FindByProviderCode(ctx context.Context, providerCode string) ([]*WaybillTemplate, error)
	FindDefaultByProvider(ctx context.Context, providerCode string) (*WaybillTemplate, error)
	FindAll(ctx context.Context) ([]*WaybillTemplate, error)
	Update(ctx context.Context, template *WaybillTemplate) error
	Delete(ctx context.Context, id uint64) error
}

type ElectronicWaybillService interface {
	Create(ctx context.Context, orderID uint64, providerCode, templateID string) (*ElectronicWaybill, error)
	Print(ctx context.Context, req *WaybillPrintRequest) (*WaybillPrintResult, error)
	BatchPrint(ctx context.Context, waybillNos []string, templateID string) (*WaybillBatchResult, error)
	Cancel(ctx context.Context, waybillNo, reason string) error
	GetPrintUrl(ctx context.Context, waybillNo string) (string, error)
}

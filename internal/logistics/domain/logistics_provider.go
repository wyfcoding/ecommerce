package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrProviderNotFound     = errors.New("logistics provider not found")
	ErrProviderDisabled     = errors.New("logistics provider is disabled")
	ErrTrackingNotFound     = errors.New("tracking info not found")
	ErrWaybillGenerateFail  = errors.New("failed to generate waybill")
)

type ProviderType string

const (
	ProviderTypeSF       ProviderType = "sf"
	ProviderTypeYTO      ProviderType = "yto"
	ProviderTypeZTO      ProviderType = "zto"
	ProviderTypeSTO      ProviderType = "sto"
	ProviderTypeYD       ProviderType = "yd"
	ProviderTypeEMS      ProviderType = "ems"
	ProviderTypeJD       ProviderType = "jd"
	ProviderTypeKuaidi100 ProviderType = "kuaidi100"
	ProviderTypeCainiao  ProviderType = "cainiao"
)

type LogisticsProvider struct {
	ID           uint64            `json:"id"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Code         string            `json:"code"`
	Name         string            `json:"name"`
	Type         ProviderType      `json:"type"`
	Enabled      bool              `json:"enabled"`
	Priority     int               `json:"priority"`
	ApiUrl       string            `json:"api_url"`
	AppKey       string            `json:"app_key"`
	AppSecret    string            `json:"app_secret"`
	CallbackUrl  string            `json:"callback_url"`
	SupportCOD   bool              `json:"support_cod"`
	SupportInsure bool             `json:"support_insure"`
	SupportPickup bool             `json:"support_pickup"`
	Coverage     []string          `json:"coverage"`
	Config       string            `json:"config"`
}

type TrackingInfo struct {
	ID           uint64             `json:"id"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	OrderID      uint64             `json:"order_id"`
	WaybillNo    string             `json:"waybill_no"`
	ProviderCode string             `json:"provider_code"`
	ProviderName string             `json:"provider_name"`
	Status       TrackingStatus     `json:"status"`
	IsSigned     bool               `json:"is_signed"`
	SignTime     *time.Time         `json:"sign_time"`
	SignPerson   string             `json:"sign_person"`
	Traces       []*TrackingTrace   `json:"traces"`
	LastSyncAt   *time.Time         `json:"last_sync_at"`
	EstimatedTime *time.Time        `json:"estimated_time"`
}

type TrackingStatus int8

const (
	TrackingStatusPending    TrackingStatus = 1
	TrackingStatusCollected  TrackingStatus = 2
	TrackingStatusInTransit  TrackingStatus = 3
	TrackingStatusDelivering TrackingStatus = 4
	TrackingStatusSigned     TrackingStatus = 5
	TrackingStatusReturned   TrackingStatus = 6
	TrackingStatusException  TrackingStatus = 7
)

type TrackingTrace struct {
	ID          uint64         `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	TrackingID  uint64         `json:"tracking_id"`
	Time        time.Time      `json:"time"`
	Status      TrackingStatus `json:"status"`
	Location    string         `json:"location"`
	Description string         `json:"description"`
	Operator    string         `json:"operator"`
}

type WaybillRequest struct {
	OrderID       uint64            `json:"order_id"`
	ProviderCode  string            `json:"provider_code"`
	Sender        *WaybillContact   `json:"sender"`
	Receiver      *WaybillContact   `json:"receiver"`
	Package       *WaybillPackage   `json:"package"`
	ServiceType   string            `json:"service_type"`
	IsCOD         bool              `json:"is_cod"`
	CODAmount     int64             `json:"cod_amount"`
	IsInsure      bool              `json:"is_insure"`
	InsureAmount  int64             `json:"insure_amount"`
	Remark        string            `json:"remark"`
	TemplateId    string            `json:"template_id"`
}

type WaybillContact struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Address  string `json:"address"`
	ZipCode  string `json:"zip_code"`
}

type WaybillPackage struct {
	Weight   float64 `json:"weight"`
	Volume   float64 `json:"volume"`
	Length   float64 `json:"length"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Quantity int     `json:"quantity"`
	Items    []string `json:"items"`
}

type WaybillResult struct {
	WaybillNo   string    `json:"waybill_no"`
	OrderID     uint64    `json:"order_id"`
	ProviderCode string   `json:"provider_code"`
	ProviderName string   `json:"provider_name"`
	PrintUrl    string    `json:"print_url"`
	PdfUrl      string    `json:"pdf_url"`
	EstimatedTime *time.Time `json:"estimated_time"`
	Fee         int64     `json:"fee"`
}

type TrackingQueryRequest struct {
	WaybillNo    string `json:"waybill_no"`
	ProviderCode string `json:"provider_code"`
	Phone        string `json:"phone"`
}

type TrackingQueryResult struct {
	WaybillNo    string           `json:"waybill_no"`
	ProviderCode string           `json:"provider_code"`
	ProviderName string           `json:"provider_name"`
	Status       TrackingStatus   `json:"status"`
	IsSigned     bool             `json:"is_signed"`
	SignTime     *time.Time       `json:"sign_time"`
	SignPerson   string           `json:"sign_person"`
	Traces       []*TrackingTrace `json:"traces"`
	EstimatedTime *time.Time      `json:"estimated_time"`
}

type ProviderAdapter interface {
	GetCode() string
	GenerateWaybill(ctx context.Context, req *WaybillRequest) (*WaybillResult, error)
	QueryTracking(ctx context.Context, req *TrackingQueryRequest) (*TrackingQueryResult, error)
	CancelWaybill(ctx context.Context, waybillNo string) error
	GetPrintUrl(ctx context.Context, waybillNo string) (string, error)
	SupportCOD() bool
	SupportInsure() bool
}

func NewLogisticsProvider(code, name string, providerType ProviderType) *LogisticsProvider {
	return &LogisticsProvider{
		Code:     code,
		Name:     name,
		Type:     providerType,
		Enabled:  true,
		Priority: 100,
		Coverage: []string{},
	}
}

func (p *LogisticsProvider) CanUse(province, city string) bool {
	if !p.Enabled {
		return false
	}
	if len(p.Coverage) == 0 {
		return true
	}
	for _, c := range p.Coverage {
		if c == province || c == city || c == "*" {
			return true
		}
	}
	return false
}

func NewTrackingInfo(orderID uint64, waybillNo, providerCode, providerName string) *TrackingInfo {
	return &TrackingInfo{
		OrderID:      orderID,
		WaybillNo:    waybillNo,
		ProviderCode: providerCode,
		ProviderName: providerName,
		Status:       TrackingStatusPending,
		IsSigned:     false,
		Traces:       []*TrackingTrace{},
	}
}

func (t *TrackingInfo) AddTrace(status TrackingStatus, location, description, operator string) {
	trace := &TrackingTrace{
		TrackingID:  t.ID,
		Time:        time.Now(),
		Status:      status,
		Location:    location,
		Description: description,
		Operator:    operator,
	}
	t.Traces = append(t.Traces, trace)
	t.Status = status
	t.UpdatedAt = time.Now()

	if status == TrackingStatusSigned {
		t.IsSigned = true
		now := time.Now()
		t.SignTime = &now
	}
}

func (t *TrackingInfo) UpdateFromQuery(result *TrackingQueryResult) {
	t.Status = result.Status
	t.IsSigned = result.IsSigned
	t.SignTime = result.SignTime
	t.SignPerson = result.SignPerson
	t.EstimatedTime = result.EstimatedTime

	existingTimes := make(map[time.Time]bool)
	for _, trace := range t.Traces {
		existingTimes[trace.Time] = true
	}

	for _, trace := range result.Traces {
		if !existingTimes[trace.Time] {
			t.Traces = append(t.Traces, trace)
		}
	}

	now := time.Now()
	t.LastSyncAt = &now
	t.UpdatedAt = now
}

func (t *TrackingInfo) GetLatestTrace() *TrackingTrace {
	if len(t.Traces) == 0 {
		return nil
	}
	return t.Traces[len(t.Traces)-1]
}

func (s TrackingStatus) String() string {
	switch s {
	case TrackingStatusPending:
		return "PENDING"
	case TrackingStatusCollected:
		return "COLLECTED"
	case TrackingStatusInTransit:
		return "IN_TRANSIT"
	case TrackingStatusDelivering:
		return "DELIVERING"
	case TrackingStatusSigned:
		return "SIGNED"
	case TrackingStatusReturned:
		return "RETURNED"
	case TrackingStatusException:
		return "EXCEPTION"
	default:
		return "UNKNOWN"
	}
}

type LogisticsProviderRepository interface {
	Save(ctx context.Context, provider *LogisticsProvider) error
	FindByID(ctx context.Context, id uint64) (*LogisticsProvider, error)
	FindByCode(ctx context.Context, code string) (*LogisticsProvider, error)
	FindAll(ctx context.Context) ([]*LogisticsProvider, error)
	FindEnabled(ctx context.Context) ([]*LogisticsProvider, error)
	Update(ctx context.Context, provider *LogisticsProvider) error
}

type TrackingInfoRepository interface {
	Save(ctx context.Context, info *TrackingInfo) error
	FindByID(ctx context.Context, id uint64) (*TrackingInfo, error)
	FindByWaybillNo(ctx context.Context, waybillNo string) (*TrackingInfo, error)
	FindByOrderID(ctx context.Context, orderID uint64) (*TrackingInfo, error)
	FindPending(ctx context.Context, limit int) ([]*TrackingInfo, error)
	FindUnsigned(ctx context.Context, limit int) ([]*TrackingInfo, error)
	Update(ctx context.Context, info *TrackingInfo) error
}

type WaybillService interface {
	Generate(ctx context.Context, req *WaybillRequest) (*WaybillResult, error)
	Cancel(ctx context.Context, waybillNo string) error
	QueryTracking(ctx context.Context, waybillNo string) (*TrackingQueryResult, error)
	BatchQueryTracking(ctx context.Context, waybillNos []string) ([]*TrackingQueryResult, error)
}

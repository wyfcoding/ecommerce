// 变更说明：完善跨境电商领域模型，增加清关流程、税费计算等完整功能
package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// DeclarationStatus 报关状态
type DeclarationStatus int8

const (
	DeclarationStatusDraft             DeclarationStatus = 1  // 草稿
	DeclarationStatusPending           DeclarationStatus = 2  // 待提交
	DeclarationStatusSubmitted         DeclarationStatus = 3  // 已提交
	DeclarationStatusCustomsProcessing DeclarationStatus = 4  // 海关处理中
	DeclarationStatusCustomsCleared    DeclarationStatus = 5  // 海关放行
	DeclarationStatusCustomsInspection DeclarationStatus = 6  // 海关查验
	DeclarationStatusCustomsRejected   DeclarationStatus = 7  // 海关拒绝
	DeclarationStatusCleared           DeclarationStatus = 8  // 已清关
	DeclarationStatusCancelled         DeclarationStatus = 9  // 已取消
)

func (s DeclarationStatus) String() string {
	switch s {
	case DeclarationStatusDraft:
		return "DRAFT"
	case DeclarationStatusPending:
		return "PENDING"
	case DeclarationStatusSubmitted:
		return "SUBMITTED"
	case DeclarationStatusCustomsProcessing:
		return "CUSTOMS_PROCESSING"
	case DeclarationStatusCustomsCleared:
		return "CUSTOMS_CLEARED"
	case DeclarationStatusCustomsInspection:
		return "CUSTOMS_INSPECTION"
	case DeclarationStatusCustomsRejected:
		return "CUSTOMS_REJECTED"
	case DeclarationStatusCleared:
		return "CLEARED"
	case DeclarationStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

// ClearanceStatus 清关状态
type ClearanceStatus int8

const (
	ClearanceStatusPending    ClearanceStatus = 1 // 待清关
	ClearanceStatusProcessing ClearanceStatus = 2 // 清关中
	ClearanceStatusInspection ClearanceStatus = 3 // 查验中
	ClearanceStatusTaxPending ClearanceStatus = 4 // 待缴税
	ClearanceStatusTaxPaid    ClearanceStatus = 5 // 已缴税
	ClearanceStatusReleased   ClearanceStatus = 6 // 已放行
	ClearanceStatusCompleted  ClearanceStatus = 7 // 已完成
	ClearanceStatusFailed     ClearanceStatus = 8 // 清关失败
)

func (s ClearanceStatus) String() string {
	switch s {
	case ClearanceStatusPending:
		return "PENDING"
	case ClearanceStatusProcessing:
		return "PROCESSING"
	case ClearanceStatusInspection:
		return "INSPECTION"
	case ClearanceStatusTaxPending:
		return "TAX_PENDING"
	case ClearanceStatusTaxPaid:
		return "TAX_PAID"
	case ClearanceStatusReleased:
		return "RELEASED"
	case ClearanceStatusCompleted:
		return "COMPLETED"
	case ClearanceStatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// TradeMode 贸易方式
type TradeMode int8

const (
	TradeModeBonded  TradeMode = 1 // 保税
	TradeModeDirect  TradeMode = 2 // 直邮
	TradeModeGeneral TradeMode = 3 // 一般贸易
)

func (m TradeMode) String() string {
	switch m {
	case TradeModeBonded:
		return "BONDED"
	case TradeModeDirect:
		return "DIRECT"
	case TradeModeGeneral:
		return "GENERAL"
	default:
		return "UNKNOWN"
	}
}

// CustomsDeclaration 报关单聚合根
type CustomsDeclaration struct {
	gorm.Model
	DeclarationID    string            `gorm:"column:declaration_id;type:varchar(32);uniqueIndex;not null" json:"declaration_id"`
	OrderID          string            `gorm:"column:order_id;type:varchar(32);index;not null" json:"order_id"`
	UserID           uint64            `gorm:"column:user_id;index;not null" json:"user_id"`
	MerchantID       uint64            `gorm:"column:merchant_id;index" json:"merchant_id"`
	LogisticsNo      string            `gorm:"column:logistics_no;type:varchar(64)" json:"logistics_no"`
	DeclaredValue    decimal.Decimal   `gorm:"column:declared_value;type:decimal(20,2);not null" json:"declared_value"`
	Currency         string            `gorm:"column:currency;type:varchar(3);not null" json:"currency"`
	DutyAmount       decimal.Decimal   `gorm:"column:duty_amount;type:decimal(20,2)" json:"duty_amount"`
	TaxAmount        decimal.Decimal   `gorm:"column:tax_amount;type:decimal(20,2)" json:"tax_amount"`
	Status           DeclarationStatus `gorm:"column:status;type:tinyint;not null;default:1" json:"status"`
	RejectReason     string            `gorm:"column:reject_reason;type:varchar(512)" json:"reject_reason"`
	CustomsPort      string            `gorm:"column:customs_port;type:varchar(32)" json:"customs_port"`
	TradeMode        TradeMode         `gorm:"column:trade_mode;type:tinyint" json:"trade_mode"`
	CustomsDeclNo    string            `gorm:"column:customs_decl_no;type:varchar(32)" json:"customs_decl_no"`
	OriginCountry    string            `gorm:"column:origin_country;type:varchar(32)" json:"origin_country"`
	DestinationCountry string          `gorm:"column:destination_country;type:varchar(32)" json:"destination_country"`
	
	// 收件人信息
	RecipientName    string            `gorm:"column:recipient_name;type:varchar(64)" json:"recipient_name"`
	RecipientPhone   string            `gorm:"column:recipient_phone;type:varchar(32)" json:"recipient_phone"`
	RecipientAddress string            `gorm:"column:recipient_address;type:varchar(255)" json:"recipient_address"`
	RecipientIDNumber string           `gorm:"column:recipient_id_number;type:varchar(32)" json:"recipient_id_number"`
	
	// 清关信息
	ClearanceID      string            `gorm:"column:clearance_id;type:varchar(32)" json:"clearance_id"`
	ClearanceStatus  ClearanceStatus   `gorm:"column:clearance_status;type:tinyint" json:"clearance_status"`
	CustomsCode      string            `gorm:"column:customs_code;type:varchar(32)" json:"customs_code"`
	CustomsName      string            `gorm:"column:customs_name;type:varchar(128)" json:"customs_name"`
	ReleaseOrderNo   string            `gorm:"column:release_order_no;type:varchar(32)" json:"release_order_no"`
	
	SubmittedAt      *time.Time        `gorm:"column:submitted_at" json:"submitted_at"`
	ClearedAt        *time.Time        `gorm:"column:cleared_at" json:"cleared_at"`
	
	Items            []DeclarationItem `gorm:"foreignKey:DeclarationID;references:DeclarationID" json:"items"`
	Documents        []CustomsDocument `gorm:"foreignKey:DeclarationID;references:DeclarationID" json:"documents"`
	ClearanceEvents  []ClearanceEvent  `gorm:"foreignKey:DeclarationID;references:DeclarationID" json:"clearance_events"`
	
	domainEvents     []DomainEvent     `gorm:"-" json:"-"`
}

// TableName 表名
func (CustomsDeclaration) TableName() string {
	return "customs_declarations"
}

// DeclarationItem 报关明细
type DeclarationItem struct {
	gorm.Model
	DeclarationID string          `gorm:"column:declaration_id;type:varchar(32);index;not null" json:"declaration_id"`
	SKUID         string          `gorm:"column:sku_id;type:varchar(64);not null" json:"sku_id"`
	ProductName   string          `gorm:"column:product_name;type:varchar(255)" json:"product_name"`
	HSCode        string          `gorm:"column:hs_code;type:varchar(20)" json:"hs_code"`
	Price         decimal.Decimal `gorm:"column:price;type:decimal(20,2);not null" json:"price"`
	Quantity      int32           `gorm:"column:quantity;not null" json:"quantity"`
	Weight        decimal.Decimal `gorm:"column:weight;type:decimal(10,3)" json:"weight"`
	DutyRate      decimal.Decimal `gorm:"column:duty_rate;type:decimal(5,4)" json:"duty_rate"`
	TaxRate       decimal.Decimal `gorm:"column:tax_rate;type:decimal(5,4)" json:"tax_rate"`
	DutyAmount    decimal.Decimal `gorm:"column:duty_amount;type:decimal(20,2)" json:"duty_amount"`
	TaxAmount     decimal.Decimal `gorm:"column:tax_amount;type:decimal(20,2)" json:"tax_amount"`
}

// TableName 表名
func (DeclarationItem) TableName() string {
	return "declaration_items"
}

// CustomsDocument 报关证件
type CustomsDocument struct {
	gorm.Model
	DeclarationID string             `gorm:"column:declaration_id;type:varchar(32);index;not null" json:"declaration_id"`
	DocumentID    string             `gorm:"column:document_id;type:varchar(32);uniqueIndex;not null" json:"document_id"`
	DocumentType  CustomsDocumentType `gorm:"column:document_type;type:tinyint;not null" json:"document_type"`
	DocumentName  string             `gorm:"column:document_name;type:varchar(128)" json:"document_name"`
	DocumentURL   string             `gorm:"column:document_url;type:varchar(512)" json:"document_url"`
	Status        string             `gorm:"column:status;type:varchar(32)" json:"status"`
	UploadedAt    time.Time          `gorm:"column:uploaded_at" json:"uploaded_at"`
}

// CustomsDocumentType 证件类型
type CustomsDocumentType int8

const (
	CustomsDocumentTypeInvoice     CustomsDocumentType = 1 // 发票
	CustomsDocumentTypePackingList CustomsDocumentType = 2 // 装箱单
	CustomsDocumentTypeContract    CustomsDocumentType = 3 // 合同
	CustomsDocumentTypeLicense     CustomsDocumentType = 4 // 许可证
	CustomsDocumentTypeCertificate CustomsDocumentType = 5 // 证书
	CustomsDocumentTypeIDCard      CustomsDocumentType = 6 // 身份证
	CustomsDocumentTypeOther       CustomsDocumentType = 7 // 其他
)

// TableName 表名
func (CustomsDocument) TableName() string {
	return "customs_documents"
}

// ClearanceEvent 清关事件
type ClearanceEvent struct {
	gorm.Model
	DeclarationID string    `gorm:"column:declaration_id;type:varchar(32);index;not null" json:"declaration_id"`
	EventType     string    `gorm:"column:event_type;type:varchar(32);not null" json:"event_type"`
	Description   string    `gorm:"column:description;type:varchar(255)" json:"description"`
	Location      string    `gorm:"column:location;type:varchar(64)" json:"location"`
	OccurredAt    time.Time `gorm:"column:occurred_at" json:"occurred_at"`
}

// TableName 表名
func (ClearanceEvent) TableName() string {
	return "clearance_events"
}

// HSCode 海关编码
type HSCode struct {
	gorm.Model
	Code              string          `gorm:"column:code;type:varchar(20);uniqueIndex;not null" json:"code"`
	Description       string          `gorm:"column:description;type:varchar(255)" json:"description"`
	DescriptionEn     string          `gorm:"column:description_en;type:varchar(255)" json:"description_en"`
	DutyRate          decimal.Decimal `gorm:"column:duty_rate;type:decimal(5,4);not null" json:"duty_rate"`
	VATRate           decimal.Decimal `gorm:"column:vat_rate;type:decimal(5,4);not null" json:"vat_rate"`
	ConsumptionTaxRate decimal.Decimal `gorm:"column:consumption_tax_rate;type:decimal(5,4)" json:"consumption_tax_rate"`
	Unit              string          `gorm:"column:unit;type:varchar(16)" json:"unit"`
	Restrictions      string          `gorm:"column:restrictions;type:text" json:"restrictions"`
	RequiredDocuments string          `gorm:"column:required_documents;type:text" json:"required_documents"`
	Active            bool            `gorm:"column:active;not null;default:true" json:"active"`
}

// TableName 表名
func (HSCode) TableName() string {
	return "hs_codes"
}

// CrossBorderOrder 跨境订单
type CrossBorderOrder struct {
	gorm.Model
	CrossBorderOrderID string          `gorm:"column:cross_border_order_id;type:varchar(32);uniqueIndex;not null" json:"cross_border_order_id"`
	OrderID            string          `gorm:"column:order_id;type:varchar(32);index;not null" json:"order_id"`
	UserID             uint64          `gorm:"column:user_id;index;not null" json:"user_id"`
	OriginCountry      string          `gorm:"column:origin_country;type:varchar(32)" json:"origin_country"`
	DestinationCountry string          `gorm:"column:destination_country;type:varchar(32)" json:"destination_country"`
	CustomsPort        string          `gorm:"column:customs_port;type:varchar(32)" json:"customs_port"`
	TradeMode          TradeMode       `gorm:"column:trade_mode;type:tinyint" json:"trade_mode"`
	DeclarationID      string          `gorm:"column:declaration_id;type:varchar(32)" json:"declaration_id"`
	TrackingNo         string          `gorm:"column:tracking_no;type:varchar(64)" json:"tracking_no"`
	LogisticsCompany   string          `gorm:"column:logistics_company;type:varchar(64)" json:"logistics_company"`
	Status             string          `gorm:"column:status;type:varchar(32)" json:"status"`
	TotalDuty          decimal.Decimal `gorm:"column:total_duty;type:decimal(20,2)" json:"total_duty"`
	TotalTax           decimal.Decimal `gorm:"column:total_tax;type:decimal(20,2)" json:"total_tax"`
	Currency           string          `gorm:"column:currency;type:varchar(3)" json:"currency"`
}

// TableName 表名
func (CrossBorderOrder) TableName() string {
	return "cross_border_orders"
}

// NewDeclaration 创建报关单
func NewDeclaration(orderID string, userID uint64, currency string, declaredValue decimal.Decimal) *CustomsDeclaration {
	now := time.Now()
	return &CustomsDeclaration{
		DeclarationID:    fmt.Sprintf("CB%d%d", now.UnixNano(), userID),
		OrderID:         orderID,
		UserID:          userID,
		Currency:        currency,
		DeclaredValue:   declaredValue,
		Status:          DeclarationStatusDraft,
		DutyAmount:      decimal.Zero,
		TaxAmount:       decimal.Zero,
		Items:           []DeclarationItem{},
		Documents:       []CustomsDocument{},
		ClearanceEvents: []ClearanceEvent{},
		domainEvents:    []DomainEvent{},
	}
}

// AddItem 添加报关明细
func (d *CustomsDeclaration) AddItem(skuID, productName, hsCode string, price decimal.Decimal, qty int32, weight decimal.Decimal) {
	d.Items = append(d.Items, DeclarationItem{
		DeclarationID: d.DeclarationID,
		SKUID:        skuID,
		ProductName:  productName,
		HSCode:       hsCode,
		Price:        price,
		Quantity:     qty,
		Weight:       weight,
	})
}

// CalculateTax 计算税费
func (d *CustomsDeclaration) CalculateTax(hsCodes map[string]*HSCode) error {
	totalDuty := decimal.Zero
	totalTax := decimal.Zero

	for i := range d.Items {
		hsCode, ok := hsCodes[d.Items[i].HSCode]
		if !ok {
			continue
		}

		itemValue := d.Items[i].Price.Mul(decimal.NewFromInt(int64(d.Items[i].Quantity)))
		dutyAmount := itemValue.Mul(hsCode.DutyRate)
		vatAmount := itemValue.Add(dutyAmount).Mul(hsCode.VATRate)

		d.Items[i].DutyRate = hsCode.DutyRate
		d.Items[i].TaxRate = hsCode.VATRate
		d.Items[i].DutyAmount = dutyAmount
		d.Items[i].TaxAmount = vatAmount

		totalDuty = totalDuty.Add(dutyAmount)
		totalTax = totalTax.Add(vatAmount)
	}

	d.DutyAmount = totalDuty
	d.TaxAmount = totalTax
	return nil
}

// Submit 提交报关
func (d *CustomsDeclaration) Submit() error {
	if d.Status != DeclarationStatusDraft && d.Status != DeclarationStatusPending {
		return errors.New("invalid status for submit")
	}
	d.Status = DeclarationStatusSubmitted
	now := time.Now()
	d.SubmittedAt = &now
	d.addEvent(&DeclarationSubmittedEvent{
		DeclarationID: d.DeclarationID,
		OrderID:       d.OrderID,
		Timestamp:     now,
	})
	return nil
}

// StartCustomsProcessing 开始海关处理
func (d *CustomsDeclaration) StartCustomsProcessing(customsCode, customsName string) error {
	if d.Status != DeclarationStatusSubmitted {
		return errors.New("invalid status for customs processing")
	}
	d.Status = DeclarationStatusCustomsProcessing
	d.CustomsCode = customsCode
	d.CustomsName = customsName
	d.addClearanceEvent("CUSTOMS_RECEIVED", "报关单已送达海关", customsName)
	return nil
}

// CustomsClear 海关放行
func (d *CustomsDeclaration) CustomsClear() error {
	if d.Status != DeclarationStatusCustomsProcessing {
		return errors.New("invalid status for customs clear")
	}
	d.Status = DeclarationStatusCustomsCleared
	d.addClearanceEvent("CUSTOMS_CLEARED", "海关放行", d.CustomsName)
	return nil
}

// CustomsInspection 海关查验
func (d *CustomsDeclaration) CustomsInspection() error {
	d.Status = DeclarationStatusCustomsInspection
	d.addClearanceEvent("CUSTOMS_INSPECTION", "海关查验中", d.CustomsName)
	return nil
}

// CustomsReject 海关拒绝
func (d *CustomsDeclaration) CustomsReject(reason string) error {
	d.Status = DeclarationStatusCustomsRejected
	d.RejectReason = reason
	d.addClearanceEvent("CUSTOMS_REJECTED", reason, d.CustomsName)
	return nil
}

// StartClearance 开始清关
func (d *CustomsDeclaration) StartClearance(clearanceID string) error {
	d.ClearanceID = clearanceID
	d.ClearanceStatus = ClearanceStatusProcessing
	d.addClearanceEvent("CLEARANCE_STARTED", "开始清关", "")
	return nil
}

// CompleteClearance 完成清关
func (d *CustomsDeclaration) CompleteClearance(releaseOrderNo string) error {
	d.Status = DeclarationStatusCleared
	d.ClearanceStatus = ClearanceStatusCompleted
	d.ReleaseOrderNo = releaseOrderNo
	now := time.Now()
	d.ClearedAt = &now
	d.addClearanceEvent("CLEARANCE_COMPLETED", "清关完成", "")
	d.addEvent(&DeclarationClearedEvent{
		DeclarationID: d.DeclarationID,
		OrderID:       d.OrderID,
		Timestamp:     now,
	})
	return nil
}

// Cancel 取消报关
func (d *CustomsDeclaration) Cancel(reason string) error {
	if d.Status == DeclarationStatusCleared {
		return errors.New("cannot cancel cleared declaration")
	}
	d.Status = DeclarationStatusCancelled
	d.RejectReason = reason
	return nil
}

// AddDocument 添加证件
func (d *CustomsDeclaration) AddDocument(docType CustomsDocumentType, name, url string) {
	d.Documents = append(d.Documents, CustomsDocument{
		DeclarationID: d.DeclarationID,
		DocumentID:    fmt.Sprintf("DOC%d", time.Now().UnixNano()),
		DocumentType:  docType,
		DocumentName:  name,
		DocumentURL:   url,
		Status:        "UPLOADED",
		UploadedAt:    time.Now(),
	})
}

// addClearanceEvent 添加清关事件
func (d *CustomsDeclaration) addClearanceEvent(eventType, description, location string) {
	d.ClearanceEvents = append(d.ClearanceEvents, ClearanceEvent{
		DeclarationID: d.DeclarationID,
		EventType:     eventType,
		Description:   description,
		Location:      location,
		OccurredAt:    time.Now(),
	})
}

// addEvent 添加领域事件
func (d *CustomsDeclaration) addEvent(event DomainEvent) {
	d.domainEvents = append(d.domainEvents, event)
}

// GetDomainEvents 获取领域事件
func (d *CustomsDeclaration) GetDomainEvents() []DomainEvent {
	return d.domainEvents
}

// ClearDomainEvents 清除领域事件
func (d *CustomsDeclaration) ClearDomainEvents() {
	d.domainEvents = nil
}

// DomainEvent 领域事件接口
type DomainEvent interface {
	EventName() string
}

// DeclarationSubmittedEvent 报关提交事件
type DeclarationSubmittedEvent struct {
	DeclarationID string    `json:"declaration_id"`
	OrderID       string    `json:"order_id"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *DeclarationSubmittedEvent) EventName() string { return "crossborder.declaration.submitted" }

// DeclarationClearedEvent 报关完成事件
type DeclarationClearedEvent struct {
	DeclarationID string    `json:"declaration_id"`
	OrderID       string    `json:"order_id"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *DeclarationClearedEvent) EventName() string { return "crossborder.declaration.cleared" }

// TaxCalculatedEvent 税费计算事件
type TaxCalculatedEvent struct {
	DeclarationID string          `json:"declaration_id"`
	DutyAmount    decimal.Decimal `json:"duty_amount"`
	TaxAmount     decimal.Decimal `json:"tax_amount"`
	Timestamp     time.Time       `json:"timestamp"`
}

func (e *TaxCalculatedEvent) EventName() string { return "crossborder.tax.calculated" }

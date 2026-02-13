// Package domain 发票服务领域层
// 生成摘要：
// 1) 定义发票聚合根和状态机
// 2) 定义发票抬头、明细等值对象
package domain

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// InvoiceStatus 发票状态
type InvoiceStatus int8

const (
	InvoiceStatusPending    InvoiceStatus = 1 // 待开票
	InvoiceStatusIssued     InvoiceStatus = 2 // 已开票
	InvoiceStatusFailed     InvoiceStatus = 3 // 开票失败
	InvoiceStatusRedPending InvoiceStatus = 4 // 待红冲
	InvoiceStatusRedIssued  InvoiceStatus = 5 // 已红冲
	InvoiceStatusCancelled  InvoiceStatus = 6 // 已取消
)

func (s InvoiceStatus) String() string {
	switch s {
	case InvoiceStatusPending:
		return "pending"
	case InvoiceStatusIssued:
		return "issued"
	case InvoiceStatusFailed:
		return "failed"
	case InvoiceStatusRedPending:
		return "red_pending"
	case InvoiceStatusRedIssued:
		return "red_issued"
	case InvoiceStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// InvoiceType 发票类型
type InvoiceType int8

const (
	InvoiceTypePersonalNormal InvoiceType = 1 // 个人普通发票
	InvoiceTypeCompanyNormal  InvoiceType = 2 // 企业普通发票
	InvoiceTypeCompanySpecial InvoiceType = 3 // 企业专用发票
)

// InvoiceMedium 发票介质
type InvoiceMedium int8

const (
	InvoiceMediumElectronic InvoiceMedium = 1 // 电子发票
	InvoiceMediumPaper      InvoiceMedium = 2 // 纸质发票
)

// Invoice 发票聚合根
type Invoice struct {
	gorm.Model
	// ApplicationNo 申请单号
	ApplicationNo string `gorm:"column:application_no;type:varchar(32);unique_index;not null" json:"application_no"`
	// InvoiceCode 发票代码
	InvoiceCode string `gorm:"column:invoice_code;type:varchar(32)" json:"invoice_code"`
	// InvoiceNo 发票号码
	InvoiceNo string `gorm:"column:invoice_no;type:varchar(32);index" json:"invoice_no"`
	// CheckCode 校验码
	CheckCode string `gorm:"column:check_code;type:varchar(32)" json:"check_code"`
	// OrderNo 订单号
	OrderNo string `gorm:"column:order_no;type:varchar(32);index;not null" json:"order_no"`
	// UserID 用户ID
	UserID uint64 `gorm:"column:user_id;index;not null" json:"user_id"`
	// MerchantID 商家ID
	MerchantID uint64 `gorm:"column:merchant_id;index;not null" json:"merchant_id"`

	Type   InvoiceType   `gorm:"column:type;type:tinyint;not null" json:"type"`
	Medium InvoiceMedium `gorm:"column:medium;type:tinyint;not null" json:"medium"`
	Status InvoiceStatus `gorm:"column:status;type:tinyint;not null;default:1" json:"status"`

	// 金额信息
	Amount    int64  `gorm:"column:amount;not null" json:"amount"`             // 含税金额
	TaxAmount int64  `gorm:"column:tax_amount;not null" json:"tax_amount"`     // 税额
	TaxRate   string `gorm:"column:tax_rate;type:varchar(10)" json:"tax_rate"` // 税率

	// 抬头信息 (值对象展开)
	TitleName     string `gorm:"column:title_name;type:varchar(128)" json:"title_name"`
	TitleTaxID    string `gorm:"column:title_tax_id;type:varchar(32)" json:"title_tax_id"`
	TitleBank     string `gorm:"column:title_bank;type:varchar(128)" json:"title_bank"`
	TitleAccount  string `gorm:"column:title_account;type:varchar(64)" json:"title_account"`
	TitleAddress  string `gorm:"column:title_address;type:varchar(255)" json:"title_address"`
	TitlePhone    string `gorm:"column:title_phone;type:varchar(32)" json:"title_phone"`
	ReceiverEmail string `gorm:"column:receiver_email;type:varchar(128)" json:"receiver_email"`
	ReceiverPhone string `gorm:"column:receiver_phone;type:varchar(32)" json:"receiver_phone"`

	// 文件信息
	PDFUrl string `gorm:"column:pdf_url;type:varchar(512)" json:"pdf_url"`
	XMLUrl string `gorm:"column:xml_url;type:varchar(512)" json:"xml_url"`

	// 关联
	RelatedInvoiceID uint64 `gorm:"column:related_invoice_id;index" json:"related_invoice_id"` // 关联的蓝票ID（红票用）
	IsRed            bool   `gorm:"column:is_red;not null;default:false" json:"is_red"`        // 是否红票
	RedReason        string `gorm:"column:red_reason;type:varchar(255)" json:"red_reason"`     // 红冲原因

	IssuedAt *time.Time `gorm:"column:issued_at" json:"issued_at"`
	Remark   string     `gorm:"column:remark;type:text" json:"remark"`

	// 明细
	Items []InvoiceItem `gorm:"foreignKey:InvoiceID" json:"items"`

	// 领域事件
	domainEvents []DomainEvent `gorm:"-" json:"-"`
}

// TableName 表名
func (Invoice) TableName() string {
	return "invoices"
}

// InvoiceItem 发票明细
type InvoiceItem struct {
	gorm.Model
	InvoiceID   uint   `gorm:"column:invoice_id;index;not null" json:"invoice_id"`
	ProductName string `gorm:"column:product_name;type:varchar(255);not null" json:"product_name"`
	Spec        string `gorm:"column:spec;type:varchar(128)" json:"spec"`
	Unit        string `gorm:"column:unit;type:varchar(32)" json:"unit"`
	Quantity    int32  `gorm:"column:quantity;not null" json:"quantity"`
	Price       int64  `gorm:"column:price;not null" json:"price"`   // 单价(分)
	Amount      int64  `gorm:"column:amount;not null" json:"amount"` // 金额(分)
	TaxRate     string `gorm:"column:tax_rate;type:varchar(10)" json:"tax_rate"`
	TaxAmount   int64  `gorm:"column:tax_amount;not null" json:"tax_amount"`
}

// TableName 表名
func (InvoiceItem) TableName() string {
	return "invoice_items"
}

// NewInvoice 创建发票申请
func NewInvoice(orderNo string, userID, merchantID uint64, amount int64, invType InvoiceType, medium InvoiceMedium) *Invoice {
	now := time.Now()
	appNo := fmt.Sprintf("INV%s%04d", now.Format("20060102150405"), now.UnixNano()%10000)

	inv := &Invoice{
		ApplicationNo: appNo,
		OrderNo:       orderNo,
		UserID:        userID,
		MerchantID:    merchantID,
		Amount:        amount,
		Type:          invType,
		Medium:        medium,
		Status:        InvoiceStatusPending,
		domainEvents:  make([]DomainEvent, 0),
	}

	inv.addEvent(&InvoiceAppliedEvent{
		InvoiceID:     0, // 创建后设置
		ApplicationNo: appNo,
		OrderNo:       orderNo,
		UserID:        userID,
		Amount:        amount,
		Timestamp:     now,
	})

	return inv
}

// SetTitle 设置抬头
func (i *Invoice) SetTitle(name, taxID, bank, account, address, phone, email, receiverPhone string) {
	i.TitleName = name
	i.TitleTaxID = taxID
	i.TitleBank = bank
	i.TitleAccount = account
	i.TitleAddress = address
	i.TitlePhone = phone
	i.ReceiverEmail = email
	i.ReceiverPhone = receiverPhone
}

// AddItem 添加明细
func (i *Invoice) AddItem(name, spec, unit string, qty int32, price, amount int64, taxRate string, taxAmt int64) {
	i.Items = append(i.Items, InvoiceItem{
		ProductName: name,
		Spec:        spec,
		Unit:        unit,
		Quantity:    qty,
		Price:       price,
		Amount:      amount,
		TaxRate:     taxRate,
		TaxAmount:   taxAmt,
	})
	i.TaxAmount += taxAmt
}

// Issue 开具发票
func (i *Invoice) Issue(code, no, checkCode, pdfUrl, xmlUrl string) error {
	if i.Status != InvoiceStatusPending && i.Status != InvoiceStatusRedPending {
		return errors.New("invalid status for issue")
	}

	now := time.Now()
	i.InvoiceCode = code
	i.InvoiceNo = no
	i.CheckCode = checkCode
	i.PDFUrl = pdfUrl
	i.XMLUrl = xmlUrl
	i.IssuedAt = &now

	if i.IsRed {
		i.Status = InvoiceStatusRedIssued
	} else {
		i.Status = InvoiceStatusIssued
	}

	i.addEvent(&InvoiceIssuedEvent{
		InvoiceID:   uint64(i.ID),
		InvoiceCode: code,
		InvoiceNo:   no,
		PDFUrl:      pdfUrl,
		IsRed:       i.IsRed,
		Timestamp:   now,
	})

	return nil
}

// Fail 开票失败
func (i *Invoice) Fail(reason string) error {
	if i.Status != InvoiceStatusPending && i.Status != InvoiceStatusRedPending {
		return errors.New("invalid status for fail")
	}

	now := time.Now()
	i.Status = InvoiceStatusFailed
	i.Remark = reason

	i.addEvent(&InvoiceFailedEvent{
		InvoiceID: uint64(i.ID),
		Reason:    reason,
		Timestamp: now,
	})

	return nil
}

// ApplyRed 申请红冲
func (i *Invoice) ApplyRed(reason string) (*Invoice, error) {
	if i.Status != InvoiceStatusIssued {
		return nil, errors.New("only issued invoice can be red flushed")
	}

	redInv := NewInvoice(i.OrderNo, i.UserID, i.MerchantID, -i.Amount, i.Type, i.Medium)
	redInv.IsRed = true
	redInv.RelatedInvoiceID = uint64(i.ID)
	redInv.RedReason = reason
	redInv.Status = InvoiceStatusRedPending

	redInv.SetTitle(i.TitleName, i.TitleTaxID, i.TitleBank, i.TitleAccount, i.TitleAddress, i.TitlePhone, i.ReceiverEmail, i.ReceiverPhone)

	for _, item := range i.Items {
		redInv.AddItem(item.ProductName, item.Spec, item.Unit, item.Quantity, item.Price, -item.Amount, item.TaxRate, -item.TaxAmount)
	}

	return redInv, nil
}

// ApplyBlue 申请蓝冲（重新开具）
func (i *Invoice) ApplyBlue(reason string, newTitle InvoiceTitle) (*Invoice, error) {
	if i.Status != InvoiceStatusRedIssued {
		return nil, errors.New("only red-issued invoice can be blue reissued")
	}

	blueInv := NewInvoice(i.OrderNo, i.UserID, i.MerchantID, -i.Amount, i.Type, i.Medium)
	blueInv.IsRed = false
	blueInv.RelatedInvoiceID = i.RelatedInvoiceID
	blueInv.RedReason = reason
	blueInv.Status = InvoiceStatusPending

	blueInv.SetTitle(newTitle.Name, newTitle.TaxID, newTitle.Bank, newTitle.Account, newTitle.Address, newTitle.Phone, newTitle.Email, newTitle.ReceiverPhone)

	for _, item := range i.Items {
		blueInv.AddItem(item.ProductName, item.Spec, item.Unit, item.Quantity, item.Price, -item.Amount, item.TaxRate, -item.TaxAmount)
	}

	blueInv.addEvent(&InvoiceBlueAppliedEvent{
		InvoiceID:     uint64(i.ID),
		ApplicationNo: blueInv.ApplicationNo,
		Reason:        reason,
		Timestamp:     time.Now(),
	})

	return blueInv, nil
}

// Cancel 取消发票
func (i *Invoice) Cancel(reason string) error {
	if i.Status == InvoiceStatusIssued || i.Status == InvoiceStatusRedIssued {
		return errors.New("issued invoice cannot be cancelled directly, use red flush instead")
	}
	
	i.Status = InvoiceStatusCancelled
	i.Remark = reason
	
	i.addEvent(&InvoiceCancelledEvent{
		InvoiceID:     uint64(i.ID),
		ApplicationNo: i.ApplicationNo,
		Reason:        reason,
		Timestamp:     time.Now(),
	})
	
	return nil
}

// CanRedFlush 检查是否可以红冲
func (i *Invoice) CanRedFlush() bool {
	return i.Status == InvoiceStatusIssued && !i.IsRed
}

// CanBlueReissue 检查是否可以蓝冲
func (i *Invoice) CanBlueReissue() bool {
	return i.Status == InvoiceStatusRedIssued && i.IsRed
}

// GetOriginalInvoiceID 获取原始蓝票ID
func (i *Invoice) GetOriginalInvoiceID() uint64 {
	if i.IsRed {
		return i.RelatedInvoiceID
	}
	return uint64(i.ID)
}

// InvoiceTitle 发票抬头值对象
type InvoiceTitle struct {
	Name          string
	TaxID         string
	Bank          string
	Account       string
	Address       string
	Phone         string
	Email         string
	ReceiverPhone string
}

// InvoiceVerification 发票验真结果
type InvoiceVerification struct {
	InvoiceCode     string
	InvoiceNo       string
	Valid           bool
	VerifyTime      time.Time
	InvoiceStatus   string
	SellerName      string
	SellerTaxID     string
	BuyerName       string
	BuyerTaxID      string
	Amount          int64
	TaxAmount       int64
	IssueDate       string
	InvalidationMark string
}

// addEvent 添加领域事件
func (i *Invoice) addEvent(event DomainEvent) {
	i.domainEvents = append(i.domainEvents, event)
}

// GetDomainEvents 获取领域事件
func (i *Invoice) GetDomainEvents() []DomainEvent {
	return i.domainEvents
}

// ClearDomainEvents 清除领域事件
func (i *Invoice) ClearDomainEvents() {
	i.domainEvents = nil
}

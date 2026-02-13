// Package domain 发票服务仓储接口
package domain

import "context"

// InvoiceRepository 发票仓储接口
type InvoiceRepository interface {
	Save(ctx context.Context, inv *Invoice) error
	Update(ctx context.Context, inv *Invoice) error
	FindByID(ctx context.Context, id uint) (*Invoice, error)
	FindByApplicationNo(ctx context.Context, no string) (*Invoice, error)
	FindByOrderNo(ctx context.Context, orderNo string) ([]*Invoice, error)
	List(ctx context.Context, filter *InvoiceFilter) ([]*Invoice, int64, error)
}

// InvoiceFilter 发票过滤条件
type InvoiceFilter struct {
	UserID     uint64
	MerchantID uint64
	OrderNo    string
	Status     *InvoiceStatus
	Page       int
	PageSize   int
}

// InvoicePlatformService 发票平台服务接口
type InvoicePlatformService interface {
	// IssueInvoice 开具发票
	IssueInvoice(ctx context.Context, req *IssueInvoiceRequest) (*IssueInvoiceResult, error)
	
	// RedInvoice 红冲发票
	RedInvoice(ctx context.Context, req *RedInvoiceRequest) (*RedInvoiceResult, error)
	
	// VerifyInvoice 验真发票
	VerifyInvoice(ctx context.Context, req *VerifyInvoiceRequest) (*InvoiceVerification, error)
	
	// QueryInvoice 查询发票
	QueryInvoice(ctx context.Context, req *QueryInvoiceRequest) (*QueryInvoiceResult, error)
	
	// DownloadInvoice 下载发票文件
	DownloadInvoice(ctx context.Context, req *DownloadInvoiceRequest) (*DownloadInvoiceResult, error)
}

// IssueInvoiceRequest 开具发票请求
type IssueInvoiceRequest struct {
	OrderNo       string
	MerchantID    uint64
	InvoiceType   InvoiceType
	InvoiceMedium InvoiceMedium
	Title         InvoiceTitle
	Items         []InvoiceItemRequest
	Amount        int64
	TaxRate       string
	Remark        string
}

// InvoiceItemRequest 发票明细请求
type InvoiceItemRequest struct {
	ProductName string
	Spec        string
	Unit        string
	Quantity    int32
	Price       int64
	Amount      int64
	TaxRate     string
	TaxAmount   int64
}

// IssueInvoiceResult 开具发票结果
type IssueInvoiceResult struct {
	InvoiceCode string
	InvoiceNo   string
	CheckCode   string
	PDFUrl      string
	XMLUrl      string
	IssuedAt    string
}

// RedInvoiceRequest 红冲发票请求
type RedInvoiceRequest struct {
	OriginalInvoiceCode string
	OriginalInvoiceNo   string
	Reason              string
}

// RedInvoiceResult 红冲发票结果
type RedInvoiceResult struct {
	RedInvoiceCode string
	RedInvoiceNo   string
	CheckCode      string
	PDFUrl         string
	XMLUrl         string
}

// VerifyInvoiceRequest 验真发票请求
type VerifyInvoiceRequest struct {
	InvoiceCode string
	InvoiceNo   string
	CheckCode   string
	Amount      int64
	IssueDate   string
}

// QueryInvoiceRequest 查询发票请求
type QueryInvoiceRequest struct {
	InvoiceCode string
	InvoiceNo   string
}

// QueryInvoiceResult 查询发票结果
type QueryInvoiceResult struct {
	InvoiceCode string
	InvoiceNo   string
	Status      string
	PDFUrl      string
	XMLUrl      string
}

// DownloadInvoiceRequest 下载发票请求
type DownloadInvoiceRequest struct {
	InvoiceCode string
	InvoiceNo   string
	FileType    string
}

// DownloadInvoiceResult 下载发票结果
type DownloadInvoiceResult struct {
	FileUrl    string
	FileData   []byte
	ExpireTime string
}

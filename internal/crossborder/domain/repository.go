// 变更说明：完善跨境电商仓储接口定义
package domain

import (
	"context"

	"github.com/shopspring/decimal"
)

// CrossBorderRepository 跨境电商仓储接口
type CrossBorderRepository interface {
	// SaveDeclaration 保存报关单
	SaveDeclaration(ctx context.Context, decl *CustomsDeclaration) error
	
	// UpdateDeclaration 更新报关单
	UpdateDeclaration(ctx context.Context, decl *CustomsDeclaration) error
	
	// GetDeclaration 获取报关单
	GetDeclaration(ctx context.Context, declarationID string) (*CustomsDeclaration, error)
	
	// GetDeclarationByOrder 根据订单ID获取报关单
	GetDeclarationByOrder(ctx context.Context, orderID string) (*CustomsDeclaration, error)
	
	// ListDeclarations 获取报关单列表
	ListDeclarations(ctx context.Context, page, pageSize int, status DeclarationStatus, userID uint64) ([]*CustomsDeclaration, int64, error)
	
	// WithTx 在事务中执行
	WithTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

// HSCodeRepository HS编码仓储接口
type HSCodeRepository interface {
	// Save 保存HS编码
	Save(ctx context.Context, hsCode *HSCode) error
	
	// Get 获取HS编码
	Get(ctx context.Context, code string) (*HSCode, error)
	
	// GetByCodes 批量获取HS编码
	GetByCodes(ctx context.Context, codes []string) (map[string]*HSCode, error)
	
	// Search 搜索HS编码
	Search(ctx context.Context, keyword string, page, pageSize int) ([]*HSCode, int64, error)
	
	// Update 更新HS编码
	Update(ctx context.Context, hsCode *HSCode) error
}

// CrossBorderOrderRepository 跨境订单仓储接口
type CrossBorderOrderRepository interface {
	// Save 保存跨境订单
	Save(ctx context.Context, order *CrossBorderOrder) error
	
	// Get 获取跨境订单
	Get(ctx context.Context, crossBorderOrderID string) (*CrossBorderOrder, error)
	
	// GetByOrderID 根据订单ID获取
	GetByOrderID(ctx context.Context, orderID string) (*CrossBorderOrder, error)
	
	// Update 更新跨境订单
	Update(ctx context.Context, order *CrossBorderOrder) error
}

// CustomsDocumentRepository 报关证件仓储接口
type CustomsDocumentRepository interface {
	// Save 保存证件
	Save(ctx context.Context, doc *CustomsDocument) error
	
	// GetByDeclarationID 根据报关单ID获取证件列表
	GetByDeclarationID(ctx context.Context, declarationID string) ([]*CustomsDocument, error)
	
	// Delete 删除证件
	Delete(ctx context.Context, documentID string) error
}

// ClearanceEventRepository 清关事件仓储接口
type ClearanceEventRepository interface {
	// Save 保存清关事件
	Save(ctx context.Context, event *ClearanceEvent) error
	
	// GetByDeclarationID 根据报关单ID获取事件列表
	GetByDeclarationID(ctx context.Context, declarationID string) ([]*ClearanceEvent, error)
}

// CrossBorderReadRepository 跨境电商读模型仓储接口（Redis）
type CrossBorderReadRepository interface {
	// SaveDeclaration 保存报关单到缓存
	SaveDeclaration(ctx context.Context, decl *CustomsDeclaration) error
	
	// GetDeclaration 从缓存获取报关单
	GetDeclaration(ctx context.Context, declarationID string) (*CustomsDeclaration, error)
	
	// DeleteDeclaration 删除报关单缓存
	DeleteDeclaration(ctx context.Context, declarationID string) error
}

// TaxCalculatorService 税费计算服务接口
type TaxCalculatorService interface {
	// CalculateDuty 计算关税
	CalculateDuty(ctx context.Context, items []*DeclarationItem, destinationCountry, originCountry string, tradeMode TradeMode) (*TaxResult, error)
	
	// CalculateTax 计算综合税费
	CalculateTax(ctx context.Context, items []*DeclarationItem, customsPort string, tradeMode TradeMode) (*TaxResult, error)
}

// TaxResult 税费计算结果
type TaxResult struct {
	DutyAmount       decimal.Decimal
	VATAmount        decimal.Decimal
	ConsumptionTax   decimal.Decimal
	TotalTax         decimal.Decimal
	Currency         string
	Breakdown        []TaxBreakdown
}

// TaxBreakdown 税费明细
type TaxBreakdown struct {
	TaxType      string
	TaxRate      decimal.Decimal
	TaxableAmount decimal.Decimal
	TaxAmount    decimal.Decimal
}

// CustomsGatewayService 海关网关服务接口
type CustomsGatewayService interface {
	// SubmitDeclaration 提交海关申报
	SubmitDeclaration(ctx context.Context, decl *CustomsDeclaration) (*CustomsSubmitResult, error)
	
	// QueryStatus 查询海关状态
	QueryStatus(ctx context.Context, customsDeclNo string) (*CustomsStatusResult, error)
	
	// QueryResult 查询海关结果
	QueryResult(ctx context.Context, customsDeclNo string) (*CustomsResultDetail, error)
}

// CustomsSubmitResult 海关提交结果
type CustomsSubmitResult struct {
	CustomsDeclarationNo string
	Status               string
	Message              string
}

// CustomsStatusResult 海关状态结果
type CustomsStatusResult struct {
	CustomsDeclarationNo string
	Status               string
	StatusMessage        string
}

// CustomsResultDetail 海关结果详情
type CustomsResultDetail struct {
	CustomsDeclarationNo string
	Status               string
	Result               string
	Issues               []string
}

// DocumentStorageService 证件存储服务接口
type DocumentStorageService interface {
	// UploadDocument 上传证件
	UploadDocument(ctx context.Context, declarationID string, docType CustomsDocumentType, data []byte) (string, error)
	
	// GetDocumentURL 获取证件URL
	GetDocumentURL(ctx context.Context, documentID string) (string, error)
	
	// DeleteDocument 删除证件
	DeleteDocument(ctx context.Context, documentID string) error
}

// NotificationService 通知服务接口
type NotificationService interface {
	// NotifyDeclarationSubmitted 通知报关已提交
	NotifyDeclarationSubmitted(ctx context.Context, userID uint64, declarationID string) error
	
	// NotifyDeclarationCleared 通知报关已完成
	NotifyDeclarationCleared(ctx context.Context, userID uint64, declarationID string) error
	
	// NotifyDeclarationRejected 通知报关被拒绝
	NotifyDeclarationRejected(ctx context.Context, userID uint64, declarationID, reason string) error
}

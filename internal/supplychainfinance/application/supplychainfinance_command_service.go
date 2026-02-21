// Package application 供应链金融命令服务（CQRS 写侧）
// 生成摘要：
// 1) 实现供应链金融服务的写操作命令，包括融资申请、授信额度管理、还款等
// 2) 支持发票融资、订单融资、库存融资等多种融资模式
// 3) 集成风控评估、额度控制、还款计划生成等业务逻辑
package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/supplychainfinance/domain"
	ftriskpb "github.com/wyfcoding/financialtrading/go-api/risk/v1"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/messagequeue"
)

// SupplyChainFinanceCommandService 供应链金融命令服务
type SupplyChainFinanceCommandService struct {
	applicationRepo    domain.FinanceApplicationRepository
	creditLineRepo     domain.CreditLineRepository
	invoiceFinanceRepo domain.InvoiceFinancingRepository
	supplierRiskRepo   domain.SupplierRiskProfileRepository
	buyerRiskRepo      domain.BuyerRiskProfileRepository
	collateralRepo     domain.CollateralRepository
	guaranteeRepo      domain.GuaranteeRepository
	repaymentPlanRepo  domain.RepaymentPlanRepository
	accountRepo        domain.FinanceAccountRepository
	riskCli            ftriskpb.RiskServiceClient
	eventBus           messagequeue.EventBus
	logger             *slog.Logger
}

// NewSupplyChainFinanceCommandService 创建供应链金融命令服务实例
func NewSupplyChainFinanceCommandService(
	applicationRepo domain.FinanceApplicationRepository,
	creditLineRepo domain.CreditLineRepository,
	invoiceFinanceRepo domain.InvoiceFinancingRepository,
	supplierRiskRepo domain.SupplierRiskProfileRepository,
	buyerRiskRepo domain.BuyerRiskProfileRepository,
	collateralRepo domain.CollateralRepository,
	guaranteeRepo domain.GuaranteeRepository,
	repaymentPlanRepo domain.RepaymentPlanRepository,
	accountRepo domain.FinanceAccountRepository,
	riskCli ftriskpb.RiskServiceClient,
	eventBus messagequeue.EventBus,
	logger *slog.Logger,
) *SupplyChainFinanceCommandService {
	return &SupplyChainFinanceCommandService{
		applicationRepo:    applicationRepo,
		creditLineRepo:     creditLineRepo,
		invoiceFinanceRepo: invoiceFinanceRepo,
		supplierRiskRepo:   supplierRiskRepo,
		buyerRiskRepo:      buyerRiskRepo,
		collateralRepo:     collateralRepo,
		guaranteeRepo:      guaranteeRepo,
		repaymentPlanRepo:  repaymentPlanRepo,
		accountRepo:        accountRepo,
		riskCli:            riskCli,
		eventBus:           eventBus,
		logger:             logger.With("module", "supplychainfinance_command"),
	}
}

// ApplyFinanceCmd 融资申请命令
type ApplyFinanceCmd struct {
	ApplicantID     string             `json:"applicant_id" validate:"required"`
	ApplicantName   string             `json:"applicant_name" validate:"required"`
	ApplicantType   string             `json:"applicant_type" validate:"required"`
	FinanceType     domain.FinanceType `json:"finance_type" validate:"required"`
	RequestedAmount float64            `json:"requested_amount" validate:"required,gt=0"`
	Currency        string             `json:"currency" validate:"required"`
	TermDays        int                `json:"term_days" validate:"required,gt=0"`
	Purpose         string             `json:"purpose" validate:"required"`
	InvoiceInfo     *InvoiceInfo       `json:"invoice_info,omitempty"`
	OrderInfo       *OrderInfo         `json:"order_info,omitempty"`
	CollateralInfo  []*CollateralInfo  `json:"collateral_info,omitempty"`
	GuaranteeInfo   []*GuaranteeInfo   `json:"guarantee_info,omitempty"`
}

// InvoiceInfo 发票信息（用于发票融资）
type InvoiceInfo struct {
	InvoiceID      string  `json:"invoice_id" validate:"required"`
	InvoiceNumber  string  `json:"invoice_number" validate:"required"`
	InvoiceAmount  float64 `json:"invoice_amount" validate:"required,gt=0"`
	InvoiceDate    string  `json:"invoice_date" validate:"required"`
	InvoiceDueDate string  `json:"invoice_due_date" validate:"required"`
	BuyerID        string  `json:"buyer_id" validate:"required"`
	BuyerName      string  `json:"buyer_name" validate:"required"`
}

// OrderInfo 订单信息（用于订单融资）
type OrderInfo struct {
	OrderID      string  `json:"order_id" validate:"required"`
	OrderNumber  string  `json:"order_number" validate:"required"`
	OrderAmount  float64 `json:"order_amount" validate:"required,gt=0"`
	BuyerID      string  `json:"buyer_id" validate:"required"`
	BuyerName    string  `json:"buyer_name" validate:"required"`
	DeliveryDate string  `json:"delivery_date" validate:"required"`
}

// CollateralInfo 抵押品信息
type CollateralInfo struct {
	CollateralType domain.CollateralType `json:"collateral_type" validate:"required"`
	Description    string                `json:"description" validate:"required"`
	OriginalValue  float64               `json:"original_value" validate:"required,gt=0"`
	Currency       string                `json:"currency" validate:"required"`
	Location       string                `json:"location"`
	Custodian      string                `json:"custodian"`
}

// GuaranteeInfo 担保信息
type GuaranteeInfo struct {
	GuarantorID   string               `json:"guarantor_id" validate:"required"`
	GuarantorName string               `json:"guarantor_name" validate:"required"`
	GuaranteeType domain.GuaranteeType `json:"guarantee_type" validate:"required"`
	Amount        float64              `json:"amount" validate:"required,gt=0"`
	Currency      string               `json:"currency" validate:"required"`
}

// ApplyFinance 提交融资申请
func (s *SupplyChainFinanceCommandService) ApplyFinance(ctx context.Context, cmd *ApplyFinanceCmd) (*domain.FinanceApplication, error) {
	start := time.Now()

	// 创建融资申请
	application := domain.NewFinanceApplication(
		cmd.ApplicantID,
		cmd.ApplicantName,
		cmd.FinanceType,
		decimal.NewFromFloat(cmd.RequestedAmount),
		cmd.Currency,
		cmd.TermDays,
		cmd.Purpose,
	)
	application.ID = fmt.Sprintf("FA%s", idgen.GenIDString())

	// 添加发票信息（如果是发票融资）
	if cmd.FinanceType == domain.FinanceTypeInvoice && cmd.InvoiceInfo != nil {
		// 这里可以添加发票验证逻辑
		application.AddDocument("invoice", cmd.InvoiceInfo.InvoiceID)
	}

	// 添加订单信息（如果是订单融资）
	if cmd.FinanceType == domain.FinanceTypePurchaseOrder && cmd.OrderInfo != nil {
		application.AddDocument("purchase_order", cmd.OrderInfo.OrderID)
	}

	// 添加抵押品信息
	for _, collateral := range cmd.CollateralInfo {
		collateralID := fmt.Sprintf("COL%s", idgen.GenIDString())
		collateralEntity := domain.NewCollateral(
			cmd.ApplicantID,
			cmd.ApplicantName,
			collateral.CollateralType,
			decimal.NewFromFloat(collateral.OriginalValue),
			collateral.Currency,
		)
		collateralEntity.ID = collateralID
		collateralEntity.Description = collateral.Description
		collateralEntity.Location = collateral.Location
		collateralEntity.Custodian = collateral.Custodian

		if err := s.collateralRepo.Create(ctx, collateralEntity); err != nil {
			s.logger.WarnContext(ctx, "failed to create collateral", "error", err)
		} else {
			application.AddCollateral(collateralID)
		}
	}

	// 添加担保信息
	for _, guarantee := range cmd.GuaranteeInfo {
		guaranteeID := fmt.Sprintf("GUA%s", idgen.GenIDString())
		guaranteeEntity := domain.NewGuarantee(
			guarantee.GuarantorID,
			guarantee.GuarantorName,
			cmd.ApplicantID,
			cmd.ApplicantName,
			guarantee.GuaranteeType,
			decimal.NewFromFloat(guarantee.Amount),
			guarantee.Currency,
		)
		guaranteeEntity.ID = guaranteeID

		if err := s.guaranteeRepo.Create(ctx, guaranteeEntity); err != nil {
			s.logger.WarnContext(ctx, "failed to create guarantee", "error", err)
		} else {
			application.AddGuarantee(guaranteeID)
		}
	}

	// 异步或同步请求金融交易系统的风控评估
	riskResp, err := s.riskCli.AssessRisk(ctx, &ftriskpb.AssessRiskRequest{
		UserId:   cmd.ApplicantID,
		Symbol:   "SCF_FINANCE", // 虚拟 Symbol 用于风控引擎识别业务类型
		Quantity: fmt.Sprintf("%f", cmd.RequestedAmount),
	})
	if err == nil && riskResp != nil {
		score, _ := decimal.NewFromString(riskResp.RiskScore)
		application.SetRiskAssessment(domain.RiskLevel(riskResp.RiskLevel), score)
	} else {
		s.logger.WarnContext(ctx, "failed to call FT risk assessment", "error", err)
	}

	// 提交申请
	application.Submit()

	// 保存申请
	if err := s.applicationRepo.Create(ctx, application); err != nil {
		s.logger.ErrorContext(ctx, "failed to create finance application",
			"applicant_id", cmd.ApplicantID, "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("create finance application: %w", err)
	}

	s.logger.InfoContext(ctx, "finance application submitted",
		"application_id", application.ID, "applicant", cmd.ApplicantName, "duration", time.Since(start))
	return application, nil
}

// ApproveFinanceCmd 审批融资申请命令
type ApproveFinanceCmd struct {
	ApplicationID  string  `json:"application_id" validate:"required"`
	ApprovedAmount float64 `json:"approved_amount" validate:"required,gt=0"`
	InterestRate   float64 `json:"interest_rate" validate:"required,gt=0"`
	FeeAmount      float64 `json:"fee_amount" validate:"gte=0"`
	Operator       string  `json:"operator" validate:"required"`
	Remark         string  `json:"remark"`
}

// ApproveFinance 审批通过融资申请
func (s *SupplyChainFinanceCommandService) ApproveFinance(ctx context.Context, cmd *ApproveFinanceCmd) error {
	start := time.Now()

	application, err := s.applicationRepo.FindByID(ctx, cmd.ApplicationID)
	if err != nil || application == nil {
		return fmt.Errorf("finance application not found: %w", err)
	}

	// 检查申请状态
	if application.Status != domain.FinanceStatusPending && application.Status != domain.FinanceStatusUnderReview {
		return fmt.Errorf("application status is %s, cannot approve", application.Status)
	}

	// 执行审批
	application.Approve(
		decimal.NewFromFloat(cmd.ApprovedAmount),
		decimal.NewFromFloat(cmd.InterestRate),
		decimal.NewFromFloat(cmd.FeeAmount),
		cmd.Operator,
	)

	// 更新申请状态
	if err := s.applicationRepo.Update(ctx, application); err != nil {
		s.logger.ErrorContext(ctx, "failed to update finance application",
			"application_id", cmd.ApplicationID, "error", err)
		return fmt.Errorf("update finance application: %w", err)
	}

	// 如果是发票融资，创建发票融资记录
	if application.FinanceType == domain.FinanceTypeInvoice {
		// 这里需要从申请中提取发票信息
		// 简化处理，创建空的发票融资记录
		invoiceFinance := &domain.InvoiceFinancing{
			ID:                 fmt.Sprintf("IF%s", idgen.GenIDString()),
			ApplicationID:      application.ID,
			BorrowerID:         application.ApplicantID,
			BorrowerName:       application.ApplicantName,
			FinancingAmount:    decimal.NewFromFloat(cmd.ApprovedAmount),
			InterestRate:       decimal.NewFromFloat(cmd.InterestRate),
			FeeAmount:          decimal.NewFromFloat(cmd.FeeAmount),
			DisbursementAmount: decimal.NewFromFloat(cmd.ApprovedAmount - cmd.FeeAmount),
			OutstandingAmount:  decimal.NewFromFloat(cmd.ApprovedAmount - cmd.FeeAmount),
			Status:             domain.FinanceStatusApproved,
			MaturityDate:       time.Now().AddDate(0, 0, application.TermDays),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := s.invoiceFinanceRepo.Create(ctx, invoiceFinance); err != nil {
			s.logger.WarnContext(ctx, "failed to create invoice financing", "error", err)
		}
	}

	s.logger.InfoContext(ctx, "finance application approved",
		"application_id", cmd.ApplicationID, "operator", cmd.Operator, "duration", time.Since(start))
	return nil
}

// CreateCreditLineCmd 创建授信额度命令
type CreateCreditLineCmd struct {
	OwnerID      string  `json:"owner_id" validate:"required"`
	OwnerName    string  `json:"owner_name" validate:"required"`
	OwnerType    string  `json:"owner_type" validate:"required"`
	TotalLimit   float64 `json:"total_limit" validate:"required,gt=0"`
	Currency     string  `json:"currency" validate:"required"`
	InterestRate float64 `json:"interest_rate" validate:"gt=0"`
	AnnualFee    float64 `json:"annual_fee" validate:"gte=0"`
	EffectiveTo  string  `json:"effective_to" validate:"required"`
}

// CreateCreditLine 创建授信额度
func (s *SupplyChainFinanceCommandService) CreateCreditLine(ctx context.Context, cmd *CreateCreditLineCmd) (*domain.CreditLine, error) {
	start := time.Now()

	// 解析有效期
	effectiveTo, err := time.Parse("2006-01-02", cmd.EffectiveTo)
	if err != nil {
		return nil, fmt.Errorf("invalid effective_to format: %w", err)
	}

	// 创建授信额度
	creditLine := domain.NewCreditLine(
		cmd.OwnerID,
		cmd.OwnerName,
		cmd.OwnerType,
		decimal.NewFromFloat(cmd.TotalLimit),
		cmd.Currency,
	)
	creditLine.ID = fmt.Sprintf("CL%s", idgen.GenIDString())

	// 设置自定义参数
	if cmd.InterestRate > 0 {
		creditLine.InterestRate = decimal.NewFromFloat(cmd.InterestRate)
	}
	if cmd.AnnualFee > 0 {
		creditLine.AnnualFee = decimal.NewFromFloat(cmd.AnnualFee)
	}
	creditLine.EffectiveTo = effectiveTo

	// 保存授信额度
	if err := s.creditLineRepo.Create(ctx, creditLine); err != nil {
		s.logger.ErrorContext(ctx, "failed to create credit line",
			"owner_id", cmd.OwnerID, "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("create credit line: %w", err)
	}

	s.logger.InfoContext(ctx, "credit line created",
		"credit_line_id", creditLine.ID, "owner", cmd.OwnerName, "duration", time.Since(start))
	return creditLine, nil
}

// RepayFinanceCmd 还款命令
type RepayFinanceCmd struct {
	FinancingID     string  `json:"financing_id" validate:"required"`
	Amount          float64 `json:"amount" validate:"required,gt=0"`
	RepaymentMethod string  `json:"repayment_method" validate:"required"`
	ReferenceNo     string  `json:"reference_no"`
}

// RepayFinance 执行还款
func (s *SupplyChainFinanceCommandService) RepayFinance(ctx context.Context, cmd *RepayFinanceCmd) error {
	start := time.Now()

	// 查找融资记录
	financing, err := s.invoiceFinanceRepo.FindByID(ctx, cmd.FinancingID)
	if err != nil || financing == nil {
		return fmt.Errorf("financing not found: %w", err)
	}

	// 执行还款
	financing.Repay(decimal.NewFromFloat(cmd.Amount))

	// 更新融资记录
	if err := s.invoiceFinanceRepo.Update(ctx, financing); err != nil {
		s.logger.ErrorContext(ctx, "failed to update financing",
			"financing_id", cmd.FinancingID, "error", err)
		return fmt.Errorf("update financing: %w", err)
	}

	// 更新授信额度（释放额度）
	if financing.CreditLineID != "" {
		creditLine, err := s.creditLineRepo.FindByID(ctx, financing.CreditLineID)
		if err == nil && creditLine != nil {
			creditLine.ReleaseCredit(decimal.NewFromFloat(cmd.Amount))
			if err := s.creditLineRepo.Update(ctx, creditLine); err != nil {
				s.logger.WarnContext(ctx, "failed to update credit line", "error", err)
			}
		}
	}

	s.logger.InfoContext(ctx, "finance repayment completed",
		"financing_id", cmd.FinancingID, "amount", cmd.Amount, "duration", time.Since(start))
	return nil
}

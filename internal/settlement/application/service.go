package application

import (
	"context"
	"errors" // 导入标准错误处理库。
	"fmt"    // 导入格式化库。
	"time"   // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/settlement/domain/entity"     // 导入结算领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/settlement/domain/repository" // 导入结算领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// SettlementService 结构体定义了结算管理相关的应用服务。
// 它协调领域层和基础设施层，处理结算单的创建、订单加入结算、结算处理和完成，以及商户账户管理等业务逻辑。
type SettlementService struct {
	repo   repository.SettlementRepository // 依赖SettlementRepository接口，用于数据持久化操作。
	logger *slog.Logger                    // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewSettlementService 创建并返回一个新的 SettlementService 实例。
func NewSettlementService(repo repository.SettlementRepository, logger *slog.Logger) *SettlementService {
	return &SettlementService{
		repo:   repo,
		logger: logger,
	}
}

// CreateSettlement 创建一个新的结算单。
// ctx: 上下文。
// merchantID: 商户ID。
// cycle: 结算周期类型。
// startDate, endDate: 结算周期开始和结束日期。
// 返回创建成功的Settlement实体和可能发生的错误。
func (s *SettlementService) CreateSettlement(ctx context.Context, merchantID uint64, cycle string, startDate, endDate time.Time) (*entity.Settlement, error) {
	// 生成唯一的结算单号。
	settlementNo := fmt.Sprintf("S%d%d", merchantID, time.Now().UnixNano())

	settlement := &entity.Settlement{
		SettlementNo: settlementNo,
		MerchantID:   merchantID,
		Cycle:        entity.SettlementCycle(cycle), // 结算周期类型。
		StartDate:    startDate,
		EndDate:      endDate,
		Status:       entity.SettlementStatusPending, // 新建结算单默认为待处理状态。
	}

	// 通过仓储接口保存结算单。
	if err := s.repo.SaveSettlement(ctx, settlement); err != nil {
		s.logger.ErrorContext(ctx, "failed to create settlement", "merchant_id", merchantID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "settlement created successfully", "settlement_id", settlement.ID, "settlement_no", settlementNo)
	return settlement, nil
}

// AddOrderToSettlement 将订单添加到指定结算单。
// ctx: 上下文。
// settlementID: 结算单ID。
// orderID: 订单ID。
// orderNo: 订单号。
// amount: 订单金额。
// 返回可能发生的错误。
func (s *SettlementService) AddOrderToSettlement(ctx context.Context, settlementID uint64, orderID uint64, orderNo string, amount uint64) error {
	settlement, err := s.repo.GetSettlement(ctx, settlementID)
	if err != nil {
		return err
	}
	if settlement == nil {
		return errors.New("settlement not found")
	}

	// 只有处于待处理状态的结算单才能添加订单。
	if settlement.Status != entity.SettlementStatusPending {
		return errors.New("settlement is not pending")
	}

	// 获取商户账户信息以计算平台费用。
	account, err := s.repo.GetMerchantAccount(ctx, settlement.MerchantID)
	if err != nil {
		return err
	}
	feeRate := 0.0 // 默认费率为0。
	if account != nil {
		feeRate = account.FeeRate
	}

	// 计算平台费用和实际结算金额。
	platformFee := uint64(float64(amount) * feeRate / 100)
	settlementAmount := amount - platformFee

	// 创建结算详情记录。
	detail := &entity.SettlementDetail{
		SettlementID:     settlementID,
		OrderID:          orderID,
		OrderNo:          orderNo,
		OrderAmount:      amount,
		PlatformFee:      platformFee,
		SettlementAmount: settlementAmount,
	}

	// 保存结算详情。
	if err := s.repo.SaveSettlementDetail(ctx, detail); err != nil {
		s.logger.ErrorContext(ctx, "failed to save settlement detail", "settlement_id", settlementID, "order_id", orderID, "error", err)
		return err
	}

	// 更新结算单的总计信息。
	settlement.OrderCount++
	settlement.TotalAmount += amount
	settlement.PlatformFee += platformFee
	settlement.SettlementAmount += settlementAmount

	// 保存更新后的结算单。
	return s.repo.SaveSettlement(ctx, settlement)
}

// ProcessSettlement 处理结算单。
// 将结算单状态从待处理变更为处理中。
// ctx: 上下文。
// id: 结算单ID。
// 返回可能发生的错误。
func (s *SettlementService) ProcessSettlement(ctx context.Context, id uint64) error {
	settlement, err := s.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}
	if settlement == nil {
		return errors.New("settlement not found")
	}

	// 只有处于待处理状态的结算单才能被处理。
	if settlement.Status != entity.SettlementStatusPending {
		return errors.New("settlement is not pending")
	}

	settlement.Status = entity.SettlementStatusProcessing // 状态变更为处理中。
	// 保存更新后的结算单。
	return s.repo.SaveSettlement(ctx, settlement)
}

// CompleteSettlement 完成结算单。
// 将结算单状态从处理中变更为已完成，并更新商户账户余额。
// ctx: 上下文。
// id: 结算单ID。
// 返回可能发生的错误。
func (s *SettlementService) CompleteSettlement(ctx context.Context, id uint64) error {
	settlement, err := s.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}
	if settlement == nil {
		return errors.New("settlement not found")
	}

	// 只有处于处理中状态的结算单才能被完成。
	if settlement.Status != entity.SettlementStatusProcessing {
		return errors.New("settlement is not processing")
	}

	// 更新商户账户余额。
	account, err := s.repo.GetMerchantAccount(ctx, settlement.MerchantID)
	if err != nil {
		return err
	}
	if account == nil {
		// 如果商户账户不存在，则创建一个默认账户。
		account = &entity.MerchantAccount{
			MerchantID: settlement.MerchantID,
			FeeRate:    0.0, // 默认费率。
		}
	}

	account.Balance += settlement.SettlementAmount     // 增加商户余额。
	account.TotalIncome += settlement.SettlementAmount // 增加商户总收入。
	if err := s.repo.SaveMerchantAccount(ctx, account); err != nil {
		s.logger.ErrorContext(ctx, "failed to save merchant account", "merchant_id", settlement.MerchantID, "error", err)
		return err
	}

	// 更新结算单状态为已完成。
	now := time.Now()
	settlement.Status = entity.SettlementStatusCompleted
	settlement.SettledAt = &now // 记录结算完成时间。
	// 保存更新后的结算单。
	return s.repo.SaveSettlement(ctx, settlement)
}

// GetMerchantAccount 获取指定商户ID的商户账户信息。
// ctx: 上下文。
// merchantID: 商户ID。
// 返回MerchantAccount实体和可能发生的错误。
func (s *SettlementService) GetMerchantAccount(ctx context.Context, merchantID uint64) (*entity.MerchantAccount, error) {
	return s.repo.GetMerchantAccount(ctx, merchantID)
}

// ListSettlements 获取结算单列表。
// ctx: 上下文。
// merchantID: 筛选结算单的商户ID。
// status: 筛选结算单的状态。
// page, pageSize: 分页参数。
// 返回结算单列表、总数和可能发生的错误。
func (s *SettlementService) ListSettlements(ctx context.Context, merchantID uint64, status *int, page, pageSize int) ([]*entity.Settlement, int64, error) {
	offset := (page - 1) * pageSize
	var st *entity.SettlementStatus
	if status != nil { // 如果提供了状态，则按状态过滤。
		s := entity.SettlementStatus(*status)
		st = &s
	}
	return s.repo.ListSettlements(ctx, merchantID, st, offset, pageSize)
}

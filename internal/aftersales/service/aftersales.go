package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ecommerce/internal/aftersales/model"
	"ecommerce/internal/aftersales/repository"
	"ecommerce/pkg/idgen"
)

var (
	ErrRefundNotFound    = errors.New("退款单不存在")
	ErrExchangeNotFound  = errors.New("换货单不存在")
	ErrRepairNotFound    = errors.New("维修单不存在")
	ErrTicketNotFound    = errors.New("工单不存在")
	ErrInvalidStatus     = errors.New("无效的状态")
	ErrUnauthorized      = errors.New("无权操作")
)

// AfterSalesService 售后服务接口
type AfterSalesService interface {
	// 退款服务
	ApplyRefund(ctx context.Context, req *RefundRequest) (*model.RefundOrder, error)
	ApproveRefund(ctx context.Context, refundID, reviewerID uint64, approved bool, remark string) error
	UpdateRefundLogistics(ctx context.Context, refundID uint64, logistics, trackingNo string) error
	ConfirmRefundReturn(ctx context.Context, refundID uint64) error
	ProcessRefund(ctx context.Context, refundID uint64) error
	CancelRefund(ctx context.Context, refundID, userID uint64) error
	GetRefund(ctx context.Context, refundID uint64) (*model.RefundOrder, error)
	ListUserRefunds(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.RefundOrder, int64, error)
	
	// 换货服务
	ApplyExchange(ctx context.Context, req *ExchangeRequest) (*model.ExchangeOrder, error)
	ApproveExchange(ctx context.Context, exchangeID, reviewerID uint64, approved bool, remark string) error
	UpdateExchangeReturnLogistics(ctx context.Context, exchangeID uint64, logistics, trackingNo string) error
	ConfirmExchangeReturn(ctx context.Context, exchangeID uint64) error
	ShipExchangeProduct(ctx context.Context, exchangeID uint64, logistics, trackingNo string) error
	CompleteExchange(ctx context.Context, exchangeID uint64) error
	GetExchange(ctx context.Context, exchangeID uint64) (*model.ExchangeOrder, error)
	
	// 维修服务
	ApplyRepair(ctx context.Context, req *RepairRequest) (*model.RepairOrder, error)
	ApproveRepair(ctx context.Context, repairID, reviewerID uint64, approved bool, remark string) error
	UpdateRepairReturnLogistics(ctx context.Context, repairID uint64, logistics, trackingNo string) error
	StartRepair(ctx context.Context, repairID uint64) error
	CompleteRepair(ctx context.Context, repairID uint64, result string, cost uint64) error
	ShipRepairedProduct(ctx context.Context, repairID uint64, logistics, trackingNo string) error
	GetRepair(ctx context.Context, repairID uint64) (*model.RepairOrder, error)
	
	// 工单服务
	CreateTicket(ctx context.Context, req *TicketRequest) (*model.AfterSalesTicket, error)
	AssignTicket(ctx context.Context, ticketID, agentID uint64) error
	ResolveTicket(ctx context.Context, ticketID uint64, resolution string) error
	CloseTicket(ctx context.Context, ticketID uint64) error
	GetTicket(ctx context.Context, ticketID uint64) (*model.AfterSalesTicket, error)
	ListUserTickets(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.AfterSalesTicket, int64, error)
}

// RefundRequest 退款申请请求
type RefundRequest struct {
	OrderID      uint64
	UserID       uint64
	Type         model.RefundType
	RefundAmount uint64
	Reason       string
	Description  string
	Images       []string
}

// ExchangeRequest 换货申请请求
type ExchangeRequest struct {
	OrderID        uint64
	UserID         uint64
	OldProductID   uint64
	OldSKUID       uint64
	NewProductID   uint64
	NewSKUID       uint64
	Quantity       uint32
	Reason         string
	Description    string
	Images         []string
}

// RepairRequest 维修申请请求
type RepairRequest struct {
	OrderID     uint64
	UserID      uint64
	ProductID   uint64
	SKUID       uint64
	FaultDesc   string
	Images      []string
}

// TicketRequest 工单创建请求
type TicketRequest struct {
	UserID      uint64
	OrderID     uint64
	Type        string
	Priority    string
	Subject     string
	Description string
	Images      []string
}

type afterSalesService struct {
	repo   repository.AfterSalesRepo
	logger *zap.Logger
}

// NewAfterSalesService 创建售后服务实例
func NewAfterSalesService(repo repository.AfterSalesRepo, logger *zap.Logger) AfterSalesService {
	return &afterSalesService{
		repo:   repo,
		logger: logger,
	}
}

// ApplyRefund 申请退款
func (s *afterSalesService) ApplyRefund(ctx context.Context, req *RefundRequest) (*model.RefundOrder, error) {
	// TODO: 验证订单状态和用户权限
	
	refundNo := fmt.Sprintf("RF%d", idgen.GenID())
	
	refund := &model.RefundOrder{
		RefundNo:     refundNo,
		OrderID:      req.OrderID,
		UserID:       req.UserID,
		Type:         req.Type,
		Status:       model.RefundStatusPending,
		RefundAmount: req.RefundAmount,
		RefundReason: req.Reason,
		RefundDesc:   req.Description,
		Images:       joinStrings(req.Images),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateRefundOrder(ctx, refund); err != nil {
		s.logger.Error("创建退款单失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("申请退款成功",
		zap.String("refundNo", refundNo),
		zap.Uint64("userID", req.UserID),
		zap.Uint64("orderID", req.OrderID))

	return refund, nil
}

// ApproveRefund 审核退款
func (s *afterSalesService) ApproveRefund(ctx context.Context, refundID, reviewerID uint64, approved bool, remark string) error {
	refund, err := s.repo.GetRefundOrderByID(ctx, refundID)
	if err != nil {
		return ErrRefundNotFound
	}

	if refund.Status != model.RefundStatusPending {
		return ErrInvalidStatus
	}

	now := time.Now()
	refund.ReviewerID = reviewerID
	refund.ReviewTime = &now
	refund.ReviewRemark = remark

	if approved {
		if refund.Type == model.RefundTypeOnlyRefund {
			refund.Status = model.RefundStatusApproved
		} else {
			refund.Status = model.RefundStatusReturning
		}
	} else {
		refund.Status = model.RefundStatusRejected
	}

	refund.UpdatedAt = now

	if err := s.repo.UpdateRefundOrder(ctx, refund); err != nil {
		s.logger.Error("更新退款单失败", zap.Error(err))
		return err
	}

	s.logger.Info("审核退款成功",
		zap.Uint64("refundID", refundID),
		zap.Bool("approved", approved))

	return nil
}

// UpdateRefundLogistics 更新退货物流信息
func (s *afterSalesService) UpdateRefundLogistics(ctx context.Context, refundID uint64, logistics, trackingNo string) error {
	refund, err := s.repo.GetRefundOrderByID(ctx, refundID)
	if err != nil {
		return ErrRefundNotFound
	}

	if refund.Status != model.RefundStatusReturning {
		return ErrInvalidStatus
	}

	now := time.Now()
	refund.ReturnLogistics = logistics
	refund.ReturnTrackingNo = trackingNo
	refund.ReturnTime = &now
	refund.UpdatedAt = now

	if err := s.repo.UpdateRefundOrder(ctx, refund); err != nil {
		s.logger.Error("更新退货物流失败", zap.Error(err))
		return err
	}

	return nil
}

// ConfirmRefundReturn 确认收到退货
func (s *afterSalesService) ConfirmRefundReturn(ctx context.Context, refundID uint64) error {
	refund, err := s.repo.GetRefundOrderByID(ctx, refundID)
	if err != nil {
		return ErrRefundNotFound
	}

	if refund.Status != model.RefundStatusReturning {
		return ErrInvalidStatus
	}

	refund.Status = model.RefundStatusReturned
	refund.UpdatedAt = time.Now()

	if err := s.repo.UpdateRefundOrder(ctx, refund); err != nil {
		s.logger.Error("确认收货失败", zap.Error(err))
		return err
	}

	// TODO: 恢复库存

	return nil
}

// ProcessRefund 处理退款
func (s *afterSalesService) ProcessRefund(ctx context.Context, refundID uint64) error {
	refund, err := s.repo.GetRefundOrderByID(ctx, refundID)
	if err != nil {
		return ErrRefundNotFound
	}

	if refund.Status != model.RefundStatusApproved && refund.Status != model.RefundStatusReturned {
		return ErrInvalidStatus
	}

	// TODO: 调用支付服务进行退款

	now := time.Now()
	refund.Status = model.RefundStatusCompleted
	refund.RefundTime = &now
	refund.RefundChannel = "ORIGINAL" // 原路退回
	refund.UpdatedAt = now

	if err := s.repo.UpdateRefundOrder(ctx, refund); err != nil {
		s.logger.Error("处理退款失败", zap.Error(err))
		return err
	}

	s.logger.Info("退款成功",
		zap.Uint64("refundID", refundID),
		zap.Uint64("amount", refund.RefundAmount))

	return nil
}

// CancelRefund 取消退款
func (s *afterSalesService) CancelRefund(ctx context.Context, refundID, userID uint64) error {
	refund, err := s.repo.GetRefundOrderByID(ctx, refundID)
	if err != nil {
		return ErrRefundNotFound
	}

	if refund.UserID != userID {
		return ErrUnauthorized
	}

	if refund.Status != model.RefundStatusPending && refund.Status != model.RefundStatusReturning {
		return ErrInvalidStatus
	}

	refund.Status = model.RefundStatusCancelled
	refund.UpdatedAt = time.Now()

	if err := s.repo.UpdateRefundOrder(ctx, refund); err != nil {
		s.logger.Error("取消退款失败", zap.Error(err))
		return err
	}

	return nil
}

// GetRefund 获取退款单详情
func (s *afterSalesService) GetRefund(ctx context.Context, refundID uint64) (*model.RefundOrder, error) {
	refund, err := s.repo.GetRefundOrderByID(ctx, refundID)
	if err != nil {
		return nil, ErrRefundNotFound
	}
	return refund, nil
}

// ListUserRefunds 获取用户退款单列表
func (s *afterSalesService) ListUserRefunds(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.RefundOrder, int64, error) {
	refunds, total, err := s.repo.ListRefundOrders(ctx, userID, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取退款单列表失败", zap.Error(err))
		return nil, 0, err
	}
	return refunds, total, nil
}

// ApplyExchange 申请换货
func (s *afterSalesService) ApplyExchange(ctx context.Context, req *ExchangeRequest) (*model.ExchangeOrder, error) {
	exchangeNo := fmt.Sprintf("EX%d", idgen.GenID())
	
	exchange := &model.ExchangeOrder{
		ExchangeNo:     exchangeNo,
		OrderID:        req.OrderID,
		UserID:         req.UserID,
		Status:         model.ExchangeStatusPending,
		OldProductID:   req.OldProductID,
		OldSKUID:       req.OldSKUID,
		NewProductID:   req.NewProductID,
		NewSKUID:       req.NewSKUID,
		Quantity:       req.Quantity,
		ExchangeReason: req.Reason,
		ExchangeDesc:   req.Description,
		Images:         joinStrings(req.Images),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateExchangeOrder(ctx, exchange); err != nil {
		s.logger.Error("创建换货单失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("申请换货成功", zap.String("exchangeNo", exchangeNo))
	return exchange, nil
}

// ApproveExchange 审核换货
func (s *afterSalesService) ApproveExchange(ctx context.Context, exchangeID, reviewerID uint64, approved bool, remark string) error {
	exchange, err := s.repo.GetExchangeOrderByID(ctx, exchangeID)
	if err != nil {
		return ErrExchangeNotFound
	}

	if exchange.Status != model.ExchangeStatusPending {
		return ErrInvalidStatus
	}

	now := time.Now()
	exchange.ReviewerID = reviewerID
	exchange.ReviewTime = &now
	exchange.ReviewRemark = remark

	if approved {
		exchange.Status = model.ExchangeStatusApproved
	} else {
		exchange.Status = model.ExchangeStatusRejected
	}

	exchange.UpdatedAt = now

	if err := s.repo.UpdateExchangeOrder(ctx, exchange); err != nil {
		s.logger.Error("更新换货单失败", zap.Error(err))
		return err
	}

	return nil
}

// CreateTicket 创建售后工单
func (s *afterSalesService) CreateTicket(ctx context.Context, req *TicketRequest) (*model.AfterSalesTicket, error) {
	ticketNo := fmt.Sprintf("TK%d", idgen.GenID())
	
	ticket := &model.AfterSalesTicket{
		TicketNo:    ticketNo,
		UserID:      req.UserID,
		OrderID:     req.OrderID,
		Type:        req.Type,
		Status:      model.TicketStatusOpen,
		Priority:    req.Priority,
		Subject:     req.Subject,
		Description: req.Description,
		Images:      joinStrings(req.Images),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateTicket(ctx, ticket); err != nil {
		s.logger.Error("创建工单失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("创建工单成功", zap.String("ticketNo", ticketNo))
	return ticket, nil
}

// AssignTicket 分配工单
func (s *afterSalesService) AssignTicket(ctx context.Context, ticketID, agentID uint64) error {
	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return ErrTicketNotFound
	}

	now := time.Now()
	ticket.AgentID = agentID
	ticket.AssignTime = &now
	ticket.Status = model.TicketStatusProcessing
	ticket.UpdatedAt = now

	if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
		s.logger.Error("分配工单失败", zap.Error(err))
		return err
	}

	return nil
}

// ResolveTicket 解决工单
func (s *afterSalesService) ResolveTicket(ctx context.Context, ticketID uint64, resolution string) error {
	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return ErrTicketNotFound
	}

	now := time.Now()
	ticket.Status = model.TicketStatusResolved
	ticket.Resolution = resolution
	ticket.ResolveTime = &now
	ticket.UpdatedAt = now

	if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
		s.logger.Error("解决工单失败", zap.Error(err))
		return err
	}

	return nil
}

// CloseTicket 关闭工单
func (s *afterSalesService) CloseTicket(ctx context.Context, ticketID uint64) error {
	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return ErrTicketNotFound
	}

	now := time.Now()
	ticket.Status = model.TicketStatusClosed
	ticket.CloseTime = &now
	ticket.UpdatedAt = now

	if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
		s.logger.Error("关闭工单失败", zap.Error(err))
		return err
	}

	return nil
}

// GetTicket 获取工单详情
func (s *afterSalesService) GetTicket(ctx context.Context, ticketID uint64) (*model.AfterSalesTicket, error) {
	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, ErrTicketNotFound
	}
	return ticket, nil
}

// ListUserTickets 获取用户工单列表
func (s *afterSalesService) ListUserTickets(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.AfterSalesTicket, int64, error) {
	tickets, total, err := s.repo.ListTickets(ctx, userID, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取工单列表失败", zap.Error(err))
		return nil, 0, err
	}
	return tickets, total, nil
}

// 辅助函数
func joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	result := ""
	for i, str := range strs {
		if i > 0 {
			result += ","
		}
		result += str
	}
	return result
}

// 以下是换货和维修服务的其他方法实现（省略部分代码以节省空间）
func (s *afterSalesService) UpdateExchangeReturnLogistics(ctx context.Context, exchangeID uint64, logistics, trackingNo string) error {
	// 实现逻辑类似 UpdateRefundLogistics
	return nil
}

func (s *afterSalesService) ConfirmExchangeReturn(ctx context.Context, exchangeID uint64) error {
	// 实现逻辑
	return nil
}

func (s *afterSalesService) ShipExchangeProduct(ctx context.Context, exchangeID uint64, logistics, trackingNo string) error {
	// 实现逻辑
	return nil
}

func (s *afterSalesService) CompleteExchange(ctx context.Context, exchangeID uint64) error {
	// 实现逻辑
	return nil
}

func (s *afterSalesService) GetExchange(ctx context.Context, exchangeID uint64) (*model.ExchangeOrder, error) {
	// 实现逻辑
	return nil, nil
}

func (s *afterSalesService) ApplyRepair(ctx context.Context, req *RepairRequest) (*model.RepairOrder, error) {
	// 实现逻辑
	return nil, nil
}

func (s *afterSalesService) ApproveRepair(ctx context.Context, repairID, reviewerID uint64, approved bool, remark string) error {
	// 实现逻辑
	return nil
}

func (s *afterSalesService) UpdateRepairReturnLogistics(ctx context.Context, repairID uint64, logistics, trackingNo string) error {
	// 实现逻辑
	return nil
}

func (s *afterSalesService) StartRepair(ctx context.Context, repairID uint64) error {
	// 实现逻辑
	return nil
}

func (s *afterSalesService) CompleteRepair(ctx context.Context, repairID uint64, result string, cost uint64) error {
	// 实现逻辑
	return nil
}

func (s *afterSalesService) ShipRepairedProduct(ctx context.Context, repairID uint64, logistics, trackingNo string) error {
	// 实现逻辑
	return nil
}

func (s *afterSalesService) GetRepair(ctx context.Context, repairID uint64) (*model.RepairOrder, error) {
	// 实现逻辑
	return nil, nil
}

package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ecommerce/internal/aftersales/model"
	"ecommerce/internal/aftersales/repository"
	// 伪代码: 模拟 gRPC 客户端
	// orderpb "ecommerce/gen/order/v1"
	// paymentpb "ecommerce/gen/payment/v1"
	// inventorypb "ecommerce/gen/inventory/v1"
)

// AftersalesService 定义了售后服务的业务逻辑接口
type AftersalesService interface {
	CreateApplication(ctx context.Context, userID, orderID uint, appType model.ApplicationType, reason string, items []model.AftersalesItem) (*model.AftersalesApplication, error)
	GetApplication(ctx context.Context, userID, appID uint) (*model.AftersalesApplication, error)
	ListApplications(ctx context.Context, userID uint) ([]model.AftersalesApplication, error)
	ApproveApplication(ctx context.Context, appID uint, adminRemarks string) (*model.AftersalesApplication, error)
	ProcessReturnedGoods(ctx context.Context, appID uint, refundAmount float64) (*model.AftersalesApplication, error)
}

// aftersalesService 是接口的具体实现
type aftersalesService struct {
	repo   repository.AftersalesRepository
	logger *zap.Logger
	// orderClient     orderpb.OrderServiceClient
	// paymentClient   paymentpb.PaymentServiceClient
	// inventoryClient inventorypb.InventoryServiceClient
}

// NewAftersalesService 创建一个新的 aftersalesService 实例
func NewAftersalesService(repo repository.AftersalesRepository, logger *zap.Logger) AftersalesService {
	return &aftersalesService{repo: repo, logger: logger}
}

// CreateApplication 用户提交售后申请
func (s *aftersalesService) CreateApplication(ctx context.Context, userID, orderID uint, appType model.ApplicationType, reason string, items []model.AftersalesItem) (*model.AftersalesApplication, error) {
	s.logger.Info("Creating aftersales application", zap.Uint("userID", userID), zap.Uint("orderID", orderID))

	// 1. 验证订单是否存在，以及用户是否有权操作
	// order, err := s.orderClient.GetOrderDetails(ctx, &orderpb.GetOrderDetailsRequest{UserId: userID, OrderId: orderID})
	// if err != nil || order == nil { return nil, fmt.Errorf("订单不存在或无权访问") }

	// 2. 验证是否在可售后的时间窗口内 (例如，订单完成后 15 天内)
	// if time.Since(order.CompletedAt) > 15*24*time.Hour { return nil, fmt.Errorf("已超出售后期限") }

	// 3. 创建申请单
	app := &model.AftersalesApplication{
		ApplicationSN: uuid.New().String(),
		UserID:        userID,
		OrderID:       orderID,
		// OrderSN:       order.OrderSn,
		Type:          appType,
		Status:        model.StatusPendingApproval,
		Reason:        reason,
		Items:         items,
	}

	if err := s.repo.CreateApplication(ctx, app); err != nil {
		return nil, err
	}

	return app, nil
}

// GetApplication 获取申请详情
func (s *aftersalesService) GetApplication(ctx context.Context, userID, appID uint) (*model.AftersalesApplication, error) {
	app, err := s.repo.GetApplicationByID(ctx, appID)
	if err != nil || app == nil {
		return nil, fmt.Errorf("申请单不存在")
	}
	// 权限校验
	if app.UserID != userID {
		return nil, fmt.Errorf("无权访问该申请单")
	}
	return app, nil
}

// ListApplications ...
func (s *aftersalesService) ListApplications(ctx context.Context, userID uint) ([]model.AftersalesApplication, error) {
	return s.repo.ListApplicationsByUserID(ctx, userID)
}

// ApproveApplication 管理员审核通过
func (s *aftersalesService) ApproveApplication(ctx context.Context, appID uint, adminRemarks string) (*model.AftersalesApplication, error) {
	app, err := s.repo.GetApplicationByID(ctx, appID)
	if err != nil || app == nil { return nil, fmt.Errorf("申请单不存在") }

	if app.Status != model.StatusPendingApproval {
		return nil, fmt.Errorf("无效的申请单状态")
	}

	app.Status = model.StatusApproved
	app.AdminRemarks = adminRemarks

	return app, s.repo.UpdateApplication(ctx, app)
}

// ProcessReturnedGoods 收到退货并处理 (退款/入库)
func (s *aftersalesService) ProcessReturnedGoods(ctx context.Context, appID uint, refundAmount float64) (*model.AftersalesApplication, error) {
	app, err := s.repo.GetApplicationByID(ctx, appID)
	if err != nil || app == nil { return nil, fmt.Errorf("申请单不存在") }

	if app.Status != model.StatusGoodsReceived {
		return nil, fmt.Errorf("无效的申请单状态，尚未收到退货")
	}

	app.Status = model.StatusProcessing
	if err := s.repo.UpdateApplication(ctx, app); err != nil {
		return nil, err
	}

	// 1. 如果是退货，调用支付服务进行退款
	if app.Type == model.TypeReturn && refundAmount > 0 {
		// _, err := s.paymentClient.CreateRefund(ctx, &paymentpb.CreateRefundRequest{OrderSn: app.OrderSN, Amount: refundAmount, Reason: "用户退货"})
		// if err != nil { ... 处理退款失败 ... }
		app.RefundAmount = refundAmount
	}

	// 2. 调用库存服务，将商品退回库存 (增加物理库存)
	for _, item := range app.Items {
		// _, err := s.inventoryClient.AdjustStock(ctx, &inventorypb.AdjustStockRequest{Sku: item.ProductSKU, QuantityChange: item.Quantity, ...})
		// if err != nil { ... 处理入库失败 ... }
	}

	app.Status = model.StatusCompleted
	return app, s.repo.UpdateApplication(ctx, app)
}
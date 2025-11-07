package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ecommerce/internal/aftersales/model"
	"ecommerce/internal/aftersales/repository"
	// 伪代码: 模拟 gRPC 客户端
	// orderpb "ecommerce/gen/order/v1"
	// paymentpb "ecommerce/gen/payment/v1"
	// inventorypb "ecommerce/gen/inventory/v1"
)

var (
	// ErrApplicationNotFound 表示未找到售后申请。
	ErrApplicationNotFound = errors.New("aftersales application not found")
	// ErrInvalidApplicationStatus 表示售后申请状态无效。
	ErrInvalidApplicationStatus = errors.New("invalid aftersales application status")
	// ErrPermissionDenied 表示权限不足。
	ErrPermissionDenied = errors.New("permission denied")
	// ErrOrderNotFound 表示订单不存在。
	ErrOrderNotFound = errors.New("order not found")
	// ErrOrderNotInAftersalesWindow 表示订单不在售后有效期内。
	ErrOrderNotInAftersalesWindow = errors.New("order not in aftersales window")
)

// AftersalesService 定义了售后服务的业务逻辑接口。
// 包含了用户提交申请、管理员审核、处理退货退款等核心功能。
type AftersalesService interface {
	// CreateApplication 用户提交售后申请。
	// 需要验证订单状态、商品是否可售后等。
	CreateApplication(ctx context.Context, userID, orderID uint, appType model.ApplicationType, reason string, items []model.AftersalesItem) (*model.AftersalesApplication, error)
	// GetApplication 获取指定用户或管理员可访问的售后申请详情。
	GetApplication(ctx context.Context, appID uint, callerUserID *uint, isAdmin bool) (*model.AftersalesApplication, error)
	// ListApplications 列出售后申请列表，支持按用户ID、状态等过滤。
	ListApplications(ctx context.Context, userID *uint, statusFilter *model.ApplicationStatus, page, pageSize int) ([]model.AftersalesApplication, int, error)

	// 管理员操作
	// ApproveApplication 管理员审核通过售后申请。
	ApproveApplication(ctx context.Context, appID uint, adminRemarks string) (*model.AftersalesApplication, error)
	// RejectApplication 管理员拒绝售后申请。
	RejectApplication(ctx context.Context, appID uint, adminRemarks string) (*model.AftersalesApplication, error)
	// ProcessReturnedGoods 收到退货并处理 (退款/入库)。
	ProcessReturnedGoods(ctx context.Context, appID uint, refundAmount float64) (*model.AftersalesApplication, error)
	// CompleteApplication 完成售后流程。
	CompleteApplication(ctx context.Context, appID uint, adminRemarks string) (*model.AftersalesApplication, error)
	// CancelApplication 取消售后申请。
	CancelApplication(ctx context.Context, appID uint, adminRemarks string) (*model.AftersalesApplication, error)
}

// aftersalesService 是 AftersalesService 接口的具体实现。
type aftersalesService struct {
	repo   repository.AftersalesRepository
	logger *zap.Logger
	// orderClient     orderpb.OrderServiceClient
	// paymentClient   paymentpb.PaymentServiceClient
	// inventoryClient inventorypb.InventoryServiceClient
}

// NewAftersalesService 创建一个新的 aftersalesService 实例。
// 接收 AftersalesRepository 和 zap.Logger 实例，并返回 AftersalesService 接口。
func NewAftersalesService(repo repository.AftersalesRepository, logger *zap.Logger) AftersalesService {
	return &aftersalesService{repo: repo, logger: logger}
}

// CreateApplication 用户提交售后申请。
// 包含订单验证、售后期限检查、申请单创建等业务逻辑。
func (s *aftersalesService) CreateApplication(ctx context.Context, userID, orderID uint, appType model.ApplicationType, reason string, items []model.AftersalesItem) (*model.AftersalesApplication, error) {
	s.logger.Info("Creating aftersales application", zap.Uint("user_id", userID), zap.Uint("order_id", orderID), zap.String("application_type", string(appType)))

	// 1. 验证订单是否存在，以及用户是否有权操作
	// 伪代码: 调用订单服务获取订单详情
	// order, err := s.orderClient.GetOrderDetails(ctx, &orderpb.GetOrderDetailsRequest{UserId: userID, OrderId: orderID})
	// if err != nil { 
	//	s.logger.Error("Failed to get order details from order service", zap.Error(err), zap.Uint("order_id", orderID))
	//	return nil, ErrOrderNotFound 
	// }
	// if order == nil || order.UserId != uint64(userID) { 
	//	s.logger.Warn("Order not found or permission denied", zap.Uint("order_id", orderID), zap.Uint("user_id", userID))
	//	return nil, ErrPermissionDenied 
	// }

	// 2. 验证是否在可售后的时间窗口内 (例如，订单完成后 15 天内)
	// 伪代码: 假设订单已完成时间为 order.CompletedAt
	// if time.Since(order.CompletedAt) > 15*24*time.Hour { 
	//	s.logger.Warn("Order out of aftersales window", zap.Uint("order_id", orderID))
	//	return nil, ErrOrderNotInAftersalesWindow 
	// }

	// 3. 检查商品是否符合售后条件 (例如，是否已发货，是否为虚拟商品等)
	// 伪代码: 遍历 items，调用产品服务或订单服务检查
	// for _, item := range items {
	//	// product, err := s.productClient.GetProductInfo(ctx, &productpb.GetProductInfoRequest{ProductId: item.ProductID})
	//	// if product.IsVirtual { return nil, errors.New("virtual goods cannot be returned") }
	// }

	// 4. 创建申请单
	app := &model.AftersalesApplication{
		ApplicationSN: uuid.New().String(), // 生成唯一申请单号
		UserID:        userID,
		OrderID:       orderID,
		// OrderSN:       order.OrderSn, // 从订单服务获取
		Type:          appType,
		Status:        model.StatusPendingApproval, // 初始状态为待审核
		Reason:        reason,
		Items:         items,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateApplication(ctx, app); err != nil {
		s.logger.Error("Failed to create aftersales application in repository", zap.Error(err), zap.Uint("user_id", userID), zap.Uint("order_id", orderID))
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	s.logger.Info("Aftersales application created successfully", zap.String("application_sn", app.ApplicationSN))
	return app, nil
}

// GetApplication 获取指定用户或管理员可访问的售后申请详情。
// 包含权限校验逻辑。
func (s *aftersalesService) GetApplication(ctx context.Context, appID uint, callerUserID *uint, isAdmin bool) (*model.AftersalesApplication, error) {
	s.logger.Info("Getting aftersales application details", zap.Uint("application_id", appID), zap.Any("caller_user_id", callerUserID), zap.Bool("is_admin", isAdmin))

	app, err := s.repo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application by ID from repository", zap.Error(err), zap.Uint("application_id", appID))
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		s.logger.Warn("Aftersales application not found", zap.Uint("application_id", appID))
		return nil, ErrApplicationNotFound
	}

	// 权限校验: 如果不是管理员，则只能查看自己的申请
	if !isAdmin && (callerUserID == nil || app.UserID != *callerUserID) {
		s.logger.Warn("Permission denied to access application", zap.Uint("application_id", appID), zap.Any("caller_user_id", callerUserID))
		return nil, ErrPermissionDenied
	}

	s.logger.Info("Aftersales application details retrieved successfully", zap.Uint("application_id", appID))
	return app, nil
}

// ListApplications 列出售后申请列表，支持按用户ID、状态等过滤。
// 包含权限校验逻辑。
func (s *aftersalesService) ListApplications(ctx context.Context, userID *uint, statusFilter *model.ApplicationStatus, page, pageSize int) ([]model.AftersalesApplication, int, error) {
	s.logger.Info("Listing aftersales applications", zap.Any("user_id_filter", userID), zap.Any("status_filter", statusFilter), zap.Int("page", page), zap.Int("page_size", pageSize))

	apps, total, err := s.repo.ListApplications(ctx, userID, statusFilter, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to list applications from repository", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list applications: %w", err)
	}

	s.logger.Info("Aftersales applications listed successfully", zap.Int("count", len(apps)), zap.Int("total", total))
	return apps, total, nil
}

// ApproveApplication 管理员审核通过售后申请。
// 包含状态流转验证和管理员备注更新。
func (s *aftersalesService) ApproveApplication(ctx context.Context, appID uint, adminRemarks string) (*model.AftersalesApplication, error) {
	s.logger.Info("Approving aftersales application", zap.Uint("application_id", appID))

	app, err := s.repo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application by ID for approval", zap.Error(err), zap.Uint("application_id", appID))
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		s.logger.Warn("Aftersales application not found for approval", zap.Uint("application_id", appID))
		return nil, ErrApplicationNotFound
	}

	// 状态流转验证
	if app.Status != model.StatusPendingApproval {
		s.logger.Warn("Invalid application status for approval", zap.Uint("application_id", appID), zap.String("current_status", string(app.Status)))
		return nil, ErrInvalidApplicationStatus
	}

	app.Status = model.StatusApproved
	app.AdminRemarks = adminRemarks
	app.UpdatedAt = time.Now()

	if err := s.repo.UpdateApplication(ctx, app); err != nil {
		s.logger.Error("Failed to update application status to approved", zap.Error(err), zap.Uint("application_id", appID))
		return nil, fmt.Errorf("failed to approve application: %w", err)
	}

	s.logger.Info("Aftersales application approved successfully", zap.Uint("application_id", appID))
	return app, nil
}

// RejectApplication 管理员拒绝售后申请。
// 包含状态流转验证和管理员备注更新。
func (s *aftersalesService) RejectApplication(ctx context.Context, appID uint, adminRemarks string) (*model.AftersalesApplication, error) {
	s.logger.Info("Rejecting aftersales application", zap.Uint("application_id", appID))

	app, err := s.repo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application by ID for rejection", zap.Error(err), zap.Uint("application_id", appID))
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		s.logger.Warn("Aftersales application not found for rejection", zap.Uint("application_id", appID))
		return nil, ErrApplicationNotFound
	}

	// 状态流转验证
	if app.Status != model.StatusPendingApproval {
		s.logger.Warn("Invalid application status for rejection", zap.Uint("application_id", appID), zap.String("current_status", string(app.Status)))
		return nil, ErrInvalidApplicationStatus
	}

	app.Status = model.StatusRejected
	app.AdminRemarks = adminRemarks
	app.UpdatedAt = time.Now()

	if err := s.repo.UpdateApplication(ctx, app); err != nil {
		s.logger.Error("Failed to update application status to rejected", zap.Error(err), zap.Uint("application_id", appID))
		return nil, fmt.Errorf("failed to reject application: %w", err)
	}

	s.logger.Info("Aftersales application rejected successfully", zap.Uint("application_id", appID))
	return app, nil
}

// ProcessReturnedGoods 收到退货并处理 (退款/入库)。
// 包含状态流转验证、调用支付服务退款和库存服务入库等复杂业务逻辑。
func (s *aftersalesService) ProcessReturnedGoods(ctx context.Context, appID uint, refundAmount float64) (*model.AftersalesApplication, error) {
	s.logger.Info("Processing returned goods for application", zap.Uint("application_id", appID), zap.Float64("refund_amount", refundAmount))

	app, err := s.repo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application by ID for processing returned goods", zap.Error(err), zap.Uint("application_id", appID))
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		s.logger.Warn("Aftersales application not found for processing returned goods", zap.Uint("application_id", appID))
		return nil, ErrApplicationNotFound
	}

	// 状态流转验证
	if app.Status != model.StatusGoodsReceived {
		s.logger.Warn("Invalid application status for processing returned goods, expected GOODS_RECEIVED", zap.Uint("application_id", appID), zap.String("current_status", string(app.Status)))
		return nil, ErrInvalidApplicationStatus
	}

	app.Status = model.StatusProcessing
	app.UpdatedAt = time.Now()
	if err := s.repo.UpdateApplication(ctx, app); err != nil {
		s.logger.Error("Failed to update application status to processing", zap.Error(err), zap.Uint("application_id", appID))
		return nil, fmt.Errorf("failed to process returned goods: %w", err)
	}

	// 1. 如果是退货，调用支付服务进行退款
	if app.Type == model.TypeReturn && refundAmount > 0 {
		s.logger.Info("Initiating refund for application", zap.Uint("application_id", appID), zap.Float64("refund_amount", refundAmount))
		// 伪代码: 调用支付服务创建退款
		// _, err := s.paymentClient.CreateRefund(ctx, &paymentpb.CreateRefundRequest{OrderSn: app.OrderSN, Amount: refundAmount, Reason: "用户退货"})
		// if err != nil { 
		//	s.logger.Error("Failed to create refund via payment service", zap.Error(err), zap.Uint("application_id", appID))
		//	// 实际项目中可能需要回滚状态或记录退款失败日志
		//	return nil, fmt.Errorf("failed to create refund: %w", err)
		// }
		app.RefundAmount = refundAmount
		s.logger.Info("Refund initiated successfully", zap.Uint("application_id", appID))
	}

	// 2. 调用库存服务，将商品退回库存 (增加物理库存)
	s.logger.Info("Adjusting inventory for returned items", zap.Uint("application_id", appID))
	for _, item := range app.Items {
		// 伪代码: 调用库存服务调整库存
		// _, err := s.inventoryClient.AdjustStock(ctx, &inventorypb.AdjustStockRequest{Sku: item.ProductSKU, QuantityChange: item.Quantity, ...})
		// if err != nil { 
		//	s.logger.Error("Failed to adjust stock for returned item", zap.Error(err), zap.Uint("application_id", appID), zap.Uint("product_id", item.ProductID))
		//	// 实际项目中可能需要回滚状态或记录库存调整失败日志
		//	return nil, fmt.Errorf("failed to adjust stock: %w", err)
		// }
		s.logger.Debug("Inventory adjusted for item", zap.Uint("application_id", appID), zap.Uint("product_id", item.ProductID), zap.Int("quantity", item.Quantity))
	}

	app.Status = model.StatusCompleted // 假设处理完成后即完成
	app.UpdatedAt = time.Now()
	if err := s.repo.UpdateApplication(ctx, app); err != nil {
		s.logger.Error("Failed to update application status to completed after processing", zap.Error(err), zap.Uint("application_id", appID))
		return nil, fmt.Errorf("failed to complete application after processing: %w", err)
	}

	s.logger.Info("Returned goods processed and application completed successfully", zap.Uint("application_id", appID))
	return app, nil
}

// CompleteApplication 完成售后流程。
// 将售后申请状态设置为已完成。
func (s *aftersalesService) CompleteApplication(ctx context.Context, appID uint, adminRemarks string) (*model.AftersalesApplication, error) {
	s.logger.Info("Completing aftersales application", zap.Uint("application_id", appID))

	app, err := s.repo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application by ID for completion", zap.Error(err), zap.Uint("application_id", appID))
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		s.logger.Warn("Aftersales application not found for completion", zap.Uint("application_id", appID))
		return nil, ErrApplicationNotFound
	}

	// 状态流转验证
	// 只有在处理中或已批准的申请才能完成
	if app.Status != model.StatusProcessing && app.Status != model.StatusApproved {
		s.logger.Warn("Invalid application status for completion", zap.Uint("application_id", appID), zap.String("current_status", string(app.Status)))
		return nil, ErrInvalidApplicationStatus
	}

	app.Status = model.StatusCompleted
	app.AdminRemarks = adminRemarks
	app.UpdatedAt = time.Now()

	if err := s.repo.UpdateApplication(ctx, app); err != nil {
		s.logger.Error("Failed to update application status to completed", zap.Error(err), zap.Uint("application_id", appID))
		return nil, fmt.Errorf("failed to complete application: %w", err)
	}

	s.logger.Info("Aftersales application completed successfully", zap.Uint("application_id", appID))
	return app, nil
}

// CancelApplication 取消售后申请。
// 只有在待审核或已批准但未处理的申请才能取消。
func (s *aftersalesService) CancelApplication(ctx context.Context, appID uint, adminRemarks string) (*model.AftersalesApplication, error) {
	s.logger.Info("Cancelling aftersales application", zap.Uint("application_id", appID))

	app, err := s.repo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application by ID for cancellation", zap.Error(err), zap.Uint("application_id", appID))
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		s.logger.Warn("Aftersales application not found for cancellation", zap.Uint("application_id", appID))
		return nil, ErrApplicationNotFound
	}

	// 状态流转验证
	// 只有在待审核或已批准的申请才能取消
	if app.Status != model.StatusPendingApproval && app.Status != model.StatusApproved {
		s.logger.Warn("Invalid application status for cancellation", zap.Uint("application_id", appID), zap.String("current_status", string(app.Status)))
		return nil, ErrInvalidApplicationStatus
	}

	app.Status = model.StatusRejected // 取消可以视为一种拒绝
	app.AdminRemarks = adminRemarks
	app.UpdatedAt = time.Now()

	if err := s.repo.UpdateApplication(ctx, app); err != nil {
		s.logger.Error("Failed to update application status to cancelled", zap.Error(err), zap.Uint("application_id", appID))
		return nil, fmt.Errorf("failed to cancel application: %w", err)
	}

	s.logger.Info("Aftersales application cancelled successfully", zap.Uint("application_id", appID))
	return app, nil
}

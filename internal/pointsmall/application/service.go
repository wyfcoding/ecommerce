package application

import (
	"context"
	"errors" // 导入标准错误处理库。
	"fmt"    // 导入格式化库。
	"time"   // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain/entity"     // 导入积分商城领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain/repository" // 导入积分商城领域的仓储接口。
	"github.com/wyfcoding/ecommerce/pkg/idgen"                             // 导入ID生成器。

	"log/slog" // 导入结构化日志库。
)

// PointsService 结构体定义了积分商城相关的应用服务。
// 它协调领域层和基础设施层，处理积分商品的管理、用户积分兑换商品、用户积分账户管理等业务逻辑。
type PointsService struct {
	repo   repository.PointsRepository // 依赖PointsRepository接口，用于数据持久化操作。
	idGen  idgen.Generator             // ID生成器，用于生成积分订单ID。
	logger *slog.Logger                // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewPointsService 创建并返回一个新的 PointsService 实例。
func NewPointsService(repo repository.PointsRepository, idGen idgen.Generator, logger *slog.Logger) *PointsService {
	return &PointsService{
		repo:   repo,
		idGen:  idGen,
		logger: logger,
	}
}

// CreateProduct 创建积分商品。
// ctx: 上下文。
// product: 待创建的PointsProduct实体。
// 返回可能发生的错误。
func (s *PointsService) CreateProduct(ctx context.Context, product *entity.PointsProduct) error {
	return s.repo.SaveProduct(ctx, product)
}

// ListProducts 获取积分商品列表。
// ctx: 上下文。
// status: 筛选商品状态。
// page, pageSize: 分页参数。
// 返回积分商品列表、总数和可能发生的错误。
func (s *PointsService) ListProducts(ctx context.Context, status *int, page, pageSize int) ([]*entity.PointsProduct, int64, error) {
	offset := (page - 1) * pageSize
	var prodStatus *entity.PointsProductStatus
	if status != nil { // 如果提供了状态，则按状态过滤。
		s := entity.PointsProductStatus(*status)
		prodStatus = &s
	}
	return s.repo.ListProducts(ctx, prodStatus, offset, pageSize)
}

// ExchangeProduct 兑换商品。
// 这是积分商城的核心业务逻辑，涉及积分扣减、库存扣减、订单创建和交易记录。
// ctx: 上下文。
// userID: 兑换用户ID。
// productID: 兑换商品ID。
// quantity: 兑换数量。
// address, phone, receiver: 收货地址、电话和收货人信息。
// 返回创建成功的PointsOrder实体和可能发生的错误。
func (s *PointsService) ExchangeProduct(ctx context.Context, userID, productID uint64, quantity int32, address, phone, receiver string) (*entity.PointsOrder, error) {
	// 1. 获取商品信息。
	product, err := s.repo.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	if product.Status != entity.PointsProductStatusOnline {
		return nil, errors.New("product not online")
	}
	if product.Stock < quantity {
		return nil, errors.New("insufficient stock")
	}

	// 2. 获取用户积分账户信息。
	account, err := s.repo.GetAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	// TODO: 如果账户不存在，应用层是否应该自动创建？目前仓储GetAccount会返回nil。
	// 这里假设 GetAccount 总是能返回一个账户，否则需要处理账户不存在的情况。

	// 3. 检查用户积分是否充足。
	totalPoints := product.Points * int64(quantity) // 计算所需总积分。
	if account.TotalPoints-account.UsedPoints < totalPoints {
		// totalPoints - usedPoints 可能是AvailablePoints
		return nil, errors.New("insufficient points")
	}

	// TODO: 以下的扣减积分、扣减库存、创建订单和记录交易操作应该封装在一个数据库事务中，
	// 以确保原子性。如果任何一步失败，所有操作都应回滚。
	// 目前为简化示例省略了事务管理，但在生产环境中这是非常关键的。

	// 4. 扣减用户积分。
	account.UsedPoints += totalPoints // 增加已使用积分。
	if err := s.repo.SaveAccount(ctx, account); err != nil {
		return nil, err
	}

	// 5. 扣减商品库存。
	product.Stock -= quantity     // 扣减库存。
	product.SoldCount += quantity // 增加已售数量。
	if err := s.repo.SaveProduct(ctx, product); err != nil {
		// TODO: 如果这里失败，需要回滚积分扣减操作。
		return nil, err
	}

	// 6. 创建积分订单。
	orderID := s.idGen.Generate()
	orderNo := fmt.Sprintf("P%s%d", time.Now().Format("20060102"), orderID) // 生成积分订单号。
	order := &entity.PointsOrder{
		OrderNo:     orderNo,
		UserID:      userID,
		ProductID:   productID,
		ProductName: product.Name,
		Quantity:    quantity,
		Points:      product.Points,
		TotalPoints: totalPoints,
		Status:      entity.PointsOrderStatusPending, // 初始状态为待处理。
		Address:     address,
		Phone:       phone,
		Receiver:    receiver,
	}
	if err := s.repo.SaveOrder(ctx, order); err != nil {
		// TODO: 如果这里失败，需要回滚积分扣减和库存扣减操作。
		return nil, err
	}

	// 7. 记录积分交易明细。
	tx := &entity.PointsTransaction{
		UserID:      userID,
		Type:        "spend", // 交易类型为支出。
		Points:      -totalPoints,
		Description: fmt.Sprintf("Exchange product: %s", product.Name),
		RefID:       orderNo, // 关联积分订单号。
	}
	if err := s.repo.SaveTransaction(ctx, tx); err != nil {
		// TODO: 如果这里失败，需要回滚积分扣减、库存扣减和订单创建操作。
		return nil, err
	}

	s.logger.Info("product exchanged successfully", "user_id", userID, "product_id", productID, "order_no", orderNo)
	return order, nil
}

// GetAccount 获取用户积分账户信息。
// 如果账户不存在，则返回nil而非错误，由应用层进行判断或创建。
// ctx: 上下文。
// userID: 用户ID。
// 返回PointsAccount实体和可能发生的错误。
func (s *PointsService) GetAccount(ctx context.Context, userID uint64) (*entity.PointsAccount, error) {
	return s.repo.GetAccount(ctx, userID)
}

// AddPoints 增加用户积分（通常由管理员或系统触发）。
// ctx: 上下文。
// userID: 用户ID。
// points: 待增加的积分数量。
// description: 积分增加的描述。
// refID: 关联的参考ID（例如，活动ID）。
// 返回可能发生的错误。
func (s *PointsService) AddPoints(ctx context.Context, userID uint64, points int64, description, refID string) error {
	account, err := s.repo.GetAccount(ctx, userID)
	if err != nil {
		return err
	}
	// TODO: 如果账户不存在，应用层是否应该自动创建？

	account.TotalPoints += points // 增加总积分。
	// TODO: 这里只增加了TotalPoints，没有增加AvailablePoints。
	// 如果业务逻辑中TotalPoints和AvailablePoints概念不同，需要进一步明确。
	// 如果是奖励积分，通常应该增加AvailablePoints。

	if err := s.repo.SaveAccount(ctx, account); err != nil {
		return err
	}

	// 记录积分交易明细。
	tx := &entity.PointsTransaction{
		UserID:      userID,
		Type:        "earn", // 交易类型为收入。
		Points:      points,
		Description: description,
		RefID:       refID,
	}
	return s.repo.SaveTransaction(ctx, tx)
}

// ListOrders 获取积分订单列表。
// ctx: 上下文。
// userID: 筛选用户的订单。
// status: 筛选订单状态。
// page, pageSize: 分页参数。
// 返回积分订单列表、总数和可能发生的错误。
func (s *PointsService) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*entity.PointsOrder, int64, error) {
	offset := (page - 1) * pageSize
	var orderStatus *entity.PointsOrderStatus
	if status != nil { // 如果提供了状态，则按状态过滤。
		s := entity.PointsOrderStatus(*status)
		orderStatus = &s
	}
	return s.repo.ListOrders(ctx, userID, orderStatus, offset, pageSize)
}

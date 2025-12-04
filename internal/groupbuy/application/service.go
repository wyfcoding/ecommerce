package application

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain/entity"     // 导入拼团领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain/repository" // 导入拼团领域的仓储接口。
	"github.com/wyfcoding/ecommerce/pkg/idgen"                           // 导入ID生成器接口。

	"log/slog" // 导入结构化日志库。
)

// GroupbuyService 结构体定义了拼团活动相关的应用服务。
// 它协调领域层和基础设施层，处理拼团活动的创建、团队的组建和加入、以及拼团订单的管理等业务逻辑。
type GroupbuyService struct {
	repo        repository.GroupbuyRepository // 依赖GroupbuyRepository接口，用于数据持久化操作。
	idGenerator idgen.Generator               // 依赖ID生成器接口，用于生成唯一的团队编号。
	logger      *slog.Logger                  // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewGroupbuyService 创建并返回一个新的 GroupbuyService 实例。
func NewGroupbuyService(repo repository.GroupbuyRepository, idGenerator idgen.Generator, logger *slog.Logger) *GroupbuyService {
	return &GroupbuyService{
		repo:        repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// CreateGroupbuy 创建一个新的拼团活动。
// ctx: 上下文。
// name: 活动名称。
// productID, skuID: 关联的商品ID和SKU ID。
// originalPrice, groupPrice: 商品原价和拼团价格。
// minPeople, maxPeople: 拼团的最小和最大人数。
// totalStock: 拼团活动的总库存。
// startTime, endTime: 活动的开始和结束时间。
// 返回创建成功的Groupbuy实体和可能发生的错误。
func (s *GroupbuyService) CreateGroupbuy(ctx context.Context, name string, productID, skuID, originalPrice, groupPrice uint64,
	minPeople, maxPeople, totalStock int32, startTime, endTime time.Time) (*entity.Groupbuy, error) {

	groupbuy := entity.NewGroupbuy(name, productID, skuID, originalPrice, groupPrice, minPeople, maxPeople, totalStock, startTime, endTime) // 创建Groupbuy实体。

	// 通过仓储接口保存拼团活动。
	if err := s.repo.CreateGroupbuy(ctx, groupbuy); err != nil {
		s.logger.ErrorContext(ctx, "failed to create groupbuy", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "groupbuy created successfully", "groupbuy_id", groupbuy.ID, "name", name)

	return groupbuy, nil
}

// ListGroupbuys 获取拼团活动列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回拼团活动列表、总数和可能发生的错误。
func (s *GroupbuyService) ListGroupbuys(ctx context.Context, page, pageSize int) ([]*entity.Groupbuy, int64, error) {
	return s.repo.ListGroupbuys(ctx, page, pageSize)
}

// InitiateTeam 发起一个新的拼团。
// ctx: 上下文。
// groupbuyID: 关联的拼团活动ID。
// userID: 发起拼团的用户ID。
// 返回新创建的GroupbuyTeam实体、Leader的GroupbuyOrder实体和可能发生的错误。
func (s *GroupbuyService) InitiateTeam(ctx context.Context, groupbuyID, userID uint64) (*entity.GroupbuyTeam, *entity.GroupbuyOrder, error) {
	// 1. 获取拼团活动详情。
	groupbuy, err := s.repo.GetGroupbuyByID(ctx, groupbuyID)
	if err != nil {
		return nil, nil, err
	}
	// 检查拼团活动是否可用。
	if !groupbuy.IsAvailable() {
		return nil, nil, fmt.Errorf("groupbuy is not available")
	}

	// 2. 创建拼团团队。
	teamNo := fmt.Sprintf("T%d", s.idGenerator.Generate()) // 生成唯一的团队编号。
	// 设置团队过期时间，默认为24小时后，但不能超过活动结束时间。
	expireAt := time.Now().Add(24 * time.Hour)
	if expireAt.After(groupbuy.EndTime) {
		expireAt = groupbuy.EndTime
	}

	team := entity.NewGroupbuyTeam(groupbuyID, teamNo, userID, groupbuy.MaxPeople, expireAt) // 创建GroupbuyTeam实体。
	if err := s.repo.CreateTeam(ctx, team); err != nil {
		s.logger.ErrorContext(ctx, "failed to create groupbuy team", "groupbuy_id", groupbuyID, "error", err)
		return nil, nil, err
	}
	s.logger.InfoContext(ctx, "groupbuy team created successfully", "team_id", team.ID, "team_no", teamNo)

	// 3. 为团长创建拼团订单。
	// 订单数量为1，IsLeader设置为true。
	order := entity.NewGroupbuyOrder(groupbuyID, uint64(team.ID), teamNo, userID, groupbuy.ProductID, groupbuy.SkuID, groupbuy.GroupPrice, 1, true)
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		s.logger.ErrorContext(ctx, "failed to create groupbuy order", "team_id", team.ID, "user_id", userID, "error", err)
		return nil, nil, err
	}
	s.logger.InfoContext(ctx, "groupbuy order created successfully", "order_id", order.ID, "team_id", team.ID)

	return team, order, nil
}

// JoinTeam 加入一个已存在的拼团。
// ctx: 上下文。
// teamNo: 待加入团队的编号。
// userID: 加入团队的用户ID。
// 返回用户新创建的GroupbuyOrder实体和可能发生的错误。
func (s *GroupbuyService) JoinTeam(ctx context.Context, teamNo string, userID uint64) (*entity.GroupbuyOrder, error) {
	// 1. 获取拼团团队详情。
	team, err := s.repo.GetTeamByNo(ctx, teamNo)
	if err != nil {
		return nil, err
	}

	// 2. 检查团队是否可以加入。
	if err := team.Join(); err != nil {
		return nil, err
	}

	// 3. 获取拼团活动详情，以便创建订单。
	groupbuy, err := s.repo.GetGroupbuyByID(ctx, team.GroupbuyID)
	if err != nil {
		return nil, err
	}

	// 4. 更新团队状态。
	if err := s.repo.UpdateTeam(ctx, team); err != nil {
		return nil, err
	}

	// 5. 为加入团队的用户创建拼团订单。
	// 订单数量为1，IsLeader设置为false。
	order := entity.NewGroupbuyOrder(team.GroupbuyID, uint64(team.ID), teamNo, userID, groupbuy.ProductID, groupbuy.SkuID, groupbuy.GroupPrice, 1, false)
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		s.logger.ErrorContext(ctx, "failed to join groupbuy team", "team_no", teamNo, "user_id", userID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "joined groupbuy team successfully", "team_no", teamNo, "user_id", userID)

	return order, nil
}

// GetTeamDetails 获取指定ID的拼团团队详情。
// ctx: 上下文。
// teamID: 团队ID。
// 返回团队实体、所有成员订单列表和可能发生的错误。
func (s *GroupbuyService) GetTeamDetails(ctx context.Context, teamID uint64) (*entity.GroupbuyTeam, []*entity.GroupbuyOrder, error) {
	// 获取团队实体。
	team, err := s.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, nil, err
	}

	// 获取团队所有成员的订单列表。
	orders, err := s.repo.ListOrdersByTeamID(ctx, teamID)
	if err != nil {
		return nil, nil, err
	}

	return team, orders, nil
}

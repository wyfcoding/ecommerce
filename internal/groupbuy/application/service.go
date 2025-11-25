package application

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain/repository"
	"github.com/wyfcoding/ecommerce/pkg/idgen"

	"log/slog"
)

type GroupbuyService struct {
	repo        repository.GroupbuyRepository
	idGenerator idgen.Generator
	logger      *slog.Logger
}

func NewGroupbuyService(repo repository.GroupbuyRepository, idGenerator idgen.Generator, logger *slog.Logger) *GroupbuyService {
	return &GroupbuyService{
		repo:        repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// CreateGroupbuy 创建拼团活动
func (s *GroupbuyService) CreateGroupbuy(ctx context.Context, name string, productID, skuID, originalPrice, groupPrice uint64,
	minPeople, maxPeople, totalStock int32, startTime, endTime time.Time) (*entity.Groupbuy, error) {

	groupbuy := entity.NewGroupbuy(name, productID, skuID, originalPrice, groupPrice, minPeople, maxPeople, totalStock, startTime, endTime)

	if err := s.repo.CreateGroupbuy(ctx, groupbuy); err != nil {
		s.logger.Error("failed to create groupbuy", "error", err)
		return nil, err
	}

	return groupbuy, nil
}

// ListGroupbuys 获取拼团活动列表
func (s *GroupbuyService) ListGroupbuys(ctx context.Context, page, pageSize int) ([]*entity.Groupbuy, int64, error) {
	return s.repo.ListGroupbuys(ctx, page, pageSize)
}

// InitiateTeam 发起拼团
func (s *GroupbuyService) InitiateTeam(ctx context.Context, groupbuyID, userID uint64) (*entity.GroupbuyTeam, *entity.GroupbuyOrder, error) {
	// 1. 获取拼团活动
	groupbuy, err := s.repo.GetGroupbuyByID(ctx, groupbuyID)
	if err != nil {
		return nil, nil, err
	}
	if !groupbuy.IsAvailable() {
		return nil, nil, fmt.Errorf("groupbuy is not available")
	}

	// 2. 创建团队
	teamNo := fmt.Sprintf("T%d", s.idGenerator.Generate())
	expireAt := time.Now().Add(24 * time.Hour) // 默认24小时过期
	if expireAt.After(groupbuy.EndTime) {
		expireAt = groupbuy.EndTime
	}

	team := entity.NewGroupbuyTeam(groupbuyID, teamNo, userID, groupbuy.MaxPeople, expireAt)
	if err := s.repo.CreateTeam(ctx, team); err != nil {
		return nil, nil, err
	}

	// 3. 创建订单
	order := entity.NewGroupbuyOrder(groupbuyID, uint64(team.ID), teamNo, userID, groupbuy.ProductID, groupbuy.SkuID, groupbuy.GroupPrice, 1, true)
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, nil, err
	}

	return team, order, nil
}

// JoinTeam 加入拼团
func (s *GroupbuyService) JoinTeam(ctx context.Context, teamNo string, userID uint64) (*entity.GroupbuyOrder, error) {
	// 1. 获取团队
	team, err := s.repo.GetTeamByNo(ctx, teamNo)
	if err != nil {
		return nil, err
	}

	// 2. 检查是否可以加入
	if err := team.Join(); err != nil {
		return nil, err
	}

	// 3. 获取拼团活动
	groupbuy, err := s.repo.GetGroupbuyByID(ctx, team.GroupbuyID)
	if err != nil {
		return nil, err
	}

	// 4. 更新团队状态
	if err := s.repo.UpdateTeam(ctx, team); err != nil {
		return nil, err
	}

	// 5. 创建订单
	order := entity.NewGroupbuyOrder(team.GroupbuyID, uint64(team.ID), teamNo, userID, groupbuy.ProductID, groupbuy.SkuID, groupbuy.GroupPrice, 1, false)
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// GetTeamDetails 获取团队详情
func (s *GroupbuyService) GetTeamDetails(ctx context.Context, teamID uint64) (*entity.GroupbuyTeam, []*entity.GroupbuyOrder, error) {
	team, err := s.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, nil, err
	}

	orders, err := s.repo.ListOrdersByTeamID(ctx, teamID)
	if err != nil {
		return nil, nil, err
	}

	return team, orders, nil
}

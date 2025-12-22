package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
	"github.com/wyfcoding/pkg/idgen"
)

// GroupbuyManager 负责处理 Groupbuy 相关的写操作和业务逻辑。
type GroupbuyManager struct {
	repo        domain.GroupbuyRepository
	idGenerator idgen.Generator
	logger      *slog.Logger
}

// NewGroupbuyManager 负责处理 NewGroupbuy 相关的写操作和业务逻辑。
func NewGroupbuyManager(repo domain.GroupbuyRepository, idGenerator idgen.Generator, logger *slog.Logger) *GroupbuyManager {
	return &GroupbuyManager{
		repo:        repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

func (m *GroupbuyManager) CreateGroupbuy(ctx context.Context, name string, productID, skuID, originalPrice, groupPrice uint64,
	minPeople, maxPeople, totalStock int32, startTime, endTime time.Time,
) (*domain.Groupbuy, error) {
	groupbuy := domain.NewGroupbuy(name, productID, skuID, originalPrice, groupPrice, minPeople, maxPeople, totalStock, startTime, endTime)

	if err := m.repo.CreateGroupbuy(ctx, groupbuy); err != nil {
		m.logger.ErrorContext(ctx, "failed to create groupbuy", "name", name, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "groupbuy created successfully", "groupbuy_id", groupbuy.ID, "name", name)

	return groupbuy, nil
}

func (m *GroupbuyManager) InitiateTeam(ctx context.Context, groupbuyID, userID uint64) (*domain.GroupbuyTeam, *domain.GroupbuyOrder, error) {
	groupbuy, err := m.repo.GetGroupbuyByID(ctx, groupbuyID)
	if err != nil {
		return nil, nil, err
	}
	if !groupbuy.IsAvailable() {
		return nil, nil, fmt.Errorf("groupbuy is not available")
	}

	teamNo := fmt.Sprintf("T%d", m.idGenerator.Generate())
	expireAt := time.Now().Add(24 * time.Hour)
	if expireAt.After(groupbuy.EndTime) {
		expireAt = groupbuy.EndTime
	}

	team := domain.NewGroupbuyTeam(groupbuyID, teamNo, userID, groupbuy.MaxPeople, expireAt)
	if err := m.repo.CreateTeam(ctx, team); err != nil {
		m.logger.ErrorContext(ctx, "failed to create groupbuy team", "groupbuy_id", groupbuyID, "error", err)
		return nil, nil, err
	}
	m.logger.InfoContext(ctx, "groupbuy team created successfully", "team_id", team.ID, "team_no", teamNo)

	order := domain.NewGroupbuyOrder(groupbuyID, uint64(team.ID), teamNo, userID, groupbuy.ProductID, groupbuy.SkuID, groupbuy.GroupPrice, 1, true)
	if err := m.repo.CreateOrder(ctx, order); err != nil {
		m.logger.ErrorContext(ctx, "failed to create groupbuy order", "team_id", team.ID, "user_id", userID, "error", err)
		return nil, nil, err
	}
	m.logger.InfoContext(ctx, "groupbuy order created successfully", "order_id", order.ID, "team_id", team.ID)

	return team, order, nil
}

func (m *GroupbuyManager) JoinTeam(ctx context.Context, teamNo string, userID uint64) (*domain.GroupbuyOrder, error) {
	team, err := m.repo.GetTeamByNo(ctx, teamNo)
	if err != nil {
		return nil, err
	}

	if err := team.Join(); err != nil {
		return nil, err
	}

	groupbuy, err := m.repo.GetGroupbuyByID(ctx, team.GroupbuyID)
	if err != nil {
		return nil, err
	}

	if err := m.repo.UpdateTeam(ctx, team); err != nil {
		return nil, err
	}

	order := domain.NewGroupbuyOrder(team.GroupbuyID, uint64(team.ID), teamNo, userID, groupbuy.ProductID, groupbuy.SkuID, groupbuy.GroupPrice, 1, false)
	if err := m.repo.CreateOrder(ctx, order); err != nil {
		m.logger.ErrorContext(ctx, "failed to join groupbuy team", "team_no", teamNo, "user_id", userID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "joined groupbuy team successfully", "team_no", teamNo, "user_id", userID)

	return order, nil
}

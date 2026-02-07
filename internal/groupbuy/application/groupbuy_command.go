package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/optimization"
	"github.com/wyfcoding/pkg/idgen"
)

// GroupbuyCommandService 负责处理 Groupbuy 相关的写操作和业务逻辑。
type GroupbuyCommandService struct {
	repo        domain.GroupbuyRepository
	publisher   domain.EventPublisher
	idGenerator idgen.Generator
	logger      *slog.Logger
	matcher     *algorithm.GroupBuyMatcher
}

// NewGroupbuyCommandService 构造函数。
func NewGroupbuyCommandService(repo domain.GroupbuyRepository, publisher domain.EventPublisher, idGenerator idgen.Generator, logger *slog.Logger) *GroupbuyCommandService {
	return &GroupbuyCommandService{
		repo:        repo,
		publisher:   publisher,
		idGenerator: idGenerator,
		logger:      logger,
		matcher:     algorithm.NewGroupBuyMatcher(),
	}
}

func (m *GroupbuyCommandService) CreateGroupbuy(ctx context.Context, name string, productID, skuID, originalPrice, groupPrice uint64,
	minPeople, maxPeople, totalStock int32, startTime, endTime time.Time,
) (*domain.Groupbuy, error) {
	groupbuy := domain.NewGroupbuy(name, productID, skuID, originalPrice, groupPrice, minPeople, maxPeople, totalStock, startTime, endTime)

	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.CreateGroupbuyInTx(ctx, tx, groupbuy); err != nil {
			m.logger.ErrorContext(ctx, "failed to create groupbuy", "name", name, "error", err)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		return m.publisher.PublishInTx(ctx, tx, domain.GroupbuyCreatedEventType, fmt.Sprintf("%d", groupbuy.ID), &domain.GroupBuyCreatedEvent{
			GroupBuyID: uint64(groupbuy.ID),
			ProductID:  groupbuy.ProductID,
			CreatorID:  0,
			Timestamp:  time.Now(),
		})
	}); err != nil {
		return nil, err
	}
	m.logger.InfoContext(ctx, "groupbuy created successfully", "groupbuy_id", groupbuy.ID, "name", name)

	return groupbuy, nil
}

func (m *GroupbuyCommandService) InitiateTeam(ctx context.Context, groupbuyID, userID uint64) (*domain.GroupbuyTeam, *domain.GroupbuyOrder, error) {
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
	order := domain.NewGroupbuyOrder(groupbuyID, uint64(team.ID), teamNo, userID, groupbuy.ProductID, groupbuy.SkuID, groupbuy.GroupPrice, 1, true)
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.CreateTeamInTx(ctx, tx, team); err != nil {
			m.logger.ErrorContext(ctx, "failed to create groupbuy team", "groupbuy_id", groupbuyID, "error", err)
			return err
		}
		order.TeamID = uint64(team.ID)
		if err := m.repo.CreateOrderInTx(ctx, tx, order); err != nil {
			m.logger.ErrorContext(ctx, "failed to create groupbuy order", "team_id", team.ID, "user_id", userID, "error", err)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		return m.publisher.PublishInTx(ctx, tx, domain.GroupbuyJoinedEventType, fmt.Sprintf("%d", order.ID), &domain.GroupBuyJoinedEvent{
			GroupBuyID: groupbuyID,
			UserID:     userID,
			OrderID:    uint64(order.ID),
			TeamID:     uint64(team.ID),
			Timestamp:  time.Now(),
		})
	}); err != nil {
		return nil, nil, err
	}
	m.logger.InfoContext(ctx, "groupbuy team created successfully", "team_id", team.ID, "team_no", teamNo)
	m.logger.InfoContext(ctx, "groupbuy order created successfully", "order_id", order.ID, "team_id", team.ID)

	return team, order, nil
}

func (m *GroupbuyCommandService) JoinTeam(ctx context.Context, teamNo string, userID uint64) (*domain.GroupbuyOrder, error) {
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

	order := domain.NewGroupbuyOrder(team.GroupbuyID, uint64(team.ID), teamNo, userID, groupbuy.ProductID, groupbuy.SkuID, groupbuy.GroupPrice, 1, false)
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.UpdateTeamInTx(ctx, tx, team); err != nil {
			return err
		}
		if err := m.repo.CreateOrderInTx(ctx, tx, order); err != nil {
			m.logger.ErrorContext(ctx, "failed to join groupbuy team", "team_no", teamNo, "user_id", userID, "error", err)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		if err := m.publisher.PublishInTx(ctx, tx, domain.GroupbuyJoinedEventType, fmt.Sprintf("%d", order.ID), &domain.GroupBuyJoinedEvent{
			GroupBuyID: team.GroupbuyID,
			UserID:     userID,
			OrderID:    uint64(order.ID),
			TeamID:     uint64(team.ID),
			Timestamp:  time.Now(),
		}); err != nil {
			return err
		}
		if team.Status == domain.GroupbuyTeamStatusSuccess {
			return m.publisher.PublishInTx(ctx, tx, domain.GroupbuyCompletedEventType, fmt.Sprintf("%d", team.ID), &domain.GroupBuyCompletedEvent{
				GroupBuyID: team.GroupbuyID,
				TeamID:     uint64(team.ID),
				Timestamp:  time.Now(),
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	m.logger.InfoContext(ctx, "joined groupbuy team successfully", "team_no", teamNo, "user_id", userID)

	return order, nil
}

// AutoJoinTeam 自动匹配并加入一个最合适的拼团团队。
func (m *GroupbuyCommandService) AutoJoinTeam(ctx context.Context, groupbuyID, userID uint64) (*domain.GroupbuyOrder, error) {
	// 1. 获取活跃的团队列表 (获取前100个作为候选)
	teams, _, err := m.repo.ListTeamsByGroupbuyID(ctx, &domain.GroupbuyTeamQuery{
		GroupbuyID: groupbuyID,
		Page:       1,
		PageSize:   100,
	})
	if err != nil {
		return nil, err
	}

	// 2. 转换为算法需要的格式
	candidates := make([]algorithm.GroupBuyGroup, 0, len(teams))
	for _, t := range teams {
		if t.CanJoin() {
			candidates = append(candidates, algorithm.GroupBuyGroup{
				ID:            uint64(t.ID),
				ActivityID:    t.GroupbuyID,
				LeaderID:      t.LeaderID,
				RequiredCount: int(t.MaxPeople),
				CurrentCount:  int(t.CurrentPeople),
				CreatedAt:     t.CreatedAt,
				ExpireAt:      t.ExpireAt,
				Region:        "default", // 暂无地域信息
				Lat:           0,
				Lon:           0,
			})
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available teams to join")
	}

	// 3. 使用算法找到最佳团队 (优先即将成团)
	bestGroup := m.matcher.FindBestGroup(groupbuyID, 0, 0, "default", candidates, algorithm.MatchStrategyAlmostFull)
	if bestGroup == nil {
		return nil, fmt.Errorf("no suitable team found")
	}

	// 4. 找到对应的 teamNo
	var teamNo string
	for _, t := range teams {
		if uint64(t.ID) == bestGroup.ID {
			teamNo = t.TeamNo
			break
		}
	}

	// 5. 加入团队
	return m.JoinTeam(ctx, teamNo, userID)
}

// OptimizeTeamAssignments 优化团员与团长的匹配方案 (基于 Gale-Shapley 算法)
func (m *GroupbuyCommandService) OptimizeTeamAssignments(ctx context.Context, members, leaders []algorithm.Participant) map[int]int {
	if len(members) == 0 || len(leaders) == 0 {
		return nil
	}

	sm := algorithm.NewStableMarriage(members, leaders)
	assignments := sm.Match()

	m.logger.InfoContext(ctx, "optimized team assignments completed", "assignments_count", len(assignments))
	return assignments
}

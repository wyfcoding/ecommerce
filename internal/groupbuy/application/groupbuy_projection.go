// 生成摘要：新增拼团读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
)

// GroupbuyProjectionService 负责将拼团事件投影到读模型。
type GroupbuyProjectionService struct {
	repo           domain.GroupbuyRepository
	groupbuyRead   domain.GroupbuyReadRepository
	teamRead       domain.GroupbuyTeamReadRepository
	orderRead      domain.GroupbuyOrderReadRepository
	groupbuySearch domain.GroupbuySearchRepository
	orderSearch    domain.GroupbuyOrderSearchRepository
	logger         *slog.Logger
}

// NewGroupbuyProjectionService 创建投影服务。
func NewGroupbuyProjectionService(
	repo domain.GroupbuyRepository,
	groupbuyRead domain.GroupbuyReadRepository,
	teamRead domain.GroupbuyTeamReadRepository,
	orderRead domain.GroupbuyOrderReadRepository,
	groupbuySearch domain.GroupbuySearchRepository,
	orderSearch domain.GroupbuyOrderSearchRepository,
	logger *slog.Logger,
) *GroupbuyProjectionService {
	return &GroupbuyProjectionService{
		repo:           repo,
		groupbuyRead:   groupbuyRead,
		teamRead:       teamRead,
		orderRead:      orderRead,
		groupbuySearch: groupbuySearch,
		orderSearch:    orderSearch,
		logger:         logger,
	}
}

func (s *GroupbuyProjectionService) OnGroupbuyCreated(ctx context.Context, event *domain.GroupBuyCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshGroupbuy(ctx, event.GroupBuyID)
}

func (s *GroupbuyProjectionService) OnGroupbuyJoined(ctx context.Context, event *domain.GroupBuyJoinedEvent) error {
	if event == nil {
		return nil
	}
	if err := s.refreshTeam(ctx, event.TeamID); err != nil {
		return err
	}
	return s.refreshOrder(ctx, event.OrderID)
}

func (s *GroupbuyProjectionService) OnGroupbuyCompleted(ctx context.Context, event *domain.GroupBuyCompletedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshTeam(ctx, event.TeamID)
}

func (s *GroupbuyProjectionService) refreshGroupbuy(ctx context.Context, id uint64) error {
	if s.groupbuyRead == nil && s.groupbuySearch == nil {
		return nil
	}
	groupbuy, err := s.repo.GetGroupbuyByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load groupbuy for projection", "groupbuy_id", id, "error", err)
		return err
	}
	if groupbuy == nil {
		if s.groupbuyRead != nil {
			_ = s.groupbuyRead.Delete(ctx, id)
		}
		if s.groupbuySearch != nil {
			_ = s.groupbuySearch.Delete(ctx, id)
		}
		return nil
	}
	if s.groupbuyRead != nil {
		if err := s.groupbuyRead.Save(ctx, groupbuy); err != nil {
			s.logger.ErrorContext(ctx, "failed to save groupbuy cache", "groupbuy_id", id, "error", err)
			return err
		}
	}
	if s.groupbuySearch != nil {
		if err := s.groupbuySearch.Index(ctx, groupbuy); err != nil {
			s.logger.ErrorContext(ctx, "failed to index groupbuy", "groupbuy_id", id, "error", err)
			return err
		}
	}
	return nil
}

func (s *GroupbuyProjectionService) refreshTeam(ctx context.Context, teamID uint64) error {
	if s.teamRead == nil {
		return nil
	}
	team, err := s.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load groupbuy team for projection", "team_id", teamID, "error", err)
		return err
	}
	if team == nil {
		_ = s.teamRead.Delete(ctx, teamID)
		return nil
	}
	if err := s.teamRead.Save(ctx, team); err != nil {
		s.logger.ErrorContext(ctx, "failed to save team cache", "team_id", teamID, "error", err)
		return err
	}
	return nil
}

func (s *GroupbuyProjectionService) refreshOrder(ctx context.Context, orderID uint64) error {
	if s.orderRead == nil && s.orderSearch == nil {
		return nil
	}
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load groupbuy order for projection", "order_id", orderID, "error", err)
		return err
	}
	if order == nil {
		if s.orderRead != nil {
			_ = s.orderRead.Delete(ctx, orderID)
		}
		if s.orderSearch != nil {
			_ = s.orderSearch.Delete(ctx, orderID)
		}
		return nil
	}
	if s.orderRead != nil {
		if err := s.orderRead.Save(ctx, order); err != nil {
			s.logger.ErrorContext(ctx, "failed to save order cache", "order_id", orderID, "error", err)
			return err
		}
	}
	if s.orderSearch != nil {
		if err := s.orderSearch.Index(ctx, order); err != nil {
			s.logger.ErrorContext(ctx, "failed to index order", "order_id", orderID, "error", err)
			return err
		}
	}
	return nil
}

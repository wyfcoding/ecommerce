package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/social/domain"
	"github.com/wyfcoding/pkg/xerrors"
)

// 生成摘要：实现 Social 应用服务。
// 关键改动：集成仓储操作、业务校验与日志追踪。

type SocialAppService struct {
	repo domain.Repository
}

func NewSocialAppService(repo domain.Repository) *SocialAppService {
	return &SocialAppService{repo: repo}
}

// BindRelation 建立社交绑定关系。
func (s *SocialAppService) BindRelation(ctx context.Context, userID, parentID, invitationID string) error {
	// 1. 检查是否已存在
	existing, _ := s.repo.FindByUserID(ctx, userID)
	if existing != nil {
		return xerrors.AlreadyExists("user already has a social relation")
	}

	// 2. 获取上级信息，计算 Level
	level := 1
	if parentID != "" {
		parent, err := s.repo.FindByUserID(ctx, parentID)
		if err == nil && parent != nil {
			level = parent.Level + 1
		}
	}

	// 3. 创建聚合根
	rel, err := domain.NewSocialRelation(userID, parentID, invitationID, level)
	if err != nil {
		return xerrors.InvalidArg(err.Error())
	}

	// 4. 持久化 (仓储会自动处理领域分发)
	if err := s.repo.Save(ctx, rel); err != nil {
		slog.Error("failed to save social relation", "userID", userID, "error", err)
		return xerrors.Internal("persistence failed", err)
	}

	slog.Info("social relation bound successfully", "userID", userID, "parentID", parentID, "level", level)
	return nil
}

func (s *SocialAppService) GetUserNetwork(ctx context.Context, userID string) ([]*domain.SocialRelation, error) {
	return s.repo.FindChildrenByParentID(ctx, userID)
}

package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/abtest/v1"
	"github.com/wyfcoding/ecommerce/internal/abtest/domain"
	"github.com/wyfcoding/pkg/idgen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ABTestService struct {
	repo     domain.ABTestRepository
	idGen    idgen.Generator
	bucketer *domain.Bucketer
	logger   *slog.Logger
}

func NewABTestService(repo domain.ABTestRepository, idGen idgen.Generator, logger *slog.Logger) *ABTestService {
	return &ABTestService{
		repo:     repo,
		idGen:    idGen,
		bucketer: &domain.Bucketer{},
		logger:   logger.With("service", "abtest_application"),
	}
}

func (s *ABTestService) CreateExperiment(ctx context.Context, req *pb.CreateExperimentRequest) (*domain.Experiment, error) {
	variations := make([]domain.Variation, len(req.Variations))
	for i, v := range req.Variations {
		variations[i] = domain.Variation{
			Key:    v.Key,
			Value:  v.Value,
			Weight: v.Weight,
		}
	}

	exp := &domain.Experiment{
		ID:                fmt.Sprintf("exp_%d", s.idGen.Generate()),
		Name:              req.Name,
		Description:       req.Description,
		Variations:        variations,
		TrafficPercentage: req.TrafficPercentage,
		Status:            pb.ExperimentStatus_EXPERIMENT_STATUS_DRAFT,
		CreatedAt:         time.Now(),
	}

	if err := s.repo.SaveExperiment(ctx, exp); err != nil {
		return nil, err
	}
	return exp, nil
}

func (s *ABTestService) GetAssignment(ctx context.Context, userID, experimentName string) (string, string, error) {
	// 1. 获取实验定义
	exp, err := s.repo.GetExperimentByName(ctx, experimentName)
	if err != nil {
		return "", "", status.Errorf(codes.NotFound, "experiment not found: %s", experimentName)
	}

	if exp.Status != pb.ExperimentStatus_EXPERIMENT_STATUS_RUNNING {
		return "", "", nil // 实验未运行，返回空变量（默认行为）
	}

	// 2. 检查是否有持久化的分配记录
	assignment, err := s.repo.GetAssignment(ctx, userID, exp.ID)
	if err == nil {
		// 找到已存在的变量，直接返回
		for _, v := range exp.Variations {
			if v.Key == assignment.VariationKey {
				return v.Key, v.Value, nil
			}
		}
	}

	// 3. 确定性分流
	variationKey, err := s.bucketer.GetVariation(userID, exp)
	if err != nil {
		return "", "", err
	}

	if variationKey == "" {
		return "", "", nil // 不在流量范围内
	}

	// 找到 Variation Value
	variationValue := ""
	for _, v := range exp.Variations {
		if v.Key == variationKey {
			variationValue = v.Value
			break
		}
	}

	// 4. 持久化分配结果（确保同一用户在实验期间变量一致，即便权重稍有变动）
	newAssignment := &domain.Assignment{
		UserID:       userID,
		ExperimentID: exp.ID,
		VariationKey: variationKey,
	}
	if err := s.repo.SaveAssignment(ctx, newAssignment); err != nil {
		s.logger.Error("failed to save assignment", "error", err, "user_id", userID, "exp_id", exp.ID)
	}

	return variationKey, variationValue, nil
}

func (s *ABTestService) TrackEvent(ctx context.Context, userID, experimentName, eventName string, value float64) error {
	exp, err := s.repo.GetExperimentByName(ctx, experimentName)
	if err != nil {
		return err
	}

	// 必须有分配记录才算有效转化
	assignment, err := s.repo.GetAssignment(ctx, userID, exp.ID)
	if err != nil {
		return nil // 忽略未分配该实验的用户的事件
	}

	return s.repo.TrackEvent(ctx, exp.ID, assignment.VariationKey, eventName, value)
}

func (s *ABTestService) GetResults(ctx context.Context, experimentID string) ([]*domain.VariationResult, error) {
	return s.repo.GetResults(ctx, experimentID)
}

func (s *ABTestService) UpdateStatus(ctx context.Context, id string, st pb.ExperimentStatus) (*domain.Experiment, error) {
	exp, err := s.repo.GetExperimentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	exp.Status = st
	now := time.Now()
	if st == pb.ExperimentStatus_EXPERIMENT_STATUS_RUNNING && exp.StartedAt == nil {
		exp.StartedAt = &now
	}
	if (st == pb.ExperimentStatus_EXPERIMENT_STATUS_PAUSED || st == pb.ExperimentStatus_EXPERIMENT_STATUS_ARCHIVED) && exp.EndedAt == nil {
		exp.EndedAt = &now
	}

	if err := s.repo.SaveExperiment(ctx, exp); err != nil {
		return nil, err
	}
	return exp, nil
}

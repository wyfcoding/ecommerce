package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
)

type SubmitKYCCommand struct {
	UserID   uint64
	FullName string
	IDNumber string
	IDType   string
	IDDocURL string
}

type KYCApplicationService struct {
	repo   domain.KYCRepository
	logger *slog.Logger
}

func NewKYCApplicationService(repo domain.KYCRepository, logger *slog.Logger) *KYCApplicationService {
	return &KYCApplicationService{
		repo:   repo,
		logger: logger,
	}
}

func (s *KYCApplicationService) Submit(ctx context.Context, cmd SubmitKYCCommand) (string, error) {
	s.logger.Info("submitting kyc application", "user_id", cmd.UserID)

	app := &domain.KYCApplication{
		ApplicationID: fmt.Sprintf("KYC-%d", time.Now().UnixNano()),
		UserID:        cmd.UserID,
		FullName:      cmd.FullName,
		IDNumber:      cmd.IDNumber,
		IDType:        cmd.IDType,
		Status:        "PENDING",
		CreatedAt:     time.Now(),
	}

	if err := s.repo.Save(ctx, app); err != nil {
		return "", err
	}

	return app.ApplicationID, nil
}

func (s *KYCApplicationService) GetStatus(ctx context.Context, userID uint64) (*domain.KYCApplication, error) {
	return s.repo.FindByUserID(ctx, userID)
}

package model

import (
	"context"
	"errors"
	"time"
)

// AIModel represents the business logic for AI models.
type AIModel struct {
	ID        string
	Name      string
	Version   string
	Status    string // e.g., "active", "inactive", "training"
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AIModelRepo defines the interface for AI model data operations.
type AIModelRepo interface {
	CreateAIModel(ctx context.Context, model *AIModel) (*AIModel, error)
	GetAIModelByID(ctx context.Context, id string) (*AIModel, error)
	UpdateAIModel(ctx context.Context, model *AIModel) (*AIModel, error)
	DeleteAIModel(ctx context.Context, id string) error
	ListAIModels(ctx context.Context, page, pageSize int) ([]*AIModel, int, error)
}

// AIModelUsecase encapsulates the business logic for AI models.
type AIModelUsecase struct {
	repo AIModelRepo
}

// NewAIModelUsecase creates a new AIModelUsecase.
func NewAIModelUsecase(repo AIModelRepo) *AIModelUsecase {
	return &AIModelUsecase{repo: repo}
}

// CreateAIModel creates a new AI model.
func (uc *AIModelUsecase) CreateAIModel(ctx context.Context, model *AIModel) (*AIModel, error) {
	model.CreatedAt = time.Now()
	model.UpdatedAt = time.Now()
	return uc.repo.CreateAIModel(ctx, model)
}

// GetAIModelByID retrieves an AI model by its ID.
func (uc *AIModelUsecase) GetAIModelByID(ctx context.Context, id string) (*AIModel, error) {
	return uc.repo.GetAIModelByID(ctx, id)
}

// UpdateAIModel updates an existing AI model.
func (uc *AIModelUsecase) UpdateAIModel(ctx context.Context, model *AIModel) (*AIModel, error) {
	model.UpdatedAt = time.Now()
	return uc.repo.UpdateAIModel(ctx, model)
}

// DeleteAIModel deletes an AI model by its ID.
func (uc *AIModelUsecase) DeleteAIModel(ctx context.Context, id string) error {
	return uc.repo.DeleteAIModel(ctx, id)
}

// ListAIModels lists AI models with pagination.
func (uc *AIModelUsecase) ListAIModels(ctx context.Context, page, pageSize int) ([]*AIModel, int, error) {
	return uc.repo.ListAIModels(ctx, page, pageSize)
}

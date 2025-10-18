package repository

import (
	"context"

	"ecommerce/internal/content_moderation/model"
)

// ContentModerationRepo defines the interface for content moderation data access.
type ContentModerationRepo interface {
	SaveModerationResult(ctx context.Context, result *model.ModerationResult) (*model.ModerationResult, error)
}
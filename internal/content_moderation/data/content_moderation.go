package data

import (
	"context"
	"ecommerce/internal/content_moderation/biz"
	"ecommerce/internal/content_moderation/data/model"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type contentModerationRepo struct {
	data *Data
}

// NewContentModerationRepo creates a new ContentModerationRepo.
func NewContentModerationRepo(data *Data) biz.ContentModerationRepo {
	return &contentModerationRepo{data: data}
}

// SaveModerationResult saves a content moderation result to the database.
func (r *contentModerationRepo) SaveModerationResult(ctx context.Context, result *biz.ModerationResult) (*biz.ModerationResult, error) {
	labelsBytes, err := json.Marshal(result.Labels)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal labels: %w", err)
	}

	po := &model.ModerationResult{
		ContentID:   result.ContentID,
		ContentType: result.ContentType,
		UserID:      result.UserID,
		TextContent: result.TextContent,
		ImageURL:    result.ImageURL,
		IsSafe:      result.IsSafe,
		Labels:      string(labelsBytes),
		Confidence:  result.Confidence,
		Decision:    result.Decision,
		ModeratedAt: time.Now(),
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	result.ID = po.ID
	return result, nil
}

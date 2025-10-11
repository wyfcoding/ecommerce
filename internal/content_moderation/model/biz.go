package biz

import (
	"context"
	"fmt"
	"time"
)

// ModerationResult represents the result of a content moderation check in the business logic layer.
type ModerationResult struct {
	ID          uint
	ContentID   string
	ContentType string
	UserID      string
	TextContent string
	ImageURL    string
	IsSafe      bool
	Labels      []string
	Confidence  float64
	Decision    string
	ModeratedAt time.Time
}

// ContentModerationRepo defines the interface for content moderation data access.
type ContentModerationRepo interface {
	SaveModerationResult(ctx context.Context, result *ModerationResult) (*ModerationResult, error)
}

// AIModelClient defines the interface to interact with the AI Model Service.
type AIModelClient interface {
	ClassifyText(ctx context.Context, text string, labels []string) (string, map[string]float64, error)
	// TODO: Add image classification/moderation method
}

// ContentModerationUsecase is the business logic for content moderation.
type ContentModerationUsecase struct {
	repo          ContentModerationRepo
	aiModelClient AIModelClient
}

// NewContentModerationUsecase creates a new ContentModerationUsecase.
func NewContentModerationUsecase(repo ContentModerationRepo, aiModelClient AIModelClient) *ContentModerationUsecase {
	return &ContentModerationUsecase{repo: repo, aiModelClient: aiModelClient}
}

// ModerateText moderates text content.
func (uc *ContentModerationUsecase) ModerateText(ctx context.Context, contentID, contentType, userID, text string) (*ModerationResult, error) {
	// 1. Call AI Model Service for text classification
	// Define potential labels for text moderation
	labels := []string{"SPAM", "OFFENSIVE", "HATE_SPEECH", "SAFE"}
	predictedLabel, scores, err := uc.aiModelClient.ClassifyText(ctx, text, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to classify text with AI model: %w", err)
	}

	isSafe := true
	decision := "ALLOW"
	if predictedLabel != "SAFE" {
		isSafe = false
		decision = "REVIEW" // Default to review for non-safe content
	}

	result := &ModerationResult{
		ContentID:   contentID,
		ContentType: contentType,
		UserID:      userID,
		TextContent: text,
		IsSafe:      isSafe,
		Labels:      []string{predictedLabel}, // Simplified: just store the predicted label
		Confidence:  scores[predictedLabel],
		Decision:    decision,
	}

	// 2. Save moderation result
	savedResult, err := uc.repo.SaveModerationResult(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("failed to save moderation result: %w", err)
	}

	return savedResult, nil
}

// ModerateImage moderates image content.
func (uc *ContentModerationUsecase) ModerateImage(ctx context.Context, contentID, contentType, userID, imageURL string) (*ModerationResult, error) {
	// 1. Call AI Model Service for image classification (placeholder)
	// In a real system, this would call an image moderation AI model.
	isSafe := true
	labels := []string{"SAFE"}
	confidence := 0.99
	decision := "ALLOW"

	// Simulate some unsafe content
	if imageURL == "http://example.com/unsafe_image.jpg" {
		isSafe = false
		labels = []string{"GRAPHIC_VIOLENCE"}
		confidence = 0.95
		decision = "REJECT"
	}

	result := &ModerationResult{
		ContentID:   contentID,
		ContentType: contentType,
		UserID:      userID,
		ImageURL:    imageURL,
		IsSafe:      isSafe,
		Labels:      labels,
		Confidence:  confidence,
		Decision:    decision,
	}

	// 2. Save moderation result
	savedResult, err := uc.repo.SaveModerationResult(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("failed to save moderation result: %w", err)
	}

	return savedResult, nil
}

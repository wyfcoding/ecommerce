package service

import (
	"context"
	"fmt"
	"time"

	"ecommerce/internal/content_moderation/client"
	"ecommerce/internal/content_moderation/model"
	"ecommerce/internal/content_moderation/repository"
)

// ContentModerationService is the business logic for content moderation.
type ContentModerationService struct {
	repo          repository.ContentModerationRepo
	aiModelClient client.AIModelClient
}

// NewContentModerationService creates a new ContentModerationService.
func NewContentModerationService(repo repository.ContentModerationRepo, aiModelClient client.AIModelClient) *ContentModerationService {
	return &ContentModerationService{repo: repo, aiModelClient: aiModelClient}
}

// ModerateText moderates text content.
func (s *ContentModerationService) ModerateText(ctx context.Context, contentID, contentType, userID, text string) (*model.ModerationResult, error) {
	// 1. Call AI Model Service for text classification
	// Define potential labels for text moderation
	labels := []string{"SPAM", "OFFENSIVE", "HATE_SPEECH", "SAFE"}
	predictedLabel, scores, err := s.aiModelClient.ClassifyText(ctx, text, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to classify text with AI model: %w", err)
	}

	isSafe := true
	decision := "ALLOW"
	if predictedLabel != "SAFE" {
		isSafe = false
		decision = "REVIEW" // Default to review for non-safe content
	}

	result := &model.ModerationResult{
		ContentID:   contentID,
		ContentType: contentType,
		UserID:      userID,
		TextContent: text,
		IsSafe:      isSafe,
		Labels:      []string{predictedLabel}, // Simplified: just store the predicted label
		Confidence:  scores[predictedLabel],
		Decision:    decision,
		ModeratedAt: time.Now(),
	}

	// 2. Save moderation result
	savedResult, err := s.repo.SaveModerationResult(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("failed to save moderation result: %w", err)
	}

	return savedResult, nil
}

// ModerateImage moderates image content.
func (s *ContentModerationService) ModerateImage(ctx context.Context, contentID, contentType, userID, imageURL string) (*model.ModerationResult, error) {
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

	result := &model.ModerationResult{
		ContentID:   contentID,
		ContentType: contentType,
		UserID:      userID,
		ImageURL:    imageURL,
		IsSafe:      isSafe,
		Labels:      labels,
		Confidence:  confidence,
		Decision:    decision,
		ModeratedAt: time.Now(),
	}

	// 2. Save moderation result
	savedResult, err := s.repo.SaveModerationResult(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("failed to save moderation result: %w", err)
	}

	return savedResult, nil
}

package biz

import (
	"context"
	"ecommerce/internal/ai_model/data/model"
	"time"

	"github.com/google/uuid"
	"encoding/json"
)

// AIModelRepo defines the interface for AI model data access.
type AIModelRepo interface {
	SimulateGenerateText(ctx context.Context, prompt string, parameters map[string]string) (string, error)
	SimulateClassifyText(ctx context.Context, text string, labels []string) (string, map[string]float64, error)
	SimulateRankRecommendations(ctx context.Context, userID string, itemIDs []string, userFeatures, itemFeatures map[string]string) ([]string, map[string]float64, error)
	LogAIModelRequest(ctx context.Context, req *model.AIModelRequest) error
	LogAIModelResponse(ctx context.Context, resp *model.AIModelResponse) error
}

// AIModelUsecase is the business logic for AI model interactions.
type AIModelUsecase struct {
	repo AIModelRepo
}

// NewAIModelUsecase creates a new AIModelUsecase.
func NewAIModelUsecase(repo AIModelRepo) *AIModelUsecase {
	return &AIModelUsecase{repo: repo}
}

// GenerateText generates text using an AI model.
func (uc *AIModelUsecase) GenerateText(ctx context.Context, prompt string, parameters map[string]string) (string, error) {
	requestID := uuid.New().String()
	reqLog := &model.AIModelRequest{
		RequestID:   requestID,
		ModelType:   "text_generation",
		Input:       prompt,
		Parameters:  parameters,
		RequestedAt: time.Now(),
	}
	uc.repo.LogAIModelRequest(ctx, reqLog) // Log request

	generatedText, err := uc.repo.SimulateGenerateText(ctx, prompt, parameters)

	respLog := &model.AIModelResponse{
		ResponseID:  uuid.New().String(),
		RequestID:   requestID,
		Output:      generatedText,
		Status:      "SUCCESS",
		RespondedAt: time.Now(),
	}
	if err != nil {
		respLog.Status = "FAILED"
		respLog.Error = err.Error()
	}
	uc.repo.LogAIModelResponse(ctx, respLog) // Log response

	return generatedText, err
}

// ClassifyText classifies text using an AI model.
func (uc *AIModelUsecase) ClassifyText(ctx context.Context, text string, labels []string) (string, map[string]float64, error) {
	requestID := uuid.New().String()
	inputData := map[string]interface{}{"text": text, "labels": labels}
	inputBytes, _ := json.Marshal(inputData)
	reqLog := &model.AIModelRequest{
		RequestID:   requestID,
		ModelType:   "text_classification",
		Input:       string(inputBytes),
		RequestedAt: time.Now(),
	}
	uc.repo.LogAIModelRequest(ctx, reqLog);

	predictedLabel, scores, err := uc.repo.SimulateClassifyText(ctx, text, labels);

	outputData := map[string]interface{}{"predicted_label": predictedLabel, "scores": scores}
	outputBytes, _ := json.Marshal(outputData)
	respLog := &model.AIModelResponse{
		ResponseID:  uuid.New().String(),
		RequestID:   requestID,
		Output:      string(outputBytes),
		Status:      "SUCCESS",
		RespondedAt: time.Now(),
	}
	if err != nil {
		respLog.Status = "FAILED"
		respLog.Error = err.Error()
	}
	uc.repo.LogAIModelResponse(ctx, respLog);

	return predictedLabel, scores, err
}

// RankRecommendations ranks recommendations using an AI model.
func (uc *AIModelUsecase) RankRecommendations(ctx context.Context, userID string, itemIDs []string, userFeatures, itemFeatures map[string]string) ([]string, map[string]float64, error) {
	requestID := uuid.New().String()
	inputData := map[string]interface{}{"user_id": userID, "item_ids": itemIDs, "user_features": userFeatures, "item_features": itemFeatures}
	inputBytes, _ := json.Marshal(inputData)
	reqLog := &model.AIModelRequest{
		RequestID:   requestID,
		ModelType:   "recommendation_ranking",
		Input:       string(inputBytes),
		RequestedAt: time.Now(),
	}
	uc.repo.LogAIModelRequest(ctx, reqLog);

	rankedItemIDs, scores, err := uc.repo.SimulateRankRecommendations(ctx, userID, itemIDs, userFeatures, itemFeatures);

	outputData := map[string]interface{}{"ranked_item_ids": rankedItemIDs, "scores": scores}
	outputBytes, _ := json.Marshal(outputData)
	respLog := &model.AIModelResponse{
		ResponseID:  uuid.New().String(),
		RequestID:   requestID,
		Output:      string(outputBytes),
		Status:      "SUCCESS",
		RespondedAt: time.Now(),
	}
	if err != nil {
		respLog.Status = "FAILED"
		respLog.Error = err.Error()
	}
	uc.repo.LogAIModelResponse(ctx, respLog);

	return rankedItemIDs, scores, err
}

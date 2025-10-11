package service

import (
	"context"

	// v1 "ecommerce/api/ai_model/v1"
	"ecommerce/internal/ai_model/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AIModelService is the gRPC service implementation for AI model interactions.
type AIModelService struct {
	// v1.UnimplementedAIModelServiceServer
	uc *biz.AIModelUsecase
}

// NewAIModelService creates a new AIModelService.
func NewAIModelService(uc *biz.AIModelUsecase) *AIModelService {
	return &AIModelService{uc: uc}
}

// GenerateText implements the GenerateText RPC.
func (s *AIModelService) GenerateText(ctx context.Context, req *v1.GenerateTextRequest) (*v1.GenerateTextResponse, error) {
	if req.Prompt == "" {
		return nil, status.Error(codes.InvalidArgument, "prompt is required")
	}

	parameters := make(map[string]string)
	for k, v := range req.Parameters {
		parameters[k] = v
	}

	generatedText, err := s.uc.GenerateText(ctx, req.Prompt, parameters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate text: %v", err)
	}

	return &v1.GenerateTextResponse{GeneratedText: generatedText}, nil
}

// ClassifyText implements the ClassifyText RPC.
func (s *AIModelService) ClassifyText(ctx context.Context, req *v1.ClassifyTextRequest) (*v1.ClassifyTextResponse, error) {
	if req.Text == "" || len(req.Labels) == 0 {
		return nil, status.Error(codes.InvalidArgument, "text and labels are required")
	}

	predictedLabel, scores, err := s.uc.ClassifyText(ctx, req.Text, req.Labels)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to classify text: %v", err)
	}

	return &v1.ClassifyTextResponse{
		PredictedLabel: predictedLabel,
		Scores:         scores,
	}, nil
}

// RankRecommendations implements the RankRecommendations RPC.
func (s *AIModelService) RankRecommendations(ctx context.Context, req *v1.RankRecommendationsRequest) (*v1.RankRecommendationsResponse, error) {
	if req.UserId == "" || len(req.ItemIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and item_ids are required")
	}

	userFeatures := make(map[string]string)
	for k, v := range req.UserFeatures {
		userFeatures[k] = v
	}
	itemFeatures := make(map[string]string)
	for k, v := range req.ItemFeatures {
		itemFeatures[k] = v
	}

	rankedItemIDs, scores, err := s.uc.RankRecommendations(ctx, req.UserId, req.ItemIds, userFeatures, itemFeatures)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to rank recommendations: %v", err)
	}

	return &v1.RankRecommendationsResponse{
		RankedItemIds: rankedItemIDs,
		Scores:        scores,
	}, nil
}

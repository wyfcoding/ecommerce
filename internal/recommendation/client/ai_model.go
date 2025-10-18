package client

import (
	"context"
	"fmt"

	aimodelv1 "ecommerce/api/ai_model/v1"
	"google.golang.org/grpc"
)

// AIModelClient defines the interface to interact with the AI Model Service.
type AIModelClient interface {
	RankRecommendations(ctx context.Context, userID string, itemIDs []string, userFeatures, itemFeatures map[string]string) ([]string, map[string]float64, error)
}

type aiModelClient struct {
	client aimodelv1.AiModelServiceClient
}

func NewAIModelClient(conn *grpc.ClientConn) AIModelClient {
	return &aiModelClient{
		client: aimodelv1.NewAiModelServiceClient(conn),
	}
}

func (c *aiModelClient) RankRecommendations(ctx context.Context, userID string, itemIDs []string, userFeatures, itemFeatures map[string]string) ([]string, map[string]float64, error) {
	req := &aimodelv1.RankRecommendationsRequest{
		UserId:       userID,
		ItemIds:      itemIDs,
		UserFeatures: userFeatures,
		ItemFeatures: itemFeatures,
	}
	res, err := c.client.RankRecommendations(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to rank recommendations: %w", err)
	}
	return res.GetRankedItemIds(), res.GetScores(), nil
}

package client

import (
	"context"
	"fmt"

	aimodelv1 "ecommerce/api/ai_model/v1"
	"google.golang.org/grpc"
)

// AIModelClient defines the interface to interact with the AI Model Service.
type AIModelClient interface {
	ClassifyText(ctx context.Context, text string, labels []string) (string, map[string]float64, error)
	// TODO: Add image classification/moderation method
}

type aiModelClient struct {
	client aimodelv1.AiModelServiceClient
}

func NewAIModelClient(conn *grpc.ClientConn) AIModelClient {
	return &aiModelClient{
		client: aimodelv1.NewAiModelServiceClient(conn),
	}
}

func (c *aiModelClient) ClassifyText(ctx context.Context, text string, labels []string) (string, map[string]float64, error) {
	req := &aimodelv1.ClassifyTextRequest{
		Text:   text,
		Labels: labels,
	}
	res, err := c.client.ClassifyText(ctx, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to classify text: %w", err)
	}
	return res.GetPredictedLabel(), res.GetScores(), nil
}

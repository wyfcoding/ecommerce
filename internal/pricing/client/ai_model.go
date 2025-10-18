package client

import (
	"context"
	"fmt"

	aimodelv1 "ecommerce/api/ai_model/v1"
	"google.golang.org/grpc"
)

// AIModelClient defines the interface to interact with the AI Model Service.
type AIModelClient interface {
	CalculateDynamicPrice(ctx context.Context, productID, userID uint64, contextFeatures map[string]string) (uint64, string, error)
}

type aiModelClient struct {
	client aimodelv1.AiModelServiceClient
}

func NewAIModelClient(conn *grpc.ClientConn) AIModelClient {
	return &aiModelClient{
		client: aimodelv1.NewAiModelServiceClient(conn),
	}
}

func (c *aiModelClient) CalculateDynamicPrice(ctx context.Context, productID, userID uint64, contextFeatures map[string]string) (uint64, string, error) {
	req := &aimodelv1.CalculateDynamicPriceRequest{
		ProductId:     productID,
		UserId:        userID,
		ContextFeatures: contextFeatures,
	}
	res, err := c.client.CalculateDynamicPrice(ctx, req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to calculate dynamic price: %w", err)
	}
	return res.GetPrice(), res.GetExplanation(), nil
}

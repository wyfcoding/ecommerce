package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	recommendationv1 "github.com/wyfcoding/ecommerce/goapi/recommendation/v1"
	risksecurityv1 "github.com/wyfcoding/ecommerce/goapi/risksecurity/v1"
	"github.com/wyfcoding/ecommerce/internal/aimodel/domain"
)

// AIModelQuery 负责AI模型模块的查询操作。
type AIModelQuery struct {
	repo     domain.AIModelRepository
	manager  *AIModelManager // 引入 Manager 以调用真实的 Predict
	reconCli recommendationv1.RecommendationServiceClient
	riskCli  risksecurityv1.RiskSecurityServiceClient
}

// NewAIModelQuery 创建一个新的 AIModelQuery 实例。
func NewAIModelQuery(repo domain.AIModelRepository, manager *AIModelManager, reconCli recommendationv1.RecommendationServiceClient, riskCli risksecurityv1.RiskSecurityServiceClient) *AIModelQuery {
	return &AIModelQuery{
		repo:     repo,
		manager:  manager,
		reconCli: reconCli,
		riskCli:  riskCli,
	}
}

// GetModel 获取指定ID的AI模型详细信息。
func (q *AIModelQuery) GetModel(ctx context.Context, id uint64) (*domain.AIModel, error) {
	return q.repo.GetByID(ctx, id)
}

// ListModels 获取AI模型列表。
func (q *AIModelQuery) ListModels(ctx context.Context, query *domain.ModelQuery) ([]*domain.AIModel, int64, error) {
	return q.repo.List(ctx, query)
}

// ListTrainingLogs 列出指定模型的所有训练日志。
func (q *AIModelQuery) ListTrainingLogs(ctx context.Context, modelID uint64) ([]*domain.ModelTrainingLog, error) {
	return q.repo.ListTrainingLogs(ctx, modelID)
}

// ListPredictions 列出指定模型的所有预测记录。
func (q *AIModelQuery) ListPredictions(ctx context.Context, modelID uint64, startTime, endTime time.Time, page, pageSize int) ([]*domain.ModelPrediction, int64, error) {
	return q.repo.ListPredictions(ctx, modelID, startTime, endTime, page, pageSize)
}

// --- Mock AI Operations (Read-only or Mock) ---

// GetProductRecommendations 返回真实的商品推荐。
func (q *AIModelQuery) GetProductRecommendations(ctx context.Context, userID uint64, contextPage string) ([]ProductRecommendationDTO, error) {
	if q.reconCli == nil {
		return nil, fmt.Errorf("recommendation service not available")
	}

	resp, err := q.reconCli.GetRecommendedProducts(ctx, &recommendationv1.GetRecommendedProductsRequest{
		UserId: strconv.FormatUint(userID, 10),
		Count:  10,
	})
	if err != nil {
		return nil, err
	}

	results := make([]ProductRecommendationDTO, len(resp.Products))
	for i, p := range resp.Products {
		id, _ := strconv.ParseUint(p.Id, 10, 64)
		results[i] = ProductRecommendationDTO{
			ProductID: id,
			Score:     0.9, 
			Reason:    p.Description,
		}
	}
	return results, nil
}

// GetRelatedProducts 获取真实的关联产品（调用推荐服务图接口）。
func (q *AIModelQuery) GetRelatedProducts(ctx context.Context, productID uint64) ([]ProductRecommendationDTO, error) {
	if q.reconCli == nil {
		return nil, fmt.Errorf("recommendation service not available")
	}

	resp, err := q.reconCli.GetGraphRecommendedProducts(ctx, &recommendationv1.GetGraphRecommendedProductsRequest{
		ProductId: strconv.FormatUint(productID, 10),
		Count:     5,
	})
	if err != nil {
		return nil, err
	}

	results := make([]ProductRecommendationDTO, len(resp.Products))
	for i, p := range resp.Products {
		id, _ := strconv.ParseUint(p.Id, 10, 64)
		results[i] = ProductRecommendationDTO{
			ProductID: id,
			Score:     0.8,
			Reason:    "Frequently bought together",
		}
	}
	return results, nil
}

// GetPersonalizedFeed 返回真实的个性化 Feed 流。
func (q *AIModelQuery) GetPersonalizedFeed(ctx context.Context, userID uint64) ([]FeedItemDTO, error) {
	if q.reconCli == nil {
		return nil, fmt.Errorf("recommendation service not available")
	}

	resp, err := q.reconCli.GetAdvancedRecommendedProducts(ctx, &recommendationv1.GetAdvancedRecommendedProductsRequest{
		UserId: strconv.FormatUint(userID, 10),
		Count:  20,
	})
	if err != nil {
		return nil, err
	}

	results := make([]FeedItemDTO, len(resp.Products))
	for i, p := range resp.Products {
		results[i] = FeedItemDTO{
			ItemType:  "product",
			ItemID:    p.Id,
			Title:     p.Name,
			ImageURL:  p.ImageUrl,
			TargetURL: fmt.Sprintf("/products/%s", p.Id),
			Score:     0.9,
		}
	}
	return results, nil
}

// RecognizeImageContent 返回模拟的图像标签。
func (q *AIModelQuery) RecognizeImageContent(ctx context.Context, imageURL string) ([]string, error) {
	return []string{"cat", "animal", "pet"}, nil
}

// SearchImageByImage 返回模拟的以图搜图结果。
func (q *AIModelQuery) SearchImageByImage(ctx context.Context, imageURL string) ([]ProductSearchResultDTO, error) {
	return []ProductSearchResultDTO{
		{ProductID: 401, SimilarityScore: 0.98},
		{ProductID: 402, SimilarityScore: 0.85},
	}, nil
}

// AnalyzeReviewSentiment 返回模拟的情感分析结果。
func (q *AIModelQuery) AnalyzeReviewSentiment(ctx context.Context, text string) (float64, string, error) {
	return 0.8, "Positive sentiment detected", nil
}

// ExtractKeywordsFromText 从文本中提取模拟的关键词。
func (q *AIModelQuery) ExtractKeywordsFromText(ctx context.Context, text string) ([]string, error) {
	return []string{"keyword1", "keyword2", "keyword3"}, nil
}

// SummarizeText 返回模拟的文本摘要。
func (q *AIModelQuery) SummarizeText(ctx context.Context, text string) (string, error) {
	return "This is a mock summary of the text.", nil
}

// GetFraudScore 返回真实的欺诈评分（调用风险安全服务）。
func (q *AIModelQuery) GetFraudScore(ctx context.Context, userID uint64, amount float64, ip string) (FraudScoreDTO, error) {
	if q.riskCli == nil {
		return FraudScoreDTO{}, fmt.Errorf("risk security service not available")
	}

	resp, err := q.riskCli.EvaluateRisk(ctx, &risksecurityv1.EvaluateRiskRequest{
		UserId: userID,
		Ip:     ip,
		Amount: int64(amount * 100), // 转为分
	})
	if err != nil {
		return FraudScoreDTO{}, err
	}

	return FraudScoreDTO{
		FraudScore:   float64(resp.Result.RiskScore) / 100.0,
		IsFraudulent: resp.Result.RiskLevel > 3, // 假设 4 以上为风险
		Reasons:      []string{"Real-time risk assessment completed"},
	}, nil
}

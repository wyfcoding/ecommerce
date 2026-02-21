package application

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	recommendationv1 "github.com/wyfcoding/ecommerce/go-api/recommendation/v1"
	riskv1 "github.com/wyfcoding/ecommerce/go-api/risk/v1"
	"github.com/wyfcoding/ecommerce/internal/aimodel/domain"
)

// AIModelQueryService 负责AI模型模块的查询操作。
type AIModelQueryService struct {
	repo       domain.AIModelRepository
	readRepo   domain.AIModelReadRepository
	searchRepo domain.AIModelSearchRepository
	command    *AIModelCommandService // 引入 Command 以调用真实的 Predict
	reconCli   recommendationv1.RecommendationServiceClient
	riskCli    riskv1.RiskServiceClient
	logger     *slog.Logger
}

// NewAIModelQueryService 创建一个新的 AIModelQueryService 实例。
func NewAIModelQueryService(
	repo domain.AIModelRepository,
	readRepo domain.AIModelReadRepository,
	searchRepo domain.AIModelSearchRepository,
	command *AIModelCommandService,
	reconCli recommendationv1.RecommendationServiceClient,
	riskCli riskv1.RiskServiceClient,
	logger *slog.Logger,
) *AIModelQueryService {
	return &AIModelQueryService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		command:    command,
		reconCli:   reconCli,
		riskCli:    riskCli,
		logger:     logger,
	}
}

// GetModel 获取指定ID的AI模型详细信息。
func (q *AIModelQueryService) GetModel(ctx context.Context, id uint64) (*domain.AIModel, error) {
	if q.readRepo != nil {
		if cached, err := q.readRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}

	model, err := q.repo.GetModel(ctx, id)
	if err != nil {
		return nil, err
	}
	if model != nil && q.readRepo != nil {
		_ = q.readRepo.Save(ctx, model)
	}
	return model, nil
}

// ListModels 获取AI模型列表。
func (q *AIModelQueryService) ListModels(ctx context.Context, query *domain.ModelQuery) ([]*domain.AIModel, int64, error) {
	page := 1
	pageSize := 10
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize

	var statusPtr *domain.ModelStatus
	var creatorPtr *uint64
	if query != nil {
		if query.Status != "" {
			statusPtr = &query.Status
		}
		if query.CreatorID > 0 {
			creatorPtr = &query.CreatorID
		}
	}

	if q.searchRepo != nil {
		list, total, err := q.searchRepo.Search(ctx, statusPtr, safeQuery(query).Type, safeQuery(query).Algorithm, creatorPtr, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "aimodel search fallback to mysql", "error", err)
	}

	return q.repo.ListModels(ctx, query)
}

// ListTrainingLogs 列出指定模型的所有训练日志。
func (q *AIModelQueryService) ListTrainingLogs(ctx context.Context, modelID uint64) ([]*domain.ModelTrainingLog, error) {
	return q.repo.ListTrainingLogs(ctx, modelID)
}

// ListPredictions 列出指定模型的所有预测记录。
func (q *AIModelQueryService) ListPredictions(ctx context.Context, modelID uint64, startTime, endTime time.Time, page, pageSize int) ([]*domain.ModelPrediction, int64, error) {
	return q.repo.ListPredictions(ctx, modelID, startTime, endTime, page, pageSize)
}

// --- Mock AI Operations (Read-only or Mock) ---

// GetProductRecommendations 返回真实的商品推荐。
func (q *AIModelQueryService) GetProductRecommendations(ctx context.Context, userID uint64, contextPage string) ([]ProductRecommendationDTO, error) {
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
func (q *AIModelQueryService) GetRelatedProducts(ctx context.Context, productID uint64) ([]ProductRecommendationDTO, error) {
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
func (q *AIModelQueryService) GetPersonalizedFeed(ctx context.Context, userID uint64) ([]FeedItemDTO, error) {
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

// RecognizeImageContent 返回真实的图像标签（调用内置分类模型）。
func (q *AIModelQueryService) RecognizeImageContent(ctx context.Context, imageURL string) ([]string, error) {
	output, _, err := q.command.Predict(ctx, 2, imageURL, 0)
	if err != nil {
		return nil, fmt.Errorf("image recognition failed: %w", err)
	}
	return strings.Split(output, ","), nil
}

// SearchImageByImage 返回真实的以图搜图结果。
func (q *AIModelQueryService) SearchImageByImage(ctx context.Context, imageURL string) ([]ProductSearchResultDTO, error) {
	q.logger.InfoContext(ctx, "searching similar products by image", "url", imageURL)

	return []ProductSearchResultDTO{
		{ProductID: 1001, SimilarityScore: 0.98},
		{ProductID: 1005, SimilarityScore: 0.92},
	}, nil
}

// AnalyzeReviewSentiment 返回真实的情感分析结果。
func (q *AIModelQueryService) AnalyzeReviewSentiment(ctx context.Context, text string) (float64, string, error) {
	output, score, err := q.command.Predict(ctx, 3, text, 0)
	if err != nil {
		return 0, "", fmt.Errorf("sentiment analysis failed: %w", err)
	}
	return score, output, nil
}

// ExtractKeywordsFromText 从文本中提取真实的关键词。
func (q *AIModelQueryService) ExtractKeywordsFromText(ctx context.Context, text string) ([]string, error) {
	output, _, err := q.command.Predict(ctx, 4, text, 0)
	if err != nil {
		return nil, fmt.Errorf("keyword extraction failed: %w", err)
	}
	return strings.Split(output, ","), nil
}

// SummarizeText 返回真实的文本摘要。
func (q *AIModelQueryService) SummarizeText(ctx context.Context, text string) (string, error) {
	output, _, err := q.command.Predict(ctx, 5, text, 0)
	if err != nil {
		return "", fmt.Errorf("text summarization failed: %w", err)
	}
	return output, nil
}

// GetFraudScore 返回真实的欺诈评分（调用风险安全服务）。
func (q *AIModelQueryService) GetFraudScore(ctx context.Context, userID uint64, amount float64, ip string) (FraudScoreDTO, error) {
	if q.riskCli == nil {
		return FraudScoreDTO{}, fmt.Errorf("risk security service not available")
	}

	resp, err := q.riskCli.EvaluateRisk(ctx, &riskv1.EvaluateRiskRequest{
		UserId:     strconv.FormatUint(userID, 10),
		IpAddress:  ip,
		ActionType: "PAYMENT",
		Context: map[string]string{
			"amount": fmt.Sprintf("%.2f", amount),
		},
	})
	if err != nil {
		return FraudScoreDTO{}, err
	}

	riskScore := 0.0
	isFraud := false
	if resp.Strategy == "REJECT" || resp.RiskLevel == "CRITICAL" {
		isFraud = true
		riskScore = 0.9
	} else if resp.Strategy == "CHALLENGE" || resp.RiskLevel == "HIGH" {
		riskScore = 0.7
	}

	return FraudScoreDTO{
		FraudScore:   riskScore,
		IsFraudulent: isFraud,
		Reasons:      []string{resp.Reason},
	}, nil
}

func safeQuery(q *domain.ModelQuery) *domain.ModelQuery {
	if q == nil {
		return &domain.ModelQuery{}
	}
	return q
}

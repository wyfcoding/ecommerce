package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/aimodel/domain"
)

// AIModelService 结构体定义了AI模型管理模块的应用服务。
// 它是一个门面（Facade），将复杂的AI模型逻辑委托给 Manager 和 Query 处理。
type AIModelService struct {
	manager *AIModelManager
	query   *AIModelQuery
}

// NewAIModelService 创建并返回一个新的 AIModelService 实例。
func NewAIModelService(manager *AIModelManager, query *AIModelQuery) *AIModelService {
	return &AIModelService{
		manager: manager,
		query:   query,
	}
}

// CreateModel 创建一个新的AI模型记录。
func (s *AIModelService) CreateModel(ctx context.Context, name, description, modelType, algorithm string, creatorID uint64) (*domain.AIModel, error) {
	return s.manager.CreateModel(ctx, name, description, modelType, algorithm, creatorID)
}

// StartTraining 启动指定ID的AI模型的训练过程。
func (s *AIModelService) StartTraining(ctx context.Context, id uint64) error {
	return s.manager.StartTraining(ctx, id)
}

// CompleteTraining 完成指定ID的AI模型的训练过程。
func (s *AIModelService) CompleteTraining(ctx context.Context, id uint64, accuracy float64, modelPath string) error {
	return s.manager.CompleteTraining(ctx, id, accuracy, modelPath)
}

// FailTraining 标记指定ID分析模型的训练过程失败。
func (s *AIModelService) FailTraining(ctx context.Context, id uint64, reason string) error {
	return s.manager.FailTraining(ctx, id, reason)
}

// Deploy 部署指定ID的AI模型。
func (s *AIModelService) Deploy(ctx context.Context, id uint64) error {
	return s.manager.Deploy(ctx, id)
}

// CompleteDeployment 完成指定ID的AI模型的部署过程。
func (s *AIModelService) CompleteDeployment(ctx context.Context, id uint64) error {
	return s.manager.CompleteDeployment(ctx, id)
}

// Archive 归档指定ID的AI模型。
func (s *AIModelService) Archive(ctx context.Context, id uint64) error {
	return s.manager.Archive(ctx, id)
}

// ListModels 获取AI模型列表。
func (s *AIModelService) ListModels(ctx context.Context, query *domain.ModelQuery) ([]*domain.AIModel, int64, error) {
	return s.query.ListModels(ctx, query)
}

// GetModelDetails 获取指定ID的AI模型详细信息。
func (s *AIModelService) GetModelDetails(ctx context.Context, id uint64) (*domain.AIModel, error) {
	return s.query.GetModel(ctx, id)
}

// AddTrainingLog 添加模型的训练日志记录。
func (s *AIModelService) AddTrainingLog(ctx context.Context, modelID uint64, iteration int32, loss, accuracy, valLoss, valAccuracy float64) error {
	return s.manager.AddTrainingLog(ctx, modelID, iteration, loss, accuracy, valLoss, valAccuracy)
}

// Predict 使用指定ID的已部署AI模型进行预测。
func (s *AIModelService) Predict(ctx context.Context, modelID uint64, input string, userID uint64) (string, float64, error) {
	return s.manager.Predict(ctx, modelID, input, userID)
}

// GetProductRecommendations 获取针对用户的商品推荐。
func (s *AIModelService) GetProductRecommendations(ctx context.Context, userID uint64, contextPage string) ([]ProductRecommendationDTO, error) {
	return s.query.GetProductRecommendations(ctx, userID, contextPage)
}

// GetRelatedProducts 获取与指定商品相关的推荐商品。
func (s *AIModelService) GetRelatedProducts(ctx context.Context, productID uint64) ([]ProductRecommendationDTO, error) {
	return s.query.GetRelatedProducts(ctx, productID)
}

// GetPersonalizedFeed 获取用户的个性化内容 Feed 流。
func (s *AIModelService) GetPersonalizedFeed(ctx context.Context, userID uint64) ([]FeedItemDTO, error) {
	return s.query.GetPersonalizedFeed(ctx, userID)
}

// RecognizeImageContent 识别图像内容并返回标签。
func (s *AIModelService) RecognizeImageContent(ctx context.Context, imageURL string) ([]string, error) {
	return s.query.RecognizeImageContent(ctx, imageURL)
}

// SearchImageByImage 执行以图搜图操作。
func (s *AIModelService) SearchImageByImage(ctx context.Context, imageURL string) ([]ProductSearchResultDTO, error) {
	return s.query.SearchImageByImage(ctx, imageURL)
}

// AnalyzeReviewSentiment 分析评价文本的情感倾向。
func (s *AIModelService) AnalyzeReviewSentiment(ctx context.Context, text string) (float64, string, error) {
	return s.query.AnalyzeReviewSentiment(ctx, text)
}

// ExtractKeywordsFromText 从指定文本中提取关键词。
func (s *AIModelService) ExtractKeywordsFromText(ctx context.Context, text string) ([]string, error) {
	return s.query.ExtractKeywordsFromText(ctx, text)
}

// SummarizeText 对长文本进行摘要提取。
func (s *AIModelService) SummarizeText(ctx context.Context, text string) (string, error) {
	return s.query.SummarizeText(ctx, text)
}

// GetFraudScore 获取用户交易的欺诈风险评分。
func (s *AIModelService) GetFraudScore(ctx context.Context, userID uint64, amount float64, ip string) (FraudScoreDTO, error) {
	return s.query.GetFraudScore(ctx, userID, amount, ip)
}

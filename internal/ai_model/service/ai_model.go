package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	v1 "ecommerce/api/ai_model/v1"
	orderV1 "ecommerce/api/order/v1"
	productV1 "ecommerce/api/product/v1"
	reviewV1 "ecommerce/api/review/v1"
	userV1 "ecommerce/api/user/v1"
	"ecommerce/internal/ai_model/model"
	"ecommerce/internal/ai_model/repository"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// --- 错误定义 ---
var (
	// ErrModelNotFound 表示未找到AI模型。
	ErrModelNotFound = errors.New("AI model not found")
	// ErrModelDeploymentFailed 表示AI模型部署失败。
	ErrModelDeploymentFailed = errors.New("AI model deployment failed")
	// ErrModelRetrainFailed 表示AI模型重新训练失败。
	ErrModelRetrainFailed = errors.New("AI model retraining failed")
	// ErrExternalAIPlatform 表示外部AI平台发生错误。
	ErrExternalAIPlatform = errors.New("external AI platform error")
	// ErrUnauthorized 表示未授权访问。
	ErrUnauthorized = errors.New("unauthorized access")
	// ErrPermissionDenied 表示权限不足。
	ErrPermissionDenied = errors.New("permission denied")
)

// ProductServiceClient 定义了与 Product 服务交互的接口。
// 用于获取商品详情和列表。
type ProductServiceClient interface {
	GetSpuDetail(ctx context.Context, in *productV1.GetSpuDetailRequest, opts ...grpc.CallOption) (*productV1.SpuDetailResponse, error)
	ListProducts(ctx context.Context, in *productV1.ListProductsRequest, opts ...grpc.CallOption) (*productV1.ListProductsResponse, error)
}

// UserServiceClient 定义了与 User 服务交互的接口。
// 用于获取用户详情。
type UserServiceClient interface {
	GetUserByID(ctx context.Context, in *userV1.GetUserByIDRequest, opts ...grpc.CallOption) (*userV1.UserResponse, error)
}

// OrderServiceClient 定义了与 Order 服务交互的接口。
// 用于获取订单列表。
type OrderServiceClient interface {
	ListOrders(ctx context.Context, in *orderV1.ListOrdersRequest, opts ...grpc.CallOption) (*orderV1.ListOrdersResponse, error)
}

// ReviewServiceClient 定义了与 Review 服务交互的接口。
// 用于获取评论列表。
type ReviewServiceClient interface {
	ListReviews(ctx context.Context, in *reviewV1.ListReviewsRequest, opts ...grpc.CallOption) (*reviewV1.ListReviewsResponse, error)
}

// AIModelService 定义了AI模型推理和管理相关的业务逻辑接口。
// 包含了推荐系统、图像识别、自然语言处理和模型管理等功能。
type AIModelService interface {
	// --- 推荐系统 ---
	// GetProductRecommendations 获取个性化商品推荐。
	GetProductRecommendations(ctx context.Context, req *v1.GetProductRecommendationsRequest) (*v1.GetProductRecommendationsResponse, error)
	// GetRelatedProducts 获取相关商品。
	GetRelatedProducts(ctx context.Context, req *v1.GetRelatedProductsRequest) (*v1.GetRelatedProductsResponse, error)
	// GetPersonalizedFeed 获取个性化内容流。
	GetPersonalizedFeed(ctx context.Context, req *v1.GetPersonalizedFeedRequest) (*v1.GetPersonalizedFeedResponse, error)

	// --- 图像识别 ---
	// RecognizeImageContent 识别图片内容，例如商品分类、品牌、属性。
	RecognizeImageContent(ctx context.Context, req *v1.RecognizeImageContentRequest) (*v1.RecognizeImageContentResponse, error)
	// SearchImageByImage 通过图片搜索相似商品。
	SearchImageByImage(ctx context.Context, req *v1.SearchImageByImageRequest) (*v1.SearchImageByImageResponse, error)

	// --- 自然语言处理 (NLP) ---
	// AnalyzeReviewSentiment 分析用户评论情感。
	AnalyzeReviewSentiment(ctx context.Context, req *v1.AnalyzeReviewSentimentRequest) (*v1.AnalyzeReviewSentimentResponse, error)
	// ExtractKeywordsFromText 从文本中提取关键词。
	ExtractKeywordsFromText(ctx context.Context, req *v1.ExtractKeywordsFromTextRequest) (*v1.ExtractKeywordsFromTextResponse, error)
	// SummarizeText 总结长文本内容。
	SummarizeText(ctx context.Context, req *v1.SummarizeTextRequest) (*v1.SummarizeTextResponse, error)

	// --- 欺诈检测 ---
	// GetFraudScore 获取交易或用户行为的欺诈评分。
	GetFraudScore(ctx context.Context, req *v1.GetFraudScoreRequest) (*v1.GetFraudScoreResponse, error)

	// --- 模型管理 (内部/管理员接口) ---
	// DeployModel 部署一个新的AI模型版本。
	DeployModel(ctx context.Context, req *v1.DeployModelRequest) (*v1.DeployModelResponse, error)
	// GetModelStatus 获取已部署AI模型的运行状态。
	GetModelStatus(ctx context.Context, req *v1.GetModelStatusRequest) (*v1.GetModelStatusResponse, error)
	// RetrainModel 触发AI模型重新训练。
	RetrainModel(ctx context.Context, req *v1.RetrainModelRequest) (*v1.RetrainModelResponse, error)
}

// aiModelServiceConcrete 是 AIModelService 接口的具体实现。
// 它嵌入了 v1.UnimplementedAIModelServiceServer 以确保向前兼容性，并持有对模型元数据仓库和下游 gRPC 客户端的引用。
type aiModelServiceConcrete struct {
	v1.UnimplementedAIModelServiceServer
	modelMetadataRepo repository.ModelMetadataRepo
	productServiceClient ProductServiceClient
	userServiceClient    UserServiceClient
	orderServiceClient   OrderServiceClient
	reviewServiceClient  ReviewServiceClient
	aiPlatformEndpoint   string // 外部AI平台或模型服务地址
	aiPlatformApiKey     string // 外部AI平台API Key
}

// NewAIModelServiceConcrete 是 aiModelServiceConcrete 的构造函数。
// 它接收模型元数据仓库、所有必要的下游 gRPC 客户端、外部AI平台地址和API Key，并返回 AIModelService 接口。
func NewAIModelServiceConcrete(
	modelMetadataRepo repository.ModelMetadataRepo,
	productServiceClient ProductServiceClient,
	userServiceClient UserServiceClient,
	orderServiceClient OrderServiceClient,
	reviewServiceClient ReviewServiceClient,
	aiPlatformEndpoint string,
	aiPlatformApiKey string,
) AIModelService {
	return &aiModelServiceConcrete{
		modelMetadataRepo: modelMetadataRepo,
		productServiceClient: productServiceClient,
		userServiceClient:    userServiceClient,
		orderServiceClient:   orderServiceClient,
		reviewServiceClient:  reviewServiceClient,
		aiPlatformEndpoint:   aiPlatformEndpoint,
		aiPlatformApiKey:     aiPlatformApiKey,
	}
}

// --- 推荐系统 ---

// GetProductRecommendations 获取个性化商品推荐。
// 这是一个复杂的业务逻辑，可能涉及用户行为分析、协同过滤、内容推荐等多种算法。
// 它会根据用户ID和上下文信息，调用内部或外部AI模型生成推荐列表。
func (s *aiModelServiceConcrete) GetProductRecommendations(ctx context.Context, req *v1.GetProductRecommendationsRequest) (*v1.GetProductRecommendationsResponse, error) {
	// 权限检查：用户只能获取自己的推荐，管理员可以获取任意用户的推荐
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Authentication failed: %v", err)
	}

	isAdminUser := isAdmin(ctx)
	if !isAdminUser && req.UserId != userID {
		return nil, status.Errorf(codes.PermissionDenied, "Permission denied to get recommendations for another user")
	}

	zap.S().Infof("Getting product recommendations for user %d, count %d, context: %s", req.UserId, req.Count, req.GetContextPage())

	// TODO: 调用实际的推荐算法模型进行推理
	// 这里暂时返回模拟数据
	recommendations := make([]*v1.ProductRecommendation, 0)
	for i := 0; i < int(req.Count); i++ {
		// 模拟推荐商品ID
		productID := uint64(rand.Intn(1000) + 1)
		// 检查是否需要排除
		excluded := false
		for _, excludeID := range req.ExcludeProductIds {
			if productID == excludeID {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		recommendations = append(recommendations, &v1.ProductRecommendation{
			ProductId: productID,
			Score:     rand.Float66(),
			Reason:    fmt.Sprintf("Because you viewed product %d", req.GetCurrentProductId()),
		})
	}

	return &v1.GetProductRecommendationsResponse{
		Recommendations: recommendations,
	}, nil
}

// GetRelatedProducts 获取相关商品。
// 通常基于商品属性、共同购买行为等，调用内部或外部AI模型生成相关商品列表。
func (s *aiModelServiceConcrete) GetRelatedProducts(ctx context.Context, req *v1.GetRelatedProductsRequest) (*v1.GetRelatedProductsResponse, error) {
	zap.S().Infof("Getting related products for product %d, count %d", req.ProductId, req.Count)

	// TODO: 调用实际的相关商品模型进行推理
	// 这里暂时返回模拟数据
	relatedProducts := make([]*v1.ProductRecommendation, 0)
	for i := 0; i < int(req.Count); i++ {
		relatedProducts = append(relatedProducts, &v1.ProductRecommendation{
			ProductId: uint64(rand.Intn(1000) + 1),
			SimilarityScore: rand.Float66(),
			Reason:    "Customers who bought this also bought...",
		})
	}

	return &v1.GetRelatedProductsResponse{
		RelatedProducts: relatedProducts,
	}, nil
}

// GetPersonalizedFeed 获取个性化内容流。
// 可能包含商品、文章、广告等多种内容，根据用户行为和偏好生成。
func (s *aiModelServiceConcrete) GetPersonalizedFeed(ctx context.Context, req *v1.GetPersonalizedFeedRequest) (*v1.GetPersonalizedFeedResponse, error) {
	// 权限检查：用户只能获取自己的内容流，管理员可以获取任意用户的
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Authentication failed: %v", err)
	}

	isAdminUser := isAdmin(ctx)
	if !isAdminUser && req.UserId != userID {
		return nil, status.Errorf(codes.PermissionDenied, "Permission denied to get personalized feed for another user")
	}

	zap.S().Infof("Getting personalized feed for user %d, page %d, size %d", req.UserId, req.PageToken, req.PageSize)

	// TODO: 调用实际的个性化内容流模型进行推理
	// 这里暂时返回模拟数据
	feedItems := make([]*v1.FeedItem, 0)
	for i := 0; i < int(req.PageSize); i++ {
		feedItems = append(feedItems, &v1.FeedItem{
			ItemType:  "product",
			ItemId:    strconv.FormatUint(uint64(rand.Intn(1000)+1), 10),
			Title:     fmt.Sprintf("Recommended Product %d", i+1),
			ImageUrl:  "http://example.com/image.jpg",
			TargetUrl: "http://example.com/product/123",
			Score:     rand.Float66(),
		})
	}

	return &v1.GetPersonalizedFeedResponse{
		FeedItems:     feedItems,
		TotalCount:    100, // 模拟总数
		NextPageToken: req.PageToken + 1,
	}, nil
}

// --- 图像识别 ---

// RecognizeImageContent 识别图片内容，例如商品分类、品牌、属性。
// 它调用内部或外部的图像识别模型对提供的图片进行分析。
func (s *aiModelServiceConcrete) RecognizeImageContent(ctx context.Context, req *v1.RecognizeImageContentRequest) (*v1.RecognizeImageContentResponse, error) {
	zap.S().Infof("Recognizing image content for URL: %s", req.ImageUrl)

	// TODO: 调用外部图像识别服务或内部模型
	// 这里暂时返回模拟数据
	return &v1.RecognizeImageContentResponse{
		Tags: []*v1.ImageTag{
			{Name: "shirt", Confidence: 0.95},
			{Name: "blue", Confidence: 0.88},
		},
		Categories: []*v1.ImageCategory{
			{Name: "Apparel/Tops", Confidence: 0.92},
		},
		DominantColor: "#0000FF",
	}, nil
}

// SearchImageByImage 通过图片搜索相似商品。
// 它调用内部或外部的图像搜索模型，根据提供的图片查找相似商品。
func (s *aiModelServiceConcrete) SearchImageByImage(ctx context.Context, req *v1.SearchImageByImageRequest) (*v1.SearchImageByImageResponse, error) {
	zap.S().Infof("Searching image by image for URL: %s", req.ImageUrl)

	// TODO: 调用外部图像搜索服务或内部模型
	// 这里暂时返回模拟数据
	results := make([]*v1.ProductSearchResult, 0)
	for i := 0; i < int(req.Count); i++ {
		results = append(results, &v1.ProductSearchResult{
			ProductId:      uint64(rand.Intn(1000) + 1),
			SimilarityScore: rand.Float66(),
		})
	}

	return &v1.SearchImageByImageResponse{
		Results: results,
	}, nil
}

// --- 自然语言处理 (NLP) ---

// AnalyzeReviewSentiment 分析用户评论情感。
// 它调用内部或外部的NLP模型对评论文本进行情感分析。
func (s *aiModelServiceConcrete) AnalyzeReviewSentiment(ctx context.Context, req *v1.AnalyzeReviewSentimentRequest) (*v1.AnalyzeReviewSentimentResponse, error) {
	zap.S().Infof("Analyzing sentiment for review: %s", req.ReviewText)

	// TODO: 调用外部NLP服务或内部情感分析模型
	// 这里暂时返回模拟数据
	sentiment := v1.Sentiment_SENTIMENT_NEUTRAL
	score := 0.0
	if strings.Contains(req.ReviewText, "good") || strings.Contains(req.ReviewText, "great") {
		sentiment = v1.Sentiment_SENTIMENT_POSITIVE
		score = rand.Float66()*0.5 + 0.5 // 0.5 to 1.0
	} else if strings.Contains(req.ReviewText, "bad") || strings.Contains(req.ReviewText, "poor") {
		sentiment = v1.Sentiment_SENTIMENT_NEGATIVE
		score = rand.Float66()*(-0.5) - 0.5 // -0.5 to -1.0
	}

	return &v1.AnalyzeReviewSentimentResponse{
		Sentiment:   sentiment,
		Score:       score,
		Explanation: "Simulated sentiment analysis",
	}, nil
}

// ExtractKeywordsFromText 从文本中提取关键词。
// 它调用内部或外部的NLP模型从文本中识别并提取关键信息。
func (s *aiModelServiceConcrete) ExtractKeywordsFromText(ctx context.Context, req *v1.ExtractKeywordsFromTextRequest) (*v1.ExtractKeywordsFromTextResponse, error) {
	zap.S().Infof("Extracting keywords from text: %s", req.Text)

	// TODO: 调用外部NLP服务或内部关键词提取模型
	// 这里暂时返回模拟数据
	keywords := []string{"product", "quality", "price"}
	return &v1.ExtractKeywordsFromTextResponse{
		Keywords: keywords[:min(len(keywords), int(req.Count))],
	}, nil
}

// SummarizeText 总结长文本内容。
// 它调用内部或外部的NLP模型对长文本进行摘要。
func (s *aiModelServiceConcrete) SummarizeText(ctx context.Context, req *v1.SummarizeTextRequest) (*v1.SummarizeTextResponse, error) {
	zap.S().Infof("Summarizing text with length %d to %d", len(req.Text), req.SummaryLength)

	// TODO: 调用外部NLP服务或内部文本摘要模型
	// 这里暂时返回模拟数据
	summary := "This is a simulated summary of the provided text."
	return &v1.SummarizeTextResponse{
		Summary: summary,
	}, nil
}

// --- 欺诈检测 ---

// GetFraudScore 获取交易或用户行为的欺诈评分。
// 它调用内部或外部的欺诈检测模型，评估潜在的欺诈风险。
func (s *aiModelServiceConcrete) GetFraudScore(ctx context.Context, req *v1.GetFraudScoreRequest) (*v1.GetFraudScoreResponse, error) {
	zap.S().Infof("Getting fraud score for user %d, order %d, amount %.2f", req.UserId, req.OrderId, req.TransactionAmount)

	// TODO: 调用外部欺诈检测服务或内部模型
	// 这里暂时返回模拟数据
	fraudScore := rand.Float66() * 0.3 // 模拟较低的欺诈分数
	isFraudulent := fraudScore > 0.5
	reasons := []string{}
	if isFraudulent {
		reasons = append(reasons, "Unusual transaction pattern")
	}

	return &v1.GetFraudScoreResponse{
		FraudScore:  fraudScore,
		IsFraudulent: isFraudulent,
		Reasons:     reasons,
	}, nil
}

// --- 模型管理 (内部/管理员接口) ---

// DeployModel 部署一个新的AI模型版本。
// 此方法通常由管理员或CI/CD系统调用，用于将训练好的模型上线。
func (s *aiModelServiceConcrete) DeployModel(ctx context.Context, req *v1.DeployModelRequest) (*v1.DeployModelResponse, error) {
	// 权限检查：只有管理员可以部署模型
	if !isAdmin(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "Permission denied to deploy AI model")
	}

	zap.S().Infof("Deploying AI model %s version %s from URI %s", req.ModelName, req.ModelVersion, req.ModelUri)

	// TODO: 调用外部AI平台API进行模型部署
	// 模拟部署过程
	deploymentID := fmt.Sprintf("deploy-%s-%s-%d", req.ModelName, req.ModelVersion, time.Now().UnixNano())
	statusStr := "PENDING"

	// 记录模型元数据
	metadata := &model.ModelMetadata{
		ModelName:    req.ModelName,
		ModelVersion: req.ModelVersion,
		ModelURI:     req.ModelUri,
		DeploymentID: deploymentID,
		Status:       statusStr,
		DeployedAt:   time.Now(),
		Metadata:     req.Metadata,
	}
	_, err := s.modelMetadataRepo.CreateModelMetadata(ctx, metadata)
	if err != nil {
		zap.S().Errorf("Failed to save model metadata for deployment %s: %v", deploymentID, err)
		// 部署可能已经开始，但元数据保存失败，需要人工介入
	}

	return &v1.DeployModelResponse{
		DeploymentId: deploymentID,
		Status:       statusStr,
	}, nil
}

// GetModelStatus 获取已部署AI模型的运行状态。
// 它查询模型元数据仓库和外部AI平台，返回模型的当前状态。
func (s *aiModelServiceConcrete) GetModelStatus(ctx context.Context, req *v1.GetModelStatusRequest) (*v1.GetModelStatusResponse, error) {
	// 权限检查：只有管理员可以查看模型状态
	if !isAdmin(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "Permission denied to get AI model status")
	}

	zap.S().Infof("Getting status for deployment ID: %s", req.DeploymentId)

	// 从数据库获取模型元数据
	// TODO: 仓库需要 GetModelMetadataByDeploymentID 方法
	// 暂时通过 List 模拟
	metadatas, _, err := s.modelMetadataRepo.ListModelMetadata(ctx, "", 100, 1)
	if err != nil {
		zap.S().Errorf("Failed to list model metadatas for status check: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to get model status")
	}

	var targetMetadata *model.ModelMetadata
	for _, md := range metadatas {
		if md.DeploymentID == req.DeploymentId {
			targetMetadata = md
			break
		}
	}

	if targetMetadata == nil {
		return nil, status.Errorf(codes.NotFound, "Model deployment with ID %s not found", req.DeploymentId)
	}

	// TODO: 调用外部AI平台API获取实时状态
	// 这里暂时返回数据库中的状态
	return &v1.GetModelStatusResponse{
		DeploymentId: targetMetadata.DeploymentID,
		ModelName:    targetMetadata.ModelName,
		ModelVersion: targetMetadata.ModelVersion,
		Status:       targetMetadata.Status, // 模拟实时状态
		DeployedAt:   timestamppb.New(targetMetadata.DeployedAt),
		ErrorMessage: targetMetadata.ErrorMessage,
	}, nil
}

// RetrainModel 触发AI模型重新训练。
// 此方法通常由管理员或CI/CD系统调用，用于更新模型。
func (s *aiModelServiceConcrete) RetrainModel(ctx context.Context, req *v1.RetrainModelRequest) (*v1.RetrainModelResponse, error) {
	// 权限检查：只有管理员可以触发模型训练
	if !isAdmin(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "Permission denied to retrain AI model")
	}

	zap.S().Infof("Triggering retraining for model %s with dataset URI %s", req.ModelName, req.GetDatasetUri())

	// TODO: 调用外部AI平台API触发模型训练
	// 模拟训练过程
	trainingJobID := fmt.Sprintf("train-%s-%d", req.ModelName, time.Now().UnixNano())
	statusStr := "STARTED"

	// 更新模型元数据 (例如，记录训练任务ID)
	// TODO: 需要根据 model_name 找到最新的模型元数据并更新

	return &v1.RetrainModelResponse{
		TrainingJobId: trainingJobID,
		Status:        statusStr,
	}, nil
}

// --- 辅助函数 ---

// getUserIDFromContext 从 gRPC 上下文的 metadata 中提取用户ID。
// 用于权限校验和个性化服务。
func getUserIDFromContext(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "Cannot get metadata from context")
	}
	values := md.Get("x-user-id")
	if len(values) == 0 {
		// 尝试从 x-admin-user-id 获取，如果当前是管理员操作
		adminValues := md.Get("x-admin-user-id")
		if len(adminValues) > 0 {
			adminUserID, err := strconv.ParseUint(adminValues[0], 10, 64)
			if err != nil {
				return 0, status.Errorf(codes.Unauthenticated, "Invalid x-admin-user-id format")
			}
			return adminUserID, nil
		}
		return 0, status.Errorf(codes.Unauthenticated, "Missing user ID in request header")
	}
	userID, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "Invalid x-user-id format")
	}
	return userID, nil
}

// isAdmin 从 gRPC 上下文的 metadata 中判断当前请求是否由管理员发起。
// 用于权限校验。
func isAdmin(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	// 检查是否存在 x-admin-user-id 头部，并且 x-is-admin 头部为 "true"
	adminIDValues := md.Get("x-admin-user-id")
	isAdminValues := md.Get("x-is-admin")
	return len(adminIDValues) > 0 && len(isAdminValues) > 0 && isAdminValues[0] == "true"
}

// bizModelMetadataToProto 将 model.ModelMetadata 业务领域模型转换为 v1.GetModelStatusResponse API 模型。
func bizModelMetadataToProto(metadata *model.ModelMetadata) *v1.GetModelStatusResponse {
	if metadata == nil {
		return nil
	}
	return &v1.GetModelStatusResponse{
		DeploymentId: metadata.DeploymentID,
		ModelName:    metadata.ModelName,
		ModelVersion: metadata.ModelVersion,
		Status:       metadata.Status,
		DeployedAt:   timestamppb.New(metadata.DeployedAt),
		ErrorMessage: metadata.ErrorMessage,
	}, nil
}
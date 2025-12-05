package grpc

import (
	"context"
	"fmt"     // 导入格式化包。
	"strconv" // 导入字符串和数字转换工具。

	pb "github.com/wyfcoding/ecommerce/go-api/ai_model/v1"            // 导入AI模型模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/ai_model/application" // 导入AI模型模块的应用服务。

	// 导入AI模型模块的领域实体。
	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 AIModelService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedAIModelServiceServer                             // 嵌入生成的UnimplementedAIModelServiceServer，确保前向兼容性。
	app                                  *application.AIModelService // 依赖AIModel应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 AIModel gRPC 服务端实例。
func NewServer(app *application.AIModelService) *Server {
	return &Server{app: app}
}

// --- Model Management 模型管理 ---

// DeployModel 处理部署AI模型的gRPC请求。
// req: 包含模型部署所需信息的请求体。
// 返回部署成功的模型响应和可能发生的gRPC错误。
func (s *Server) DeployModel(ctx context.Context, req *pb.DeployModelRequest) (*pb.DeployModelResponse, error) {
	// 将protobuf请求映射到应用服务层的CreateModel方法。
	// 注意：Proto中缺少creatorID等字段，因此使用默认值或占位符。
	// model_name, model_version, model_uri, metadata等字段在CreateModel中并未全部使用。
	model, err := s.app.CreateModel(ctx, req.ModelName, "Deployed via gRPC", "generic", "unknown", 0)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create model for deployment: %v", err))
	}

	// 部署模型。
	// 注意：Proto中的model_uri（模型文件路径）未在CreateModel中直接处理。
	// 实际部署可能需要将model_uri更新到模型实体中，或在部署流程中处理。
	if err := s.app.Deploy(ctx, uint64(model.ID)); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to deploy model: %v", err))
	}

	// 返回部署成功的响应。
	return &pb.DeployModelResponse{
		DeploymentId: strconv.FormatUint(uint64(model.ID), 10), // 部署ID使用模型ID的字符串形式。
		Status:       "PENDING",                                // 对应模型状态为 Deploying。
	}, nil
}

// GetModelStatus 处理获取模型状态的gRPC请求。
// req: 包含部署ID的请求体。
// 返回模型状态响应和可能发生的gRPC错误。
func (s *Server) GetModelStatus(ctx context.Context, req *pb.GetModelStatusRequest) (*pb.GetModelStatusResponse, error) {
	// 将DeploymentId（字符串）转换为uint64，作为模型ID。
	id, err := strconv.ParseUint(req.DeploymentId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid deployment_id")
	}

	// 调用应用服务层获取模型详情。
	model, err := s.app.GetModelDetails(ctx, id)
	if err != nil {
		// 如果模型未找到，返回NotFound状态码。
		return nil, status.Error(codes.NotFound, fmt.Sprintf("model not found: %v", err))
	}

	// 将部署时间转换为protobuf的时间戳格式。
	var deployedAt *timestamppb.Timestamp
	if model.DeployedAt != nil {
		deployedAt = timestamppb.New(*model.DeployedAt)
	}

	// 返回模型状态响应。
	return &pb.GetModelStatusResponse{
		DeploymentId: req.DeploymentId,     // 部署ID。
		ModelName:    model.Name,           // 模型名称。
		ModelVersion: model.Version,        // 模型版本。
		Status:       string(model.Status), // 模型状态。
		DeployedAt:   deployedAt,           // 部署时间。
		ErrorMessage: &model.FailedReason,  // 失败原因。
	}, nil
}

// RetrainModel 处理重新训练AI模型的gRPC请求。
// req: 包含模型名称的请求体。
// 返回重新训练的响应和可能发生的gRPC错误。
func (s *Server) RetrainModel(ctx context.Context, req *pb.RetrainModelRequest) (*pb.RetrainModelResponse, error) {
	// 注意：Proto中使用model_name作为标识符，但应用服务层需要模型ID（uint64）。
	// 当前实现假设 model_name 即为模型ID的字符串形式。
	id, err := strconv.ParseUint(req.ModelName, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "model_name must be a valid ID for retraining")
	}

	// 调用应用服务层启动模型训练。
	if err := s.app.StartTraining(ctx, id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to start model training: %v", err))
	}

	// 返回重新训练响应。
	return &pb.RetrainModelResponse{
		TrainingJobId: req.ModelName, // 训练任务ID复用模型名称。
		Status:        "STARTED",     // 训练状态。
	}, nil
}

// --- Feature Methods (Unimplemented) 功能方法（未实现） ---
// 以下gRPC方法均未实现，仅返回Unimplemented错误。

// GetProductRecommendations 获取产品推荐。
func (s *Server) GetProductRecommendations(ctx context.Context, req *pb.GetProductRecommendationsRequest) (*pb.GetProductRecommendationsResponse, error) {
	recs, err := s.app.GetProductRecommendations(ctx, req.UserId, req.GetContextPage())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get product recommendations: %v", err))
	}

	var pbRecs []*pb.ProductRecommendation
	for _, r := range recs {
		pbRecs = append(pbRecs, &pb.ProductRecommendation{
			ProductId: r.ProductID,
			Score:     r.Score,
			Reason:    &r.Reason,
		})
	}
	return &pb.GetProductRecommendationsResponse{Recommendations: pbRecs}, nil
}

// GetRelatedProducts 获取相关商品。
func (s *Server) GetRelatedProducts(ctx context.Context, req *pb.GetRelatedProductsRequest) (*pb.GetRelatedProductsResponse, error) {
	recs, err := s.app.GetRelatedProducts(ctx, req.ProductId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get related products: %v", err))
	}

	var pbRecs []*pb.ProductRecommendation
	for _, r := range recs {
		pbRecs = append(pbRecs, &pb.ProductRecommendation{
			ProductId: r.ProductID,
			Score:     r.Score,
			Reason:    &r.Reason,
		})
	}
	return &pb.GetRelatedProductsResponse{RelatedProducts: pbRecs}, nil
}

// GetPersonalizedFeed 获取个性化内容流。
func (s *Server) GetPersonalizedFeed(ctx context.Context, req *pb.GetPersonalizedFeedRequest) (*pb.GetPersonalizedFeedResponse, error) {
	items, err := s.app.GetPersonalizedFeed(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get personalized feed: %v", err))
	}

	var pbItems []*pb.FeedItem
	for _, item := range items {
		pbItems = append(pbItems, &pb.FeedItem{
			ItemType:  item.ItemType,
			ItemId:    item.ItemID,
			Title:     item.Title,
			ImageUrl:  item.ImageURL,
			TargetUrl: item.TargetURL,
			Score:     item.Score,
		})
	}
	return &pb.GetPersonalizedFeedResponse{FeedItems: pbItems}, nil
}

// RecognizeImageContent 识别图片内容。
func (s *Server) RecognizeImageContent(ctx context.Context, req *pb.RecognizeImageContentRequest) (*pb.RecognizeImageContentResponse, error) {
	labels, err := s.app.RecognizeImageContent(ctx, req.ImageUrl)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to recognize image content: %v", err))
	}

	var tags []*pb.ImageTag
	for _, label := range labels {
		tags = append(tags, &pb.ImageTag{Name: label, Confidence: 0.9})
	}

	return &pb.RecognizeImageContentResponse{Tags: tags}, nil
}

// SearchImageByImage 通过图片搜索相似商品。
func (s *Server) SearchImageByImage(ctx context.Context, req *pb.SearchImageByImageRequest) (*pb.SearchImageByImageResponse, error) {
	results, err := s.app.SearchImageByImage(ctx, req.ImageUrl)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to search image by image: %v", err))
	}

	var pbResults []*pb.ProductSearchResult
	for _, r := range results {
		pbResults = append(pbResults, &pb.ProductSearchResult{ProductId: r.ProductID, SimilarityScore: r.SimilarityScore})
	}

	return &pb.SearchImageByImageResponse{Results: pbResults}, nil
}

// AnalyzeReviewSentiment 分析用户评论情感。
func (s *Server) AnalyzeReviewSentiment(ctx context.Context, req *pb.AnalyzeReviewSentimentRequest) (*pb.AnalyzeReviewSentimentResponse, error) {
	score, explanation, err := s.app.AnalyzeReviewSentiment(ctx, req.ReviewText)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to analyze review sentiment: %v", err))
	}

	var sentiment pb.Sentiment
	if score > 0.5 {
		sentiment = pb.Sentiment_SENTIMENT_POSITIVE
	} else if score < -0.5 {
		sentiment = pb.Sentiment_SENTIMENT_NEGATIVE
	} else {
		sentiment = pb.Sentiment_SENTIMENT_NEUTRAL
	}

	return &pb.AnalyzeReviewSentimentResponse{
		Sentiment:   sentiment,
		Score:       score,
		Explanation: &explanation,
	}, nil
}

// ExtractKeywordsFromText 从文本中提取关键词。
func (s *Server) ExtractKeywordsFromText(ctx context.Context, req *pb.ExtractKeywordsFromTextRequest) (*pb.ExtractKeywordsFromTextResponse, error) {
	keywords, err := s.app.ExtractKeywordsFromText(ctx, req.Text)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to extract keywords: %v", err))
	}
	return &pb.ExtractKeywordsFromTextResponse{Keywords: keywords}, nil
}

// SummarizeText 总结长文本内容。
func (s *Server) SummarizeText(ctx context.Context, req *pb.SummarizeTextRequest) (*pb.SummarizeTextResponse, error) {
	summary, err := s.app.SummarizeText(ctx, req.Text)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to summarize text: %v", err))
	}
	return &pb.SummarizeTextResponse{Summary: summary}, nil
}

// GetFraudScore 获取欺诈评分。
func (s *Server) GetFraudScore(ctx context.Context, req *pb.GetFraudScoreRequest) (*pb.GetFraudScoreResponse, error) {
	result, err := s.app.GetFraudScore(ctx, req.UserId, req.TransactionAmount, req.IpAddress)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get fraud score: %v", err))
	}
	return &pb.GetFraudScoreResponse{
		FraudScore:   result.FraudScore,
		IsFraudulent: result.IsFraudulent,
		Reasons:      result.Reasons,
	}, nil
}

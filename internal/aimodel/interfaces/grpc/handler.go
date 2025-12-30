package grpc

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/wyfcoding/ecommerce/goapi/aimodel/v1"
	"github.com/wyfcoding/ecommerce/internal/aimodel/application"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 AIModelService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedAIModelServiceServer
	app *application.AIModelService
}

// NewServer 创建并返回一个新的 AIModel gRPC 服务端实例。
func NewServer(app *application.AIModelService) *Server {
	return &Server{app: app}
}

// DeployModel 处理部署AI模型的gRPC请求。
func (s *Server) DeployModel(ctx context.Context, req *pb.DeployModelRequest) (*pb.DeployModelResponse, error) {
	model, err := s.app.CreateModel(ctx, req.ModelName, "Deployed via gRPC", "generic", "unknown", 0)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create model for deployment: %v", err))
	}

	if err := s.app.Deploy(ctx, uint64(model.ID)); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to deploy model: %v", err))
	}

	return &pb.DeployModelResponse{
		DeploymentId: strconv.FormatUint(uint64(model.ID), 10),
		Status:       "PENDING",
	}, nil
}

// GetModelStatus 处理获取模型状态的gRPC请求。
func (s *Server) GetModelStatus(ctx context.Context, req *pb.GetModelStatusRequest) (*pb.GetModelStatusResponse, error) {
	id, err := strconv.ParseUint(req.DeploymentId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid deployment_id")
	}

	model, err := s.app.GetModelDetails(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("model not found: %v", err))
	}

	var deployedAt *timestamppb.Timestamp
	if model.DeployedAt != nil {
		deployedAt = timestamppb.New(*model.DeployedAt)
	}

	return &pb.GetModelStatusResponse{
		DeploymentId: req.DeploymentId,
		ModelName:    model.Name,
		ModelVersion: model.Version,
		Status:       string(model.Status),
		DeployedAt:   deployedAt,
		ErrorMessage: &model.FailedReason,
	}, nil
}

// RetrainModel 处理重新训练AI模型的gRPC请求。
func (s *Server) RetrainModel(ctx context.Context, req *pb.RetrainModelRequest) (*pb.RetrainModelResponse, error) {
	id, err := strconv.ParseUint(req.ModelName, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "model_name must be a valid ID for retraining")
	}

	if err := s.app.StartTraining(ctx, id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to start model training: %v", err))
	}

	return &pb.RetrainModelResponse{
		TrainingJobId: req.ModelName,
		Status:        "STARTED",
	}, nil
}

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

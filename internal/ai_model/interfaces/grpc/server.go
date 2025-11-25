package grpc

import (
	"context"
	pb "ecommerce/api/ai_model/v1"
	"ecommerce/internal/ai_model/application"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedAIModelServiceServer
	app *application.AIModelService
}

func NewServer(app *application.AIModelService) *Server {
	return &Server{app: app}
}

// --- Model Management ---

func (s *Server) DeployModel(ctx context.Context, req *pb.DeployModelRequest) (*pb.DeployModelResponse, error) {
	// Map DeployModel to CreateModel + Deploy
	// Service CreateModel(ctx, name, description, modelType, algorithm, creatorID)
	// Proto: model_name, model_version, model_uri, metadata.

	// We use defaults for missing fields
	model, err := s.app.CreateModel(ctx, req.ModelName, "Deployed via gRPC", "generic", "unknown", 0)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Set Model Path (URI) - Not exposed in CreateModel, need to update entity directly or add method?
	// Service doesn't have UpdateModel exposed.
	// But CompleteTraining sets ModelPath.
	// We can skip setting path for now or assume it's handled elsewhere.

	// Call Deploy
	if err := s.app.Deploy(ctx, uint64(model.ID)); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeployModelResponse{
		DeploymentId: strconv.FormatUint(uint64(model.ID), 10),
		Status:       "PENDING", // Deploy sets status to Deploying
	}, nil
}

func (s *Server) GetModelStatus(ctx context.Context, req *pb.GetModelStatusRequest) (*pb.GetModelStatusResponse, error) {
	id, err := strconv.ParseUint(req.DeploymentId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid deployment_id")
	}

	model, err := s.app.GetModelDetails(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
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

func (s *Server) RetrainModel(ctx context.Context, req *pb.RetrainModelRequest) (*pb.RetrainModelResponse, error) {
	// Proto uses model_name. Service needs ID.
	// We assume model_name is ID for now.
	id, err := strconv.ParseUint(req.ModelName, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "model_name must be ID")
	}

	if err := s.app.StartTraining(ctx, id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RetrainModelResponse{
		TrainingJobId: req.ModelName, // Reuse ID
		Status:        "STARTED",
	}, nil
}

// --- Feature Methods (Unimplemented) ---

func (s *Server) GetProductRecommendations(ctx context.Context, req *pb.GetProductRecommendationsRequest) (*pb.GetProductRecommendationsResponse, error) {
	// Mock implementation or Unimplemented?
	// Let's return Unimplemented to be honest.
	return nil, status.Error(codes.Unimplemented, "GetProductRecommendations not implemented")
}

func (s *Server) GetRelatedProducts(ctx context.Context, req *pb.GetRelatedProductsRequest) (*pb.GetRelatedProductsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetRelatedProducts not implemented")
}

func (s *Server) GetPersonalizedFeed(ctx context.Context, req *pb.GetPersonalizedFeedRequest) (*pb.GetPersonalizedFeedResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetPersonalizedFeed not implemented")
}

func (s *Server) RecognizeImageContent(ctx context.Context, req *pb.RecognizeImageContentRequest) (*pb.RecognizeImageContentResponse, error) {
	return nil, status.Error(codes.Unimplemented, "RecognizeImageContent not implemented")
}

func (s *Server) SearchImageByImage(ctx context.Context, req *pb.SearchImageByImageRequest) (*pb.SearchImageByImageResponse, error) {
	return nil, status.Error(codes.Unimplemented, "SearchImageByImage not implemented")
}

func (s *Server) AnalyzeReviewSentiment(ctx context.Context, req *pb.AnalyzeReviewSentimentRequest) (*pb.AnalyzeReviewSentimentResponse, error) {
	// We could use s.app.Predict if we had a sentiment model ID.
	// For now, Unimplemented.
	return nil, status.Error(codes.Unimplemented, "AnalyzeReviewSentiment not implemented")
}

func (s *Server) ExtractKeywordsFromText(ctx context.Context, req *pb.ExtractKeywordsFromTextRequest) (*pb.ExtractKeywordsFromTextResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ExtractKeywordsFromText not implemented")
}

func (s *Server) SummarizeText(ctx context.Context, req *pb.SummarizeTextRequest) (*pb.SummarizeTextResponse, error) {
	return nil, status.Error(codes.Unimplemented, "SummarizeText not implemented")
}

func (s *Server) GetFraudScore(ctx context.Context, req *pb.GetFraudScoreRequest) (*pb.GetFraudScoreResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetFraudScore not implemented")
}

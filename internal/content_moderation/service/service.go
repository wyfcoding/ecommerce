package service

import (
	"context"
	"time"

	v1 "ecommerce/api/content_moderation/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ContentModerationService is the gRPC service implementation for content moderation.
type ContentModerationService struct {
	v1.UnimplementedContentModerationServiceServer
	uc *biz.ContentModerationUsecase
}

// NewContentModerationService creates a new ContentModerationService.
func NewContentModerationService(uc *biz.ContentModerationUsecase) *ContentModerationService {
	return &ContentModerationService{uc: uc}
}

// bizModerationResultToProto converts biz.ModerationResult to v1.ModerateTextResponse/ModerateImageResponse.
func bizModerationResultToProto(result *biz.ModerationResult) (bool, []string, float64, string) {
	if result == nil {
		return false, nil, 0, "ERROR"
	}
	return result.IsSafe, result.Labels, result.Confidence, result.Decision
}

// ModerateText implements the ModerateText RPC.
func (s *ContentModerationService) ModerateText(ctx context.Context, req *v1.ModerateTextRequest) (*v1.ModerateTextResponse, error) {
	if req.Text == "" || req.ContentType == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "text, content_type, and user_id are required")
	}

	result, err := s.uc.ModerateText(ctx, "text_"+time.Now().Format("20060102150405"), req.ContentType, req.UserId, req.Text)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to moderate text: %v", err)
	}

	isSafe, labels, confidence, decision := bizModerationResultToProto(result)
	return &v1.ModerateTextResponse{
		IsSafe:     isSafe,
		Labels:     labels,
		Confidence: confidence,
		Decision:   decision,
	}, nil
}

// ModerateImage implements the ModerateImage RPC.
func (s *ContentModerationService) ModerateImage(ctx context.Context, req *v1.ModerateImageRequest) (*v1.ModerateImageResponse, error) {
	if req.ImageUrl == "" || req.ContentType == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "image_url, content_type, and user_id are required")
	}

	result, err := s.uc.ModerateImage(ctx, "image_"+time.Now().Format("20060102150405"), req.ContentType, req.UserId, req.ImageUrl)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to moderate image: %v", err)
	}

	isSafe, labels, confidence, decision := bizModerationResultToProto(result)
	return &v1.ModerateImageResponse{
		IsSafe:     isSafe,
		Labels:     labels,
		Confidence: confidence,
		Decision:   decision,
	}, nil
}

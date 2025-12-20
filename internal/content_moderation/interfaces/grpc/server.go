package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/content_moderation/v1"
	"github.com/wyfcoding/ecommerce/internal/content_moderation/application"
	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server 结构体实现了 ContentModerationService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedContentModerationServer
	app *application.ModerationService
}

// NewServer 创建并返回一个新的 ContentModeration gRPC 服务端实例。
func NewServer(app *application.ModerationService) *Server {
	return &Server{app: app}
}

// ModerateText 处理文本内容审核的gRPC请求。
func (s *Server) ModerateText(ctx context.Context, req *pb.ModerateTextRequest) (*pb.ModerateTextResponse, error) {
	record, err := s.app.SubmitContent(ctx, domain.ContentTypeText, 0, req.Text, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to submit text for moderation: %v", err))
	}

	isSafe := false
	var rejectionReason *string

	switch record.Status {
	case domain.ModerationStatusApproved:
		isSafe = true
	case domain.ModerationStatusRejected:
		isSafe = false
		r := record.RejectReason
		rejectionReason = &r
	case domain.ModerationStatusPending:
		isSafe = false
		r := "Pending Review"
		rejectionReason = &r
	}

	return &pb.ModerateTextResponse{
		IsSafe:           isSafe,
		ModerationLabels: record.AITags,
		RejectionReason:  rejectionReason,
	}, nil
}

// ModerateImage 处理图片内容审核的gRPC请求。
func (s *Server) ModerateImage(ctx context.Context, req *pb.ModerateImageRequest) (*pb.ModerateImageResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ModerateImage not implemented for raw bytes directly; image upload to URL needed first.")
}

package grpc

import (
	"context"
	pb "github.com/wyfcoding/ecommerce/api/content_moderation/v1"
	"github.com/wyfcoding/ecommerce/internal/content_moderation/application"
	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedContentModerationServer
	app *application.ModerationService
}

func NewServer(app *application.ModerationService) *Server {
	return &Server{app: app}
}

func (s *Server) ModerateText(ctx context.Context, req *pb.ModerateTextRequest) (*pb.ModerateTextResponse, error) {
	// Use 0 as contentID since we are just checking text
	record, err := s.app.SubmitContent(ctx, entity.ContentTypeText, 0, req.Text, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	isSafe := false
	var rejectionReason *string

	switch record.Status {
	case entity.ModerationStatusApproved:
		isSafe = true
	case entity.ModerationStatusRejected:
		isSafe = false
		r := record.RejectReason
		rejectionReason = &r
	case entity.ModerationStatusPending:
		// Treat pending as unsafe or just pending?
		// Proto expects boolean is_safe.
		// Let's say false for pending to be safe.
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

func (s *Server) ModerateImage(ctx context.Context, req *pb.ModerateImageRequest) (*pb.ModerateImageResponse, error) {
	// Service expects string content (URL/Text), but proto provides bytes.
	// We cannot easily map this without uploading the image first.
	// For now, return Unimplemented.
	return nil, status.Error(codes.Unimplemented, "ModerateImage not implemented for raw bytes")
}

package interfaces

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/kyc"
	"github.com/wyfcoding/ecommerce/internal/kyc/application"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type KYCHandler struct {
	pb.UnimplementedKYCServiceServer
	appService *application.KYCApplicationService
}

func NewKYCHandler(appService *application.KYCApplicationService) *KYCHandler {
	return &KYCHandler{
		appService: appService,
	}
}

func (h *KYCHandler) SubmitKYC(ctx context.Context, req *pb.SubmitKYCRequest) (*pb.SubmitKYCResponse, error) {
	cmd := application.SubmitKYCCommand{
		UserID:   req.UserId,
		FullName: req.FullName,
		IDNumber: req.IdNumber,
		IDType:   req.IdType,
		IDDocURL: req.IdDocumentUrl,
	}

	appID, err := h.appService.Submit(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &pb.SubmitKYCResponse{
		ApplicationId: appID,
		Status:        "PENDING",
	}, nil
}

func (h *KYCHandler) GetKYCStatus(ctx context.Context, req *pb.GetKYCStatusRequest) (*pb.GetKYCStatusResponse, error) {
	app, err := h.appService.GetStatus(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetKYCStatusResponse{
		UserId:     app.UserID,
		Status:     app.Status,
		Reason:     app.Reason,
		VerifiedAt: timestamppb.New(app.VerifiedAt),
	}, nil
}

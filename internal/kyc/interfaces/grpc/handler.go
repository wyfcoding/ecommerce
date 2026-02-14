package grpc

import (
	"context"
	"strings"

	pb "github.com/wyfcoding/ecommerce/go-api/kyc"
	"github.com/wyfcoding/ecommerce/internal/kyc/application"
	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KYCHandler KYC gRPC 处理器。
type KYCHandler struct {
	pb.UnimplementedKYCServiceServer
	cmd   *application.KYCCommandService
	query *application.KYCQueryService
}

// NewKYCHandler 创建 KYC 处理器。
func NewKYCHandler(cmd *application.KYCCommandService, query *application.KYCQueryService) *KYCHandler {
	return &KYCHandler{cmd: cmd, query: query}
}

// SubmitKYC 提交 KYC 申请。
func (h *KYCHandler) SubmitKYC(ctx context.Context, req *pb.SubmitKYCRequest) (*pb.SubmitKYCResponse, error) {
	app, err := h.cmd.SubmitKYC(ctx, application.SubmitKYCCommand{
		UserID:   req.UserId,
		FullName: req.FullName,
		IDNumber: req.IdNumber,
		IDType:   parseIDType(req.IdType),
	})
	if err != nil {
		return nil, err
	}
	return &pb.SubmitKYCResponse{
		ApplicationId: app.ApplicationID,
		Status:        app.Status.String(),
	}, nil
}

// GetKYCStatus 获取 KYC 状态。
func (h *KYCHandler) GetKYCStatus(ctx context.Context, req *pb.GetKYCStatusRequest) (*pb.GetKYCStatusResponse, error) {
	dto, err := h.query.GetKYCStatus(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetKYCStatusResponse{
		UserId: req.UserId,
		Status: dto.Status,
		Reason: dto.Reason,
	}
	if dto.VerifiedAt != nil {
		resp.VerifiedAt = timestamppb.New(*dto.VerifiedAt)
	}
	return resp, nil
}

func parseIDType(raw string) domain.IDType {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "PASSPORT":
		return domain.IDTypePassport
	case "DRIVER_LICENSE":
		return domain.IDTypeDriversLicense
	case "RESIDENCE_PERMIT":
		return domain.IDTypeResidencePermit
	case "BUSINESS_LICENSE":
		return domain.IDTypeBusinessLicense
	default:
		return domain.IDTypeIDCard
	}
}

package service

import (
	"context"
	v1 "ecommerce/api/risk_security/v1"
	"ecommerce/internal/risk_security/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RiskSecurityService is the gRPC service implementation for risk and security.
type RiskSecurityService struct {
	v1.UnimplementedRiskSecurityServiceServer
	uc *biz.RiskSecurityUsecase
}

// NewRiskSecurityService creates a new RiskSecurityService.
func NewRiskSecurityService(uc *biz.RiskSecurityUsecase) *RiskSecurityService {
	return &RiskSecurityService{uc: uc}
}

// PerformAntiFraudCheck implements the PerformAntiFraudCheck RPC.
func (s *RiskSecurityService) PerformAntiFraudCheck(ctx context.Context, req *v1.PerformAntiFraudCheckRequest) (*v1.PerformAntiFraudCheckResponse, error) {
	if req.UserId == "" || req.IpAddress == "" || req.DeviceInfo == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, ip_address, and device_info are required")
	}

	additionalData := make(map[string]string)
	for k, v := range req.AdditionalData {
		additionalData[k] = v
	}

	result, err := s.uc.PerformAntiFraudCheck(ctx, req.UserId, req.IpAddress, req.DeviceInfo, req.OrderId, req.Amount, additionalData)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to perform anti-fraud check: %v", err)
	}

	return &v1.PerformAntiFraudCheckResponse{
		IsFraud:  result.IsFraud,
		RiskScore: result.RiskScore,
		Decision: result.Decision,
		Message:  result.Message,
	}, nil
}

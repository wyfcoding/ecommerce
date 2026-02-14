package grpc

import (
	"context"

	v1 "github.com/wyfcoding/ecommerce/go-api/insurance/v1"
	"github.com/wyfcoding/ecommerce/internal/insurance/application"
	"github.com/wyfcoding/ecommerce/internal/insurance/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	v1.UnimplementedInsuranceServiceServer
	app *application.InsuranceService
}

func NewServer(app *application.InsuranceService) *Server {
	return &Server{app: app}
}

func (s *Server) CreatePolicy(ctx context.Context, req *v1.CreatePolicyRequest) (*v1.CreatePolicyResponse, error) {
	policy, err := s.app.CreatePolicy(ctx, application.CreatePolicyRequest{
		OrderID:        req.OrderId,
		UserID:         req.UserId,
		Type:           toDomainPolicyType(req.Type),
		Premium:        req.Premium,
		CoverageAmount: req.CoverageAmount,
		DurationDays:   int(req.DurationDays),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create policy: %v", err)
	}

	return &v1.CreatePolicyResponse{
		PolicyId: policy.PolicyID,
		Status:   toProtoPolicyStatus(policy.Status),
	}, nil
}

func (s *Server) GetPolicy(ctx context.Context, req *v1.GetPolicyRequest) (*v1.GetPolicyResponse, error) {
	policy, err := s.app.GetPolicy(ctx, req.PolicyId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "policy not found: %v", err)
	}

	return &v1.GetPolicyResponse{
		Policy: toProtoPolicy(policy),
	}, nil
}

func (s *Server) FileClaim(ctx context.Context, req *v1.FileClaimRequest) (*v1.FileClaimResponse, error) {
	claim, err := s.app.FileClaim(ctx, application.FileClaimRequest{
		PolicyID:     req.PolicyId,
		UserID:       req.UserId,
		Reason:       req.Reason,
		Amount:       req.Amount,
		EvidenceURLs: req.EvidenceUrls,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to file claim: %v", err)
	}

	return &v1.FileClaimResponse{
		ClaimId: claim.ClaimID,
		Status:  toProtoClaimStatus(claim.Status),
	}, nil
}

func (s *Server) GetClaim(ctx context.Context, req *v1.GetClaimRequest) (*v1.GetClaimResponse, error) {
	claim, err := s.app.GetClaim(ctx, req.ClaimId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "claim not found: %v", err)
	}

	return &v1.GetClaimResponse{
		Claim: toProtoClaim(claim),
	}, nil
}

func (s *Server) ListPolicies(ctx context.Context, req *v1.ListPoliciesRequest) (*v1.ListPoliciesResponse, error) {
	policies, nextPageToken, err := s.app.ListPolicies(ctx, req.UserId, req.OrderId, int(req.PageSize), req.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list policies: %v", err)
	}

	protoPolicies := make([]*v1.Policy, len(policies))
	for i, p := range policies {
		protoPolicies[i] = toProtoPolicy(p)
	}

	return &v1.ListPoliciesResponse{
		Policies:      protoPolicies,
		NextPageToken: nextPageToken,
	}, nil
}

func toProtoPolicy(p *domain.InsurancePolicy) *v1.Policy {
	premium, _ := p.Premium.Float64()
	coverageAmount, _ := p.CoverageAmount.Float64()
	return &v1.Policy{
		Id:             p.PolicyID,
		OrderId:        p.OrderID,
		UserId:         p.UserID,
		Type:           toProtoPolicyType(p.Type),
		Premium:        premium,
		CoverageAmount: coverageAmount,
		Status:         toProtoPolicyStatus(p.Status),
		StartTime:      p.StartTime.Unix(),
		EndTime:        p.EndTime.Unix(),
		CreatedAt:      p.CreatedAt.Unix(),
	}
}

func toProtoClaim(c *domain.InsuranceClaim) *v1.Claim {
	amountRequested, _ := c.AmountRequested.Float64()
	amountApproved, _ := c.AmountApproved.Float64()
	return &v1.Claim{
		Id:              c.ClaimID,
		PolicyId:        c.PolicyID,
		UserId:          c.UserID,
		Reason:          c.Reason,
		AmountRequested: amountRequested,
		AmountApproved:  amountApproved,
		Status:          toProtoClaimStatus(c.Status),
		RejectReason:    c.RejectReason,
		CreatedAt:       c.CreatedAt.Unix(),
		UpdatedAt:       c.UpdatedAt.Unix(),
		// EvidenceURLs parsing skipped for brevity
	}
}

func toDomainPolicyType(policyType v1.PolicyType) domain.PolicyType {
	switch policyType {
	case v1.PolicyType_POLICY_TYPE_SHIPPING_RETURN:
		return domain.PolicyTypeShippingInsurance
	case v1.PolicyType_POLICY_TYPE_PRICE_PROTECTION:
		return domain.PolicyTypePriceProtection
	case v1.PolicyType_POLICY_TYPE_QUALITY_ASSURANCE:
		return domain.PolicyTypeQualityAssurance
	default:
		return domain.PolicyTypeShippingInsurance
	}
}

func toProtoPolicyType(policyType domain.PolicyType) v1.PolicyType {
	switch policyType {
	case domain.PolicyTypeShippingInsurance, domain.PolicyTypeReturnInsurance:
		return v1.PolicyType_POLICY_TYPE_SHIPPING_RETURN
	case domain.PolicyTypePriceProtection:
		return v1.PolicyType_POLICY_TYPE_PRICE_PROTECTION
	case domain.PolicyTypeQualityAssurance, domain.PolicyTypeExtendedWarranty, domain.PolicyTypeDamageInsurance:
		return v1.PolicyType_POLICY_TYPE_QUALITY_ASSURANCE
	default:
		return v1.PolicyType_POLICY_TYPE_UNSPECIFIED
	}
}

func toProtoPolicyStatus(status domain.PolicyStatus) v1.PolicyStatus {
	switch status {
	case domain.PolicyStatusActive:
		return v1.PolicyStatus_POLICY_STATUS_ACTIVE
	case domain.PolicyStatusExpired:
		return v1.PolicyStatus_POLICY_STATUS_EXPIRED
	case domain.PolicyStatusCancelled:
		return v1.PolicyStatus_POLICY_STATUS_CANCELLED
	case domain.PolicyStatusClaimed:
		return v1.PolicyStatus_POLICY_STATUS_CLAIMED
	default:
		return v1.PolicyStatus_POLICY_STATUS_UNSPECIFIED
	}
}

func toProtoClaimStatus(status domain.ClaimStatus) v1.ClaimStatus {
	switch status {
	case domain.ClaimStatusApproved:
		return v1.ClaimStatus_CLAIM_STATUS_APPROVED
	case domain.ClaimStatusRejected:
		return v1.ClaimStatus_CLAIM_STATUS_REJECTED
	case domain.ClaimStatusPaid, domain.ClaimStatusClosed:
		return v1.ClaimStatus_CLAIM_STATUS_PAID
	case domain.ClaimStatusSubmitted, domain.ClaimStatusUnderReview, domain.ClaimStatusProcessing:
		return v1.ClaimStatus_CLAIM_STATUS_PENDING
	default:
		return v1.ClaimStatus_CLAIM_STATUS_UNSPECIFIED
	}
}

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
		Type:           domain.PolicyType(req.Type),
		Premium:        req.Premium,
		CoverageAmount: req.CoverageAmount,
		DurationDays:   int(req.DurationDays),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create policy: %v", err)
	}

	return &v1.CreatePolicyResponse{
		PolicyId: policy.PolicyID,
		Status:   v1.PolicyStatus(policy.Status),
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
		Status:  v1.ClaimStatus(claim.Status),
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
	return &v1.Policy{
		Id:             p.PolicyID,
		OrderId:        p.OrderID,
		UserId:         p.UserID,
		Type:           v1.PolicyType(p.Type),
		Premium:        p.Premium,
		CoverageAmount: p.CoverageAmount,
		Status:         v1.PolicyStatus(p.Status),
		StartTime:      p.StartTime.Unix(),
		EndTime:        p.EndTime.Unix(),
		CreatedAt:      p.CreatedAt.Unix(),
	}
}

func toProtoClaim(c *domain.InsuranceClaim) *v1.Claim {
	return &v1.Claim{
		Id:              c.ClaimID,
		PolicyId:        c.PolicyID,
		UserId:          c.UserID,
		Reason:          c.Reason,
		AmountRequested: c.AmountRequested,
		AmountApproved:  c.AmountApproved,
		Status:          v1.ClaimStatus(c.Status),
		RejectReason:    c.RejectReason,
		CreatedAt:       c.CreatedAt.Unix(),
		UpdatedAt:       c.UpdatedAt.Unix(),
		// EvidenceURLs parsing skipped for brevity
	}
}

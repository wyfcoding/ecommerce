package service

import (
	"context"

	"ecommerce/internal/subscription/biz"
	// Assuming the generated pb.go file will be in this path
	// v1 "ecommerce/api/subscription/v1"
)

// SubscriptionService is a gRPC service that implements the SubscriptionServer interface.
// It holds a reference to the business logic layer.
type SubscriptionService struct {
	// v1.UnimplementedSubscriptionServer

	uc *biz.SubscriptionUsecase
}

// NewSubscriptionService creates a new SubscriptionService.
func NewSubscriptionService(uc *biz.SubscriptionUsecase) *SubscriptionService {
	return &SubscriptionService{uc: uc}
}

// Note: The actual RPC methods like CreateSubscriptionPlan, CreateUserSubscription, etc., will be implemented here.
// These methods will call the corresponding business logic in the 'biz' layer.

/*
Example Implementation (once gRPC code is generated):

func (s *SubscriptionService) CreateSubscriptionPlan(ctx context.Context, req *v1.CreateSubscriptionPlanRequest) (*v1.CreateSubscriptionPlanResponse, error) {
    // 1. Call business logic
    plan, err := s.uc.CreateSubscriptionPlan(ctx, req.Name, req.Description, req.Price, req.Currency, req.RecurrenceType, req.DurationMonths, req.IsActive)
    if err != nil {
        return nil, err
    }

    // 2. Convert biz model to API model and return
    return &v1.CreateSubscriptionPlanResponse{Plan: plan}, nil
}

*/

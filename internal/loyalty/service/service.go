package service

import (
	// Assuming the generated pb.go file will be in this path
	// v1 "ecommerce/api/loyalty/v1"
)

// LoyaltyService is a gRPC service that implements the LoyaltyServer interface.
// It holds a reference to the business logic layer.
type LoyaltyService struct {
	// v1.UnimplementedLoyaltyServer

	uc *biz.LoyaltyUsecase
}

// NewLoyaltyService creates a new LoyaltyService.
func NewLoyaltyService(uc *biz.LoyaltyUsecase) *LoyaltyService {
	return &LoyaltyService{uc: uc}
}

// Note: The actual RPC methods like GetUserLoyaltyProfile, AddPoints, etc., will be implemented here.
// These methods will call the corresponding business logic in the 'biz' layer.

/*
Example Implementation (once gRPC code is generated):

func (s *LoyaltyService) GetUserLoyaltyProfile(ctx context.Context, req *v1.GetUserLoyaltyProfileRequest) (*v1.GetUserLoyaltyProfileResponse, error) {
    // 1. Call business logic
    profile, err := s.uc.GetUserLoyaltyProfile(ctx, req.UserId)
    if err != nil {
        return nil, err
    }

    // 2. Convert biz model to API model and return
    return &v1.GetUserLoyaltyProfileResponse{Profile: profile}, nil
}

*/

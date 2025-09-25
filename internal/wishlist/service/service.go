package service

import (
	"context"

	"ecommerce/internal/wishlist/biz"
	// Assuming the generated pb.go file will be in this path
	// v1 "ecommerce/api/wishlist/v1"
)

// WishlistService is a gRPC service that implements the WishlistServer interface.
// It holds a reference to the business logic layer.
type WishlistService struct {
	// v1.UnimplementedWishlistServer

	uc *biz.WishlistUsecase
}

// NewWishlistService creates a new WishlistService.
func NewWishlistService(uc *biz.WishlistUsecase) *WishlistService {
	return &WishlistService{uc: uc}
}

// Note: The actual RPC methods like AddItemToWishlist, ListWishlistItems, etc., will be implemented here.
// These methods will call the corresponding business logic in the 'biz' layer.

/*
Example Implementation (once gRPC code is generated):

func (s *WishlistService) AddItemToWishlist(ctx context.Context, req *v1.AddItemToWishlistRequest) (*v1.AddItemToWishlistResponse, error) {
    // 1. Call business logic
    item, err := s.uc.AddItem(ctx, req.UserId, req.ProductId)
    if err != nil {
        return nil, err
    }

    // 2. Convert biz model to API model and return
    return &v1.AddItemToWishlistResponse{Item: item}, nil
}

*/

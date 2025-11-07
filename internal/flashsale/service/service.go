package service

import (
	// Assuming the generated pb.go file will be in this path
	// v1 "ecommerce/api/flashsale/v1"
)

// FlashSaleService is a gRPC service that implements the FlashSaleServer interface.
// It holds a reference to the business logic layer.
type FlashSaleService struct {
	// v1.UnimplementedFlashSaleServer

	uc *biz.FlashSaleUsecase
}

// NewFlashSaleService creates a new FlashSaleService.
func NewFlashSaleService(uc *biz.FlashSaleUsecase) *FlashSaleService {
	return &FlashSaleService{uc: uc}
}

// Note: The actual RPC methods like CreateFlashSaleEvent, ParticipateInFlashSale, etc., will be implemented here.
// These methods will call the corresponding business logic in the 'biz' layer.

/*
Example Implementation (once gRPC code is generated):

func (s *FlashSaleService) CreateFlashSaleEvent(ctx context.Context, req *v1.CreateFlashSaleEventRequest) (*v1.CreateFlashSaleEventResponse, error) {
    // 1. Call business logic
    event, err := s.uc.CreateFlashSaleEvent(ctx, req.Name, req.Description, req.StartTime, req.EndTime, req.Products)
    if err != nil {
        return nil, err
    }

    // 2. Convert biz model to API model and return
    return &v1.CreateFlashSaleEventResponse{Event: event}, nil
}

*/

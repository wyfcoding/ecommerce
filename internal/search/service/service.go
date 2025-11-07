package service

import (
	"context"
	v1 "ecommerce/api/search/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SearchService is the gRPC service implementation for search.
type SearchService struct {
	v1.UnimplementedSearchServiceServer
	uc *biz.SearchUsecase
}

// NewSearchService creates a new SearchService.
func NewSearchService(uc *biz.SearchUsecase) *SearchService {
	return &SearchService{uc: uc}
}

// SearchProducts implements the SearchProducts RPC.
func (s *SearchService) SearchProducts(ctx context.Context, req *v1.SearchProductsRequest) (*v1.SearchProductsResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}
	if req.PageSize <= 0 {
		req.PageSize = 10 // Default page size
	}
	if req.PageToken < 0 {
		req.PageToken = 0 // Default page token
	}

	bizProducts, totalSize, err := s.uc.SearchProducts(ctx, req.Query, req.PageSize, req.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search products: %v", err)
	}

	protoProducts := make([]*v1.Product, 0, len(bizProducts))
	for _, p := range bizProducts {
		protoProducts = append(protoProducts, &v1.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			ImageUrl:    p.ImageURL,
		})
	}

	nextPageToken := req.PageToken + req.PageSize
	if nextPageToken >= totalSize {
		nextPageToken = 0 // No more pages
	}

	return &v1.SearchProductsResponse{
		Products:      protoProducts,
		TotalSize:     totalSize,
		NextPageToken: nextPageToken,
	}, nil
}

package service

import (
	"context"

	v1 "ecommerce/ecommerce/api/search/v1"
	"ecommerce/ecommerce/app/search/internal/biz"
)

type SearchService struct {
	v1.UnimplementedSearchServer
	uc *biz.SearchUsecase
}

func NewSearchService(uc *biz.SearchUsecase) *SearchService {
	return &SearchService{uc: uc}
}

func (s *SearchService) SearchProducts(ctx context.Context, req *v1.SearchProductsRequest) (*v1.SearchProductsResponse, error) {
	// 轉換為 biz 請求
	bizReq := &biz.SearchRequest{
		Keyword:    req.Keyword,
		CategoryID: req.GetCategoryId(), // GetCategoryId 處理 optional 字段
		BrandID:    req.GetBrandId(),
		SortBy:     req.SortBy,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	// 調用 Usecase
	result, err := s.uc.SearchProducts(ctx, bizReq)
	if err != nil {
		// ... 錯誤處理
	}

	// 轉換為 gRPC 響應
	hits := make([]*v1.ProductHit, len(result.Hits))
	for i, hit := range result.Hits {
		hits[i] = &v1.ProductHit{
			SpuId:     hit.SpuID,
			Title:     hit.Title,
			SubTitle:  hit.SubTitle,
			MainImage: hit.MainImage,
			Price:     hit.Price,
		}
	}

	return &v1.SearchProductsResponse{
		Hits:  hits,
		Total: result.Total,
	}, nil
}

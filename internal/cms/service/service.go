package service

import (
	"ecommerce/internal/cms/biz"
	// Assuming the generated pb.go file will be in this path
	// v1 "ecommerce/api/cms/v1"
)

// CMSService is a gRPC service that implements the CMSServer interface.
// It holds a reference to the business logic layer.
type CMSService struct {
	// v1.UnimplementedCMSServer

	uc *biz.CmsUsecase
}

// NewCMSService creates a new CMSService.
func NewCMSService(uc *biz.CmsUsecase) *CMSService {
	return &CMSService{uc: uc}
}

// Note: The actual RPC methods like CreateContentPage, GetContentPage, etc., will be implemented here.
// These methods will call the corresponding business logic in the 'biz' layer.

/*
Example Implementation (once gRPC code is generated):

func (s *CMSService) CreateContentPage(ctx context.Context, req *v1.CreateContentPageRequest) (*v1.CreateContentPageResponse, error) {
    // 1. Call business logic
    page, err := s.uc.CreateContentPage(ctx, req.Title, req.Slug, req.ContentHtml, req.Status)
    if err != nil {
        return nil, err
    }

    // 2. Convert biz model to API model and return
    return &v1.CreateContentPageResponse{Page: page}, nil
}

*/

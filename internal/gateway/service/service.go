package service

import (
	"context"
	v1 "ecommerce/api/gateway/v1"
	"ecommerce/internal/gateway/biz"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GatewayService is the gRPC service implementation for the API Gateway.
type GatewayService struct {
	v1.UnimplementedGatewayServiceServer
	uc *biz.GatewayUsecase
}

// NewGatewayService creates a new GatewayService.
func NewGatewayService(uc *biz.GatewayUsecase) *GatewayService {
	return &GatewayService{uc: uc}
}

// HandleRequest implements the HandleRequest RPC.
func (s *GatewayService) HandleRequest(ctx context.Context, req *v1.HandleRequestRequest) (*v1.HandleRequestResponse, error) {
	if req.Method == "" || req.Path == "" {
		return nil, status.Error(codes.InvalidArgument, "method and path are required")
	}

	statusCode, responseBody, responseHeaders, err := s.uc.HandleRequest(ctx, req.Method, req.Path, req.Body, req.AccessToken, req.Headers)
	if err != nil {
		if errors.Is(err, biz.ErrUnauthorized) {
			return nil, status.Errorf(codes.Unauthenticated, "unauthorized: %v", err)
		}
		if errors.Is(err, biz.ErrServiceNotFound) {
			return nil, status.Errorf(codes.NotFound, "service not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to handle request: %v", err)
	}

	return &v1.HandleRequestResponse{
		StatusCode:   statusCode,
		ResponseBody: responseBody,
		Headers:      responseHeaders,
	}, nil
}

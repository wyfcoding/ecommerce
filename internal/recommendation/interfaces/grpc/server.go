package grpc

import (
	"context"
	pb "ecommerce/api/recommendation/v1"
	"ecommerce/internal/recommendation/application"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedRecommendationServiceServer
	app *application.RecommendationService
}

func NewServer(app *application.RecommendationService) *Server {
	return &Server{app: app}
}

func (s *Server) GetRecommendedProducts(ctx context.Context, req *pb.GetRecommendedProductsRequest) (*pb.GetRecommendedProductsResponse, error) {
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	limit := int(req.Count)
	if limit < 1 {
		limit = 10
	}

	// Trigger generation (optional, or rely on background job)
	_ = s.app.GenerateRecommendations(ctx, userID)

	recs, err := s.app.GetRecommendations(ctx, userID, "", limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbProducts := make([]*pb.Product, len(recs))
	for i, r := range recs {
		// Note: Recommendation entity has ProductID, but not full Product details.
		// Proto expects full Product details.
		// We should fetch product details from Product Service.
		// For now, returning minimal info with ID.
		pbProducts[i] = &pb.Product{
			Id:          strconv.FormatUint(r.ProductID, 10),
			Name:        "Product " + strconv.FormatUint(r.ProductID, 10), // Placeholder
			Description: r.Reason,
			Price:       0,  // Unknown
			ImageUrl:    "", // Unknown
		}
	}

	return &pb.GetRecommendedProductsResponse{
		Products: pbProducts,
	}, nil
}

func (s *Server) IndexProductRelationship(ctx context.Context, req *pb.IndexProductRelationshipRequest) (*pb.IndexProductRelationshipResponse, error) {
	return nil, status.Error(codes.Unimplemented, "IndexProductRelationship not implemented")
}

func (s *Server) GetGraphRecommendedProducts(ctx context.Context, req *pb.GetGraphRecommendedProductsRequest) (*pb.GetGraphRecommendedProductsResponse, error) {
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id")
	}

	limit := int(req.Count)
	if limit < 1 {
		limit = 10
	}

	sims, err := s.app.GetSimilarProducts(ctx, productID, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbProducts := make([]*pb.Product, len(sims))
	for i, sim := range sims {
		pbProducts[i] = &pb.Product{
			Id:          strconv.FormatUint(sim.SimilarProductID, 10),
			Name:        "Product " + strconv.FormatUint(sim.SimilarProductID, 10), // Placeholder
			Description: "Score: " + strconv.FormatFloat(sim.Similarity, 'f', 2, 64),
			Price:       0,
			ImageUrl:    "",
		}
	}

	return &pb.GetGraphRecommendedProductsResponse{
		Products: pbProducts,
	}, nil
}

func (s *Server) GetAdvancedRecommendedProducts(ctx context.Context, req *pb.GetAdvancedRecommendedProductsRequest) (*pb.GetAdvancedRecommendedProductsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetAdvancedRecommendedProducts not implemented")
}

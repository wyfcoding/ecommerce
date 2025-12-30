package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。
	"strconv" // 导入字符串转换工具。

	pb "github.com/wyfcoding/ecommerce/goapi/recommendation/v1"          // 导入推荐模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/recommendation/application" // 导入推荐模块的应用服务。

	// 导入推荐模块的领域层。
	"google.golang.org/grpc/codes"  // gRPC状态码。
	"google.golang.org/grpc/status" // gRPC状态处理。
)

// Server 结构体实现了 RecommendationService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedRecommendationServiceServer                                    // 嵌入生成的UnimplementedRecommendationServiceServer，确保前向兼容性。
	app                                         *application.RecommendationService // 依赖Recommendation应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Recommendation gRPC 服务端实例。
func NewServer(app *application.RecommendationService) *Server {
	return &Server{app: app}
}

// GetRecommendedProducts 处理获取推荐商品列表的gRPC请求。
func (s *Server) GetRecommendedProducts(ctx context.Context, req *pb.GetRecommendedProductsRequest) (*pb.GetRecommendedProductsResponse, error) {
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user_id: %v", err))
	}

	limit := int(req.Count)
	if limit < 1 {
		limit = 10
	}

	_ = s.app.GenerateRecommendations(ctx, userID)

	recs, err := s.app.GetRecommendations(ctx, userID, "", limit)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get recommendations: %v", err))
	}

	pbProducts := make([]*pb.Product, len(recs))
	for i, r := range recs {
		pbProducts[i] = &pb.Product{
			Id:          strconv.FormatUint(r.ProductID, 10),
			Name:        "Product " + strconv.FormatUint(r.ProductID, 10),
			Description: r.Reason,
			Price:       0,
			ImageUrl:    "",
		}
	}

	return &pb.GetRecommendedProductsResponse{
		Products: pbProducts,
	}, nil
}

// IndexProductRelationship 处理索引商品关系的gRPC请求。
func (s *Server) IndexProductRelationship(ctx context.Context, req *pb.IndexProductRelationshipRequest) (*pb.IndexProductRelationshipResponse, error) {
	return nil, status.Error(codes.Unimplemented, "IndexProductRelationship not implemented")
}

// GetGraphRecommendedProducts 处理获取图推荐商品列表的gRPC请求。
func (s *Server) GetGraphRecommendedProducts(ctx context.Context, req *pb.GetGraphRecommendedProductsRequest) (*pb.GetGraphRecommendedProductsResponse, error) {
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id: %v", err))
	}

	limit := int(req.Count)
	if limit < 1 {
		limit = 10
	}

	sims, err := s.app.GetSimilarProducts(ctx, productID, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get similar products: %v", err))
	}

	pbProducts := make([]*pb.Product, len(sims))
	for i, sim := range sims {
		pbProducts[i] = &pb.Product{
			Id:          strconv.FormatUint(sim.SimilarProductID, 10),
			Name:        "Product " + strconv.FormatUint(sim.SimilarProductID, 10),
			Description: "Score: " + strconv.FormatFloat(sim.Similarity, 'f', 2, 64),
			Price:       0,
			ImageUrl:    "",
		}
	}

	return &pb.GetGraphRecommendedProductsResponse{
		Products: pbProducts,
	}, nil
}

// GetAdvancedRecommendedProducts 处理获取高级推荐商品列表的gRPC请求。
func (s *Server) GetAdvancedRecommendedProducts(ctx context.Context, req *pb.GetAdvancedRecommendedProductsRequest) (*pb.GetAdvancedRecommendedProductsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetAdvancedRecommendedProducts not implemented")
}

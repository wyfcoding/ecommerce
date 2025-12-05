package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。
	"strconv" // 导入字符串转换工具。

	pb "github.com/wyfcoding/ecommerce/go-api/recommendation/v1"         // 导入推荐模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/recommendation/application" // 导入推荐模块的应用服务。
	// "github.com/wyfcoding/ecommerce/internal/recommendation/domain/entity" // 导入推荐模块的领域实体，此处未直接使用，通过应用服务层转换。

	"google.golang.org/grpc/codes"  // gRPC状态码。
	"google.golang.org/grpc/status" // gRPC状态处理。
)

// Server 结构体实现了 RecommendationService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedRecommendationServiceServer                                    // 嵌入生成的UnimplementedRecommendationServiceServer，确保前向兼容性。
	app                                         *application.RecommendationService // 依赖Recommendation应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Recommendation gRPC 服务端实例。
func NewServer(app *application.RecommendationService) *Server {
	return &Server{app: app}
}

// GetRecommendedProducts 处理获取推荐商品列表的gRPC请求。
// req: 包含用户ID和推荐数量的请求体。
// 返回推荐商品列表响应和可能发生的gRPC错误。
func (s *Server) GetRecommendedProducts(ctx context.Context, req *pb.GetRecommendedProductsRequest) (*pb.GetRecommendedProductsResponse, error) {
	// 转换用户ID。
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user_id: %v", err))
	}

	// 设置推荐数量限制。
	limit := int(req.Count)
	if limit < 1 {
		limit = 10 // 默认推荐数量。
	}

	// 触发推荐生成（可选）。
	// 在实际系统中，推荐生成可能是一个异步的后台任务，而不是同步调用。
	// 此处即使失败也不会影响后续获取现有推荐的操作。
	_ = s.app.GenerateRecommendations(ctx, userID)

	// 调用应用服务层获取推荐列表。
	// recType 参数为空字符串，表示不按特定类型过滤（获取所有推荐类型）。
	recs, err := s.app.GetRecommendations(ctx, userID, "", limit)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get recommendations: %v", err))
	}

	// 将推荐结果转换为protobuf的Product消息列表。
	pbProducts := make([]*pb.Product, len(recs))
	for i, r := range recs {
		// 注意：Recommendation 实体只包含 ProductID，而 Proto 期望完整的 Product 详情。
		// 在实际系统中，这里应该通过调用产品服务（Product Service）来获取完整的商品信息。
		// 目前为了演示，使用占位符填充商品名称、价格和图片URL。
		pbProducts[i] = &pb.Product{
			Id:          strconv.FormatUint(r.ProductID, 10),              // 商品ID。
			Name:        "Product " + strconv.FormatUint(r.ProductID, 10), // 占位符：商品名称。
			Description: r.Reason,                                         // 使用推荐理由作为描述。
			Price:       0,                                                // 占位符：价格未知。
			ImageUrl:    "",                                               // 占位符：图片URL未知。
		}
	}

	return &pb.GetRecommendedProductsResponse{
		Products: pbProducts,
	}, nil
}

// IndexProductRelationship 处理索引商品关系的gRPC请求。
// req: 包含商品关系信息的请求体。
// 当前未实现。
func (s *Server) IndexProductRelationship(ctx context.Context, req *pb.IndexProductRelationshipRequest) (*pb.IndexProductRelationshipResponse, error) {
	return nil, status.Error(codes.Unimplemented, "IndexProductRelationship not implemented")
}

// GetGraphRecommendedProducts 处理获取图推荐商品列表的gRPC请求。
// req: 包含商品ID和推荐数量的请求体。
// 返回推荐商品列表响应和可能发生的gRPC错误。
func (s *Server) GetGraphRecommendedProducts(ctx context.Context, req *pb.GetGraphRecommendedProductsRequest) (*pb.GetGraphRecommendedProductsResponse, error) {
	// 转换商品ID。
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id: %v", err))
	}

	// 设置推荐数量限制。
	limit := int(req.Count)
	if limit < 1 {
		limit = 10
	}

	// 调用应用服务层获取相似商品列表。
	sims, err := s.app.GetSimilarProducts(ctx, productID, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get similar products: %v", err))
	}

	// 将相似商品结果转换为protobuf的Product消息列表。
	pbProducts := make([]*pb.Product, len(sims))
	for i, sim := range sims {
		// 注意：ProductSimilarity 实体只包含 SimilarProductID，而 Proto 期望完整的 Product 详情。
		// 在实际系统中，这里应该通过调用产品服务（Product Service）来获取完整的商品信息。
		// 目前为了演示，使用占位符填充商品名称、价格和图片URL。
		pbProducts[i] = &pb.Product{
			Id:          strconv.FormatUint(sim.SimilarProductID, 10),                // 相似商品ID。
			Name:        "Product " + strconv.FormatUint(sim.SimilarProductID, 10),   // 占位符：商品名称。
			Description: "Score: " + strconv.FormatFloat(sim.Similarity, 'f', 2, 64), // 相似度作为描述。
			Price:       0,                                                           // 占位符：价格未知。
			ImageUrl:    "",                                                          // 占位符：图片URL未知。
		}
	}

	return &pb.GetGraphRecommendedProductsResponse{
		Products: pbProducts,
	}, nil
}

// GetAdvancedRecommendedProducts 处理获取高级推荐商品列表的gRPC请求。
// req: 包含高级推荐参数的请求体。
// 当前未实现。
func (s *Server) GetAdvancedRecommendedProducts(ctx context.Context, req *pb.GetAdvancedRecommendedProductsRequest) (*pb.GetAdvancedRecommendedProductsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetAdvancedRecommendedProducts not implemented")
}

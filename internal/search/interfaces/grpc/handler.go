package grpc

import (
	"context"       // 导入上下文。
	"encoding/json" // 导入JSON编码/解码库。
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/search/v1"          // 导入搜索模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/search/application" // 导入搜索模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/search/domain"      // 导入搜索模块的领域层。

	"google.golang.org/grpc/codes"  // gRPC状态码。
	"google.golang.org/grpc/status" // gRPC状态处理。
)

// Server 结构体实现了 Search 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedSearchServiceServer                     // 嵌入生成的UnimplementedSearchServiceServer，确保前向兼容性。
	app                                 *application.Search // 依赖Search应用服务，处理核心业务逻辑。
	logger                              *slog.Logger        // 日志记录器。
}

// NewServer 创建并返回一个新的 Search gRPC 服务端实例。
func NewServer(app *application.Search, logger *slog.Logger) *Server {
	return &Server{
		app:    app,
		logger: logger,
	}
}

// SearchProducts 处理搜索商品的gRPC请求。
// req: 包含查询关键词、分页参数的请求体。
// 返回搜索结果响应和可能发生的gRPC错误。
func (s *Server) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "gRPC SearchProducts received", "query", req.Query, "page_token", req.PageToken)

	// 获取分页参数。PageToken在gRPC中常用于表示下一页的标识，这里简化为页码。
	page := max(int(req.PageToken), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 构建SearchFilter实体。
	filter := &domain.SearchFilter{
		Keyword:  req.Query,
		Page:     page,
		PageSize: pageSize,
	}

	// 调用应用服务层执行搜索。
	result, err := s.app.Search(ctx, 0, filter)
	if err != nil {
		s.logger.ErrorContext(ctx, "gRPC SearchProducts failed", "query", req.Query, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to search products: %v", err))
	}

	// 将搜索结果中的通用项（interface{}）转换为protobuf的Product消息。
	pbProducts := make([]*pb.Product, 0, len(result.Items))
	for _, item := range result.Items {
		bytes, err := json.Marshal(item)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to marshal search result item to JSON", "item", item, "error", err)
			continue
		}

		var p pb.Product
		if err := json.Unmarshal(bytes, &p); err != nil {
			s.logger.ErrorContext(ctx, "failed to unmarshal JSON to pb.Product", "json_bytes", string(bytes), "error", err)
			continue
		}
		pbProducts = append(pbProducts, &p)
	}

	s.logger.InfoContext(ctx, "gRPC SearchProducts successful", "query", req.Query, "count", len(pbProducts), "duration", time.Since(start))
	return &pb.SearchProductsResponse{
		Products:      pbProducts,
		TotalSize:     int32(result.Total), // 搜索结果总数。
		NextPageToken: int32(page + 1),     // 建议的下一页页码。
	}, nil
}

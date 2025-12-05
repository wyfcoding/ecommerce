package grpc

import (
	"context"       // 导入上下文。
	"encoding/json" // 导入JSON编码/解码库。
	"fmt"           // 导入格式化库。

	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/search/v1"           // 导入搜索模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/search/application"   // 导入搜索模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/search/domain/entity" // 导入搜索模块的领域实体。

	"google.golang.org/grpc/codes"  // gRPC状态码。
	"google.golang.org/grpc/status" // gRPC状态处理。
	// "google.golang.org/protobuf/types/known/emptypb"      // 导入空消息类型，此文件中未直接使用。
	// "google.golang.org/protobuf/types/known/timestamppb"  // 导入时间戳消息类型，此文件中未直接使用。
)

// Server 结构体实现了 SearchService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedSearchServiceServer                            // 嵌入生成的UnimplementedSearchServiceServer，确保前向兼容性。
	app                                 *application.SearchService // 依赖Search应用服务，处理核心业务逻辑。
	logger                              *slog.Logger               // 日志记录器。
}

// NewServer 创建并返回一个新的 Search gRPC 服务端实例。
func NewServer(app *application.SearchService, logger *slog.Logger) *Server {
	return &Server{
		app:    app,
		logger: logger,
	}
}

// SearchProducts 处理搜索商品的gRPC请求。
// req: 包含查询关键词、分页参数的请求体。
// 返回搜索结果响应和可能发生的gRPC错误。
func (s *Server) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	// 获取分页参数。PageToken在gRPC中常用于表示下一页的标识，这里简化为页码。
	page := int(req.PageToken)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 构建SearchFilter实体。
	filter := &entity.SearchFilter{
		Keyword:  req.Query,
		Page:     page,
		PageSize: pageSize,
	}

	// 调用应用服务层执行搜索。
	// 用户ID为0表示匿名用户，实际项目中可能需要从上下文（例如JWT token）中获取用户ID。
	result, err := s.app.Search(ctx, 0, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to search products: %v", err))
	}

	// 将搜索结果中的通用项（interface{}）转换为protobuf的Product消息。
	// 这是一个灵活但可能低效的方式，因为需要进行JSON的编解码。
	// 理想情况下，如果搜索结果的类型是确定的（例如都是Product实体），可以直接进行类型断言和转换。
	pbProducts := make([]*pb.Product, 0, len(result.Items))
	for _, item := range result.Items {
		// 先将interface{}类型的数据编码为JSON字节数组。
		bytes, err := json.Marshal(item)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to marshal search result item to JSON", "item", item, "error", err)
			continue // 忽略单个项的转换错误，继续处理其他项。
		}

		var p pb.Product
		// 再将JSON字节数组解码为pb.Product消息。
		if err := json.Unmarshal(bytes, &p); err != nil {
			s.logger.ErrorContext(ctx, "failed to unmarshal JSON to pb.Product", "json_bytes", string(bytes), "error", err)
			continue // 忽略单个项的转换错误。
		}
		pbProducts = append(pbProducts, &p)
	}

	return &pb.SearchProductsResponse{
		Products:      pbProducts,
		TotalSize:     int32(result.Total), // 搜索结果总数。
		NextPageToken: int32(page + 1),     // 建议的下一页页码。
	}, nil
}

// TODO: 补充实现 GetHotKeywords, GetSearchHistory, ClearSearchHistory, Suggest 等搜索相关接口。

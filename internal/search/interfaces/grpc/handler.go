package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/search/v1"
	"github.com/wyfcoding/ecommerce/internal/search/application"
	"github.com/wyfcoding/ecommerce/internal/search/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server 实现 Search 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedSearchServiceServer
	app    *application.Search
	logger *slog.Logger
}

// NewServer 创建并返回 Search gRPC 服务端实例。
func NewServer(app *application.Search, logger *slog.Logger) *Server {
	return &Server{app: app, logger: logger}
}

// SearchProducts 处理商品搜索请求。
func (s *Server) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	start := time.Now()
	page := normalizePage(req.PageToken)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	filter := &domain.SearchFilter{
		Keyword:  req.Query,
		Page:     page,
		PageSize: pageSize,
		Sort:     "relevance",
	}

	result, err := s.app.Search(ctx, 0, filter)
	if err != nil {
		s.logger.ErrorContext(ctx, "gRPC SearchProducts failed", "query", req.Query, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to search products: %v", err))
	}

	pbProducts := make([]*pb.Product, 0, len(result.Items))
	for _, item := range result.Items {
		pbProduct, err := convertToPbProduct(item)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to convert search result item", "error", err)
			continue
		}
		pbProducts = append(pbProducts, pbProduct)
	}

	return &pb.SearchProductsResponse{
		Products:      pbProducts,
		TotalSize:     int32(result.Total),
		NextPageToken: int32(page + 1),
	}, nil
}

func convertToPbProduct(item any) (*pb.Product, error) {
	bytes, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, err
	}

	return &pb.Product{
		Id:          toString(raw["id"]),
		Name:        toString(raw["name"]),
		Description: toString(raw["description"]),
		Price:       toFloat64(raw["price"]),
		ImageUrl:    toString(raw["image_url"]),
	}, nil
}

func normalizePage(pageToken int32) int {
	if pageToken <= 0 {
		return 1
	}
	return int(pageToken)
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	case float64:
		return strconv.FormatUint(uint64(t), 10)
	case float32:
		return strconv.FormatUint(uint64(t), 10)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case uint64:
		return strconv.FormatUint(t, 10)
	case uint32:
		return strconv.FormatUint(uint64(t), 10)
	case uint:
		return strconv.FormatUint(uint64(t), 10)
	default:
		return ""
	}
}

func toFloat64(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case int32:
		return float64(t)
	case uint:
		return float64(t)
	case uint64:
		return float64(t)
	case uint32:
		return float64(t)
	case json.Number:
		f, err := t.Float64()
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

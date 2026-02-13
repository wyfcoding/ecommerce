package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/search/v1"
	"github.com/wyfcoding/ecommerce/internal/search/application"
	"github.com/wyfcoding/ecommerce/internal/search/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server 结构体实现了 Search 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedSearchServiceServer
	app    *application.Search
	logger *slog.Logger
}

// NewServer 创建并返回一个新的 Search gRPC 服务端实例。
func NewServer(app *application.Search, logger *slog.Logger) *Server {
	return &Server{
		app:    app,
		logger: logger,
	}
}

// SearchProducts 处理搜索商品的gRPC请求（基础版）。
func (s *Server) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "gRPC SearchProducts received", "query", req.Query, "page_token", req.PageToken)

	page := max(int(req.PageToken), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	filter := &domain.SearchFilter{
		Keyword:  req.Query,
		Page:     page,
		PageSize: pageSize,
		Sort:     convertSortOrder(req.SortOrder),
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
			s.logger.ErrorContext(ctx, "failed to convert search result item", "item", item, "error", err)
			continue
		}
		pbProducts = append(pbProducts, pbProduct)
	}

	s.logger.InfoContext(ctx, "gRPC SearchProducts successful", "query", req.Query, "count", len(pbProducts), "duration", time.Since(start))
	return &pb.SearchProductsResponse{
		Products:      pbProducts,
		TotalSize:     int32(result.Total),
		NextPageToken: int32(page + 1),
		TookMs:        time.Since(start).Milliseconds(),
	}, nil
}

// AdvancedSearch 处理高级搜索请求，支持多维度筛选和聚合统计。
func (s *Server) AdvancedSearch(ctx context.Context, req *pb.AdvancedSearchRequest) (*pb.AdvancedSearchResponse, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "gRPC AdvancedSearch received", "query", req.Query, "user_id", req.UserId)

	page := max(int(req.PageToken), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	filter := &domain.SearchFilter{
		Keyword:    req.Query,
		Page:       page,
		PageSize:   pageSize,
		Sort:       convertSortOrder(req.SortOrder),
		CategoryID: uint64(getFirstOrZero(req.CategoryIds)),
		BrandID:    uint64(getFirstOrZero(req.BrandIds)),
		Tags:       req.Tags,
	}

	if req.PriceRange != nil {
		filter.PriceMin = float64(req.PriceRange.MinPrice) / 100
		filter.PriceMax = float64(req.PriceRange.MaxPrice) / 100
	}

	result, err := s.app.Search(ctx, uint64(req.UserId), filter)
	if err != nil {
		s.logger.ErrorContext(ctx, "gRPC AdvancedSearch failed", "query", req.Query, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to advanced search: %v", err))
	}

	pbProducts := make([]*pb.Product, 0, len(result.Items))
	for _, item := range result.Items {
		pbProduct, err := convertToPbProduct(item)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to convert search result item", "item", item, "error", err)
			continue
		}
		pbProducts = append(pbProducts, pbProduct)
	}

	s.logger.InfoContext(ctx, "gRPC AdvancedSearch successful", "query", req.Query, "count", len(pbProducts), "duration", time.Since(start))
	return &pb.AdvancedSearchResponse{
		Products:      pbProducts,
		TotalSize:     result.Total,
		NextPageToken: int32(page + 1),
		TookMs:        time.Since(start).Milliseconds(),
		HasResults:    result.Total > 0,
	}, nil
}

// GetSearchSuggestions 获取搜索建议。
func (s *Server) GetSearchSuggestions(ctx context.Context, req *pb.GetSearchSuggestionsRequest) (*pb.GetSearchSuggestionsResponse, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "gRPC GetSearchSuggestions received", "prefix", req.Prefix, "user_id", req.UserId)

	limit := int(req.Limit)
	if limit < 1 {
		limit = 10
	}

	suggestions, err := s.app.Suggest(ctx, req.Prefix)
	if err != nil {
		s.logger.ErrorContext(ctx, "gRPC GetSearchSuggestions failed", "prefix", req.Prefix, "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get suggestions: %v", err))
	}

	pbSuggestions := make([]*pb.SearchSuggestion, 0, len(suggestions))
	for _, sug := range suggestions {
		pbSuggestions = append(pbSuggestions, &pb.SearchSuggestion{
			Keyword:       sug.Keyword,
			Type:          convertSuggestionType(sug.Type),
			Score:         int32(sug.Score),
			Highlighted:   highlightKeyword(sug.Keyword, req.Prefix),
			ProductCount:  0,
			CategoryId:    0,
			CategoryName:  "",
		})
	}

	if uint64(req.UserId) > 0 {
		history, err := s.app.GetSearchHistory(ctx, uint64(req.UserId), 3)
		if err == nil && len(history) > 0 {
			historySuggestions := make([]*pb.SearchSuggestion, 0, len(history))
			for _, h := range history {
				historySuggestions = append(historySuggestions, &pb.SearchSuggestion{
					Keyword: h.Keyword,
					Type:    pb.SuggestionType_SUGGESTION_TYPE_HISTORY,
					Score:   1000,
				})
			}
			pbSuggestions = append(historySuggestions, pbSuggestions...)
		}
	}

	if len(pbSuggestions) > limit {
		pbSuggestions = pbSuggestions[:limit]
	}

	s.logger.InfoContext(ctx, "gRPC GetSearchSuggestions successful", "prefix", req.Prefix, "count", len(pbSuggestions), "duration", time.Since(start))
	return &pb.GetSearchSuggestionsResponse{
		Suggestions: pbSuggestions,
	}, nil
}

// GetHotKeywords 获取热门搜索词。
func (s *Server) GetHotKeywords(ctx context.Context, req *pb.GetHotKeywordsRequest) (*pb.GetHotKeywordsResponse, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "gRPC GetHotKeywords received", "limit", req.Limit)

	limit := int(req.Limit)
	if limit < 1 {
		limit = 10
	}

	hotKeywords, err := s.app.GetHotKeywords(ctx, limit)
	if err != nil {
		s.logger.ErrorContext(ctx, "gRPC GetHotKeywords failed", "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get hot keywords: %v", err))
	}

	pbKeywords := make([]*pb.HotKeyword, 0, len(hotKeywords))
	for i, kw := range hotKeywords {
		pbKeywords = append(pbKeywords, &pb.HotKeyword{
			Keyword:     kw.Keyword,
			SearchCount: int64(kw.SearchCount),
			Rank:        int32(i + 1),
			Trend:       pb.KeywordTrend_KEYWORD_TREND_STABLE,
		})
	}

	s.logger.InfoContext(ctx, "gRPC GetHotKeywords successful", "count", len(pbKeywords), "duration", time.Since(start))
	return &pb.GetHotKeywordsResponse{
		Keywords: pbKeywords,
	}, nil
}

// GetSearchHistory 获取用户搜索历史。
func (s *Server) GetSearchHistory(ctx context.Context, req *pb.GetSearchHistoryRequest) (*pb.GetSearchHistoryResponse, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "gRPC GetSearchHistory received", "user_id", req.UserId)

	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	limit := int(req.Limit)
	if limit < 1 {
		limit = 20
	}

	history, err := s.app.GetSearchHistory(ctx, uint64(req.UserId), limit)
	if err != nil {
		s.logger.ErrorContext(ctx, "gRPC GetSearchHistory failed", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get search history: %v", err))
	}

	pbItems := make([]*pb.SearchHistoryItem, 0, len(history))
	for _, h := range history {
		pbItems = append(pbItems, &pb.SearchHistoryItem{
			Keyword:     h.Keyword,
			SearchedAt:  h.Timestamp.Format(time.RFC3339),
			ResultCount: 0,
		})
	}

	s.logger.InfoContext(ctx, "gRPC GetSearchHistory successful", "user_id", req.UserId, "count", len(pbItems), "duration", time.Since(start))
	return &pb.GetSearchHistoryResponse{
		Items: pbItems,
	}, nil
}

// ClearSearchHistory 清空用户搜索历史。
func (s *Server) ClearSearchHistory(ctx context.Context, req *pb.ClearSearchHistoryRequest) (*pb.ClearSearchHistoryResponse, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "gRPC ClearSearchHistory received", "user_id", req.UserId)

	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if err := s.app.ClearSearchHistory(ctx, uint64(req.UserId)); err != nil {
		s.logger.ErrorContext(ctx, "gRPC ClearSearchHistory failed", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to clear search history: %v", err))
	}

	s.logger.InfoContext(ctx, "gRPC ClearSearchHistory successful", "user_id", req.UserId, "duration", time.Since(start))
	return &pb.ClearSearchHistoryResponse{
		Success: true,
	}, nil
}

// GetSearchAggregations 获取搜索聚合统计信息。
func (s *Server) GetSearchAggregations(ctx context.Context, req *pb.GetSearchAggregationsRequest) (*pb.GetSearchAggregationsResponse, error) {
	start := time.Now()
	s.logger.InfoContext(ctx, "gRPC GetSearchAggregations received", "query", req.Query)

	filter := &domain.SearchFilter{
		Keyword:    req.Query,
		CategoryID: uint64(req.CategoryId),
		Page:       1,
		PageSize:   0,
	}

	result, err := s.app.Search(ctx, 0, filter)
	if err != nil {
		s.logger.ErrorContext(ctx, "gRPC GetSearchAggregations failed", "query", req.Query, "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get aggregations: %v", err))
	}

	aggregations := &pb.SearchAggregations{
		Brands:       []*pb.BrandAggregation{},
		Categories:   []*pb.CategoryAggregation{},
		PriceBuckets: []*pb.PriceBucketAggregation{},
		Attributes:   []*pb.AttributeAggregation{},
	}

	if aggMap, ok := result.Items.([]map[string]any); ok && len(aggMap) > 0 {
		for _, agg := range aggMap {
			if brandAgg, ok := agg["brands"].([]map[string]any); ok {
				for _, b := range brandAgg {
					aggregations.Brands = append(aggregations.Brands, &pb.BrandAggregation{
						BrandId:   int64(b["brand_id"].(float64)),
						BrandName: b["brand_name"].(string),
						Count:     int64(b["count"].(float64)),
					})
				}
			}
		}
	}

	s.logger.InfoContext(ctx, "gRPC GetSearchAggregations successful", "query", req.Query, "duration", time.Since(start))
	return &pb.GetSearchAggregationsResponse{
		Aggregations: aggregations,
	}, nil
}

// convertToPbProduct 将搜索结果项转换为 protobuf Product。
func convertToPbProduct(item any) (*pb.Product, error) {
	bytes, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	var p pb.Product
	if err := json.Unmarshal(bytes, &p); err != nil {
		return nil, err
	}

	return &p, nil
}

// convertSortOrder 将 protobuf 排序枚举转换为领域排序字符串。
func convertSortOrder(order pb.SortOrder) string {
	switch order {
	case pb.SortOrder_SORT_ORDER_PRICE_ASC:
		return "price_asc"
	case pb.SortOrder_SORT_ORDER_PRICE_DESC:
		return "price_desc"
	case pb.SortOrder_SORT_ORDER_SALES_DESC:
		return "sales_desc"
	case pb.SortOrder_SORT_ORDER_NEWEST:
		return "newest"
	case pb.SortOrder_SORT_ORDER_RATING_DESC:
		return "rating_desc"
	default:
		return "relevance"
	}
}

// convertSuggestionType 将领域建议类型转换为 protobuf 枚举。
func convertSuggestionType(t string) pb.SuggestionType {
	switch t {
	case "history":
		return pb.SuggestionType_SUGGESTION_TYPE_HISTORY
	case "hot":
		return pb.SuggestionType_SUGGESTION_TYPE_HOT
	case "correction":
		return pb.SuggestionType_SUGGESTION_TYPE_CORRECTION
	case "related":
		return pb.SuggestionType_SUGGESTION_TYPE_RELATED
	case "category":
		return pb.SuggestionType_SUGGESTION_TYPE_CATEGORY
	default:
		return pb.SuggestionType_SUGGESTION_TYPE_COMPLETION
	}
}

// highlightKeyword 高亮关键词匹配部分。
func highlightKeyword(keyword, prefix string) string {
	if len(prefix) == 0 || len(keyword) < len(prefix) {
		return keyword
	}
	if keyword[:len(prefix)] == prefix {
		return "<em>" + prefix + "</em>" + keyword[len(prefix):]
	}
	return keyword
}

// getFirstOrZero 获取切片第一个元素或返回0。
func getFirstOrZero(ids []int64) int64 {
	if len(ids) > 0 {
		return ids[0]
	}
	return 0
}

package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。
	"strconv" // 导入字符串转换工具。

	pb "github.com/wyfcoding/ecommerce/go-api/wishlist/v1"         // 导入收藏夹模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/wishlist/application" // 导入收藏夹模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/wishlist/domain"      // 导入收藏夹模块的领域层。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 WishlistService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedWishlistServer                              // 嵌入生成的UnimplementedWishlistServer，确保前向兼容性。
	app                            *application.WishlistService // 依赖Wishlist应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Wishlist gRPC 服务端实例。
func NewServer(app *application.WishlistService) *Server {
	return &Server{app: app}
}

// AddItemToWishlist 处理将商品添加到收藏夹的gRPC请求。
// req: 包含用户ID和商品ID的请求体。
// 返回添加成功的收藏夹条目响应和可能发生的gRPC错误。
func (s *Server) AddItemToWishlist(ctx context.Context, req *pb.AddItemToWishlistRequest) (*pb.AddItemToWishlistResponse, error) {
	// 将字符串类型的用户ID和商品ID转换为uint64。
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user_id: %v", err))
	}
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id: %v", err))
	}

	// 映射 Proto 字段到应用服务层.
	// Proto 暂时不包含 SkuID/ProductName/Price/ImageURL, 此处传默认值.
	skuID := productID
	item, err := s.app.Add(ctx, userID, productID, skuID, "Unknown Product", "Unknown SKU", 0, "") // Price 默认为 0，ImageURL 默认为空。
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add item to wishlist: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.AddItemToWishlistResponse{
		Item: s.toProto(item),
	}, nil
}

// RemoveItemFromWishlist 处理从收藏夹移除商品的gRPC请求。
// req: 包含用户ID和商品ID的请求体。
// 返回移除结果响应和可能发生的gRPC错误。
func (s *Server) RemoveItemFromWishlist(ctx context.Context, req *pb.RemoveItemFromWishlistRequest) (*pb.RemoveItemFromWishlistResponse, error) {
	// 将字符串类型的用户ID和商品ID转换为uint64。
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user_id: %v", err))
	}
	skuID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id for removal: %v", err))
	}

	err = s.app.RemoveByProduct(ctx, userID, skuID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove item from wishlist: %v", err))
	}

	return &pb.RemoveItemFromWishlistResponse{
		Success: true,
		Message: "Item removed",
	}, nil
}

// ListWishlistItems 处理列出收藏夹商品的gRPC请求。
// req: 包含用户ID和分页参数的请求体。
// 返回收藏夹商品列表响应和可能发生的gRPC错误。
func (s *Server) ListWishlistItems(ctx context.Context, req *pb.ListWishlistItemsRequest) (*pb.ListWishlistItemsResponse, error) {
	// 将字符串类型的用户ID转换为uint64。
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user_id: %v", err))
	}

	// 将PageToken作为页码进行简单处理。
	page := int(req.PageToken)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取收藏夹列表。
	items, total, err := s.app.List(ctx, userID, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list wishlist items: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbItems := make([]*pb.WishlistItem, len(items))
	for i, item := range items {
		pbItems[i] = s.toProto(item)
	}

	return &pb.ListWishlistItemsResponse{
		Items:      pbItems,
		TotalCount: int32(total),
	}, nil
}

// toProto 是一个辅助函数，将领域层的 Wishlist 实体转换为 protobuf 的 WishlistItem 消息。
func (s *Server) toProto(item *domain.Wishlist) *pb.WishlistItem {
	if item == nil {
		return nil
	}
	return &pb.WishlistItem{
		Id:        strconv.FormatUint(uint64(item.ID), 10), // 收藏夹条目ID。
		UserId:    strconv.FormatUint(item.UserID, 10),     // 用户ID。
		ProductId: strconv.FormatUint(item.ProductID, 10),  // 商品ID。
		// 备注：Proto中的 WishlistItem 定义缺少 SkuID, ProductName, Price, ImageURL 等字段。
		// CreatedAt 字段在Proto中命名为 AddedAt。
		AddedAt: timestamppb.New(item.CreatedAt), // 添加时间。
	}
}

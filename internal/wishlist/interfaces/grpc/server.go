package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。
	"strconv" // 导入字符串转换工具。

	pb "github.com/wyfcoding/ecommerce/api/wishlist/v1"              // 导入收藏夹模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/wishlist/application"   // 导入收藏夹模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/wishlist/domain/entity" // 导入收藏夹模块的领域实体。

	"google.golang.org/grpc/codes"  // gRPC状态码。
	"google.golang.org/grpc/status" // gRPC状态处理。
	// "google.golang.org/protobuf/types/known/emptypb"      // 导入空消息类型，此文件中未直接使用。
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

	// 备注：Proto请求（pb.AddItemToWishlistRequest）中缺少 SkuID, ProductName, SkuName, Price, ImageURL 等字段。
	// 但应用服务层的 Add 方法需要这些字段。
	// 当前实现暂时使用 ProductID 作为 SkuID，并使用占位符填充其他缺失字段。
	// 这是Proto定义与应用服务层期望之间的一个重大差异，需要后续调整Proto或在接口层进行更复杂的数据获取。
	skuID := productID // 假设SKUID与ProductID相同，或在此处根据ProductID查询SKU信息。

	item, err := s.app.Add(ctx, userID, productID, skuID, "Unknown Product", "Unknown SKU", 0, "") // 0 for Price, "" for ImageURL。
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
	// 备注：Proto请求中的 ProductId 字段在此处被映射为应用服务层 Remove 方法的 id 参数。
	// 而应用服务层的 Remove 方法期望的是收藏夹条目本身的 ID，而不是 ProductID 或 SkuID。
	// 这可能导致功能不符或错误。如果需要按 ProductID 或 SkuID 移除，应用服务层需要提供对应的方法。
	idToRemove, err := strconv.ParseUint(req.ProductId, 10, 64) // 假设这里的ProductId是收藏夹条目的ID。
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id for removal: %v", err))
	}

	err = s.app.Remove(ctx, userID, idToRemove)
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
func (s *Server) toProto(item *entity.Wishlist) *pb.WishlistItem {
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

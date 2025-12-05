package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/pointsmall/v1"           // 导入积分商城模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/pointsmall/application"   // 导入积分商城模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain/entity" // 导入积分商城模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 PointsmallService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedPointsmallServiceServer                            // 嵌入生成的UnimplementedPointsmallServiceServer，确保前向兼容性。
	app                                     *application.PointsService // 依赖Points应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Pointsmall gRPC 服务端实例。
func NewServer(app *application.PointsService) *Server {
	return &Server{app: app}
}

// CreateProduct 处理创建积分商品的gRPC请求。
// req: 包含商品名称、描述、图片URL、所需积分、库存、限购数量和状态的请求体。
// 返回created successfully的商品响应和可能发生的gRPC错误。
func (s *Server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	// 将protobuf请求转换为领域实体所需的 PointsProduct 实体。
	product := &entity.PointsProduct{
		Name:         req.Name,
		Description:  req.Description,
		ImageURL:     req.ImageUrl,
		Points:       req.Points,
		Stock:        req.Stock,
		LimitPerUser: req.LimitPerUser,
		Status:       entity.PointsProductStatus(req.Status), // 直接转换状态。
	}

	// 调用应用服务层创建商品。
	if err := s.app.CreateProduct(ctx, product); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create product: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.CreateProductResponse{
		Product: convertProductToProto(product),
	}, nil
}

// ListProducts 处理列出积分商品的gRPC请求。
// req: 包含分页参数和状态过滤的请求体。
// 返回积分商品列表响应和可能发生的gRPC错误。
func (s *Server) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	// 获取分页参数。
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 根据Proto的状态值构建过滤器。
	var statusPtr *int
	if req.Status != -1 { // -1通常表示不进行状态过滤。
		st := int(req.Status)
		statusPtr = &st
	}

	// 调用应用服务层获取商品列表。
	products, total, err := s.app.ListProducts(ctx, statusPtr, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list products: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbProducts := make([]*pb.PointsProduct, len(products))
	for i, p := range products {
		pbProducts[i] = convertProductToProto(p)
	}

	return &pb.ListProductsResponse{
		Products:   pbProducts,
		TotalCount: total, // 总记录数。
	}, nil
}

// ExchangeProduct 处理兑换商品的gRPC请求。
// req: 包含用户ID、商品ID、数量、收货地址、电话和收货人信息的请求体。
// 返回兑换订单响应和可能发生的gRPC错误。
func (s *Server) ExchangeProduct(ctx context.Context, req *pb.ExchangeProductRequest) (*pb.ExchangeProductResponse, error) {
	order, err := s.app.ExchangeProduct(ctx, req.UserId, req.ProductId, req.Quantity, req.Address, req.Phone, req.Receiver)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to exchange product: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.ExchangeProductResponse{
		Order: convertOrderToProto(order),
	}, nil
}

// ListOrders 处理列出积分订单的gRPC请求。
// req: 包含用户ID、状态过滤和分页参数的请求体。
// 返回积分订单列表响应和可能发生的gRPC错误。
func (s *Server) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	// 获取分页参数。
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 根据Proto的状态值构建过滤器。
	var statusPtr *int
	if req.Status != -1 { // -1通常表示不进行状态过滤。
		st := int(req.Status)
		statusPtr = &st
	}

	// 调用应用服务层获取积分订单列表。
	orders, total, err := s.app.ListOrders(ctx, req.UserId, statusPtr, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list orders: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbOrders := make([]*pb.PointsOrder, len(orders))
	for i, o := range orders {
		pbOrders[i] = convertOrderToProto(o)
	}

	return &pb.ListOrdersResponse{
		Orders:     pbOrders,
		TotalCount: total, // 总记录数。
	}, nil
}

// GetAccount 处理获取用户积分账户信息的gRPC请求。
// req: 包含用户ID的请求体。
// 返回积分账户响应和可能发生的gRPC错误。
func (s *Server) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	account, err := s.app.GetAccount(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get account: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.GetAccountResponse{
		Account: convertAccountToProto(account),
	}, nil
}

// AddPoints 处理增加用户积分的gRPC请求（通常由管理员或系统调用）。
// req: 包含用户ID、积分数量、描述和关联ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) AddPoints(ctx context.Context, req *pb.AddPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.AddPoints(ctx, req.UserId, req.Points, req.Description, req.RefId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add points: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// convertProductToProto 是一个辅助函数，将领域层的 PointsProduct 实体转换为 protobuf 的 PointsProduct 消息。
func convertProductToProto(p *entity.PointsProduct) *pb.PointsProduct {
	if p == nil {
		return nil
	}
	return &pb.PointsProduct{
		Id:           uint64(p.ID),                 // ID。
		Name:         p.Name,                       // 名称。
		Description:  p.Description,                // 描述。
		ImageUrl:     p.ImageURL,                   // 图片URL。
		Points:       p.Points,                     // 所需积分。
		Stock:        p.Stock,                      // 库存。
		SoldCount:    p.SoldCount,                  // 已售数量。
		LimitPerUser: p.LimitPerUser,               // 每人限购。
		Status:       int32(p.Status),              // 状态。
		CreatedAt:    timestamppb.New(p.CreatedAt), // 创建时间。
		UpdatedAt:    timestamppb.New(p.UpdatedAt), // 更新时间。
	}
}

// convertOrderToProto 是一个辅助函数，将领域层的 PointsOrder 实体转换为 protobuf 的 PointsOrder 消息。
func convertOrderToProto(o *entity.PointsOrder) *pb.PointsOrder {
	if o == nil {
		return nil
	}
	var shippedAt, completedAt *timestamppb.Timestamp
	if o.ShippedAt != nil {
		shippedAt = timestamppb.New(*o.ShippedAt)
	}
	if o.CompletedAt != nil {
		completedAt = timestamppb.New(*o.CompletedAt)
	}

	return &pb.PointsOrder{
		Id:          uint64(o.ID),                 // ID。
		OrderNo:     o.OrderNo,                    // 订单编号。
		UserId:      o.UserID,                     // 用户ID。
		ProductId:   o.ProductID,                  // 商品ID。
		ProductName: o.ProductName,                // 商品名称。
		Quantity:    o.Quantity,                   // 数量。
		Points:      o.Points,                     // 单价积分。
		TotalPoints: o.TotalPoints,                // 总积分。
		Status:      int32(o.Status),              // 状态。
		Address:     o.Address,                    // 收货地址。
		Phone:       o.Phone,                      // 联系电话。
		Receiver:    o.Receiver,                   // 收货人。
		ShippedAt:   shippedAt,                    // 发货时间。
		CompletedAt: completedAt,                  // 完成时间。
		CreatedAt:   timestamppb.New(o.CreatedAt), // 创建时间。
		UpdatedAt:   timestamppb.New(o.UpdatedAt), // 更新时间。
	}
}

// convertAccountToProto 是一个辅助函数，将领域层的 PointsAccount 实体转换为 protobuf 的 PointsAccount 消息。
func convertAccountToProto(a *entity.PointsAccount) *pb.PointsAccount {
	if a == nil {
		return nil
	}
	return &pb.PointsAccount{
		Id:          uint64(a.ID),                 // ID。
		UserId:      a.UserID,                     // 用户ID。
		TotalPoints: a.TotalPoints,                // 总积分。
		UsedPoints:  a.UsedPoints,                 // 已用积分。
		CreatedAt:   timestamppb.New(a.CreatedAt), // 创建时间。
		UpdatedAt:   timestamppb.New(a.UpdatedAt), // 更新时间。
	}
}

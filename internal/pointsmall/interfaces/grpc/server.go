package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/pointsmall/v1"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/application"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体定义。
type Server struct {
	pb.UnimplementedPointsmallServer
	app *application.PointsmallService
}

// NewServer 函数。
func NewServer(app *application.PointsmallService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	product := &domain.PointsProduct{
		Name:         req.Name,
		Description:  req.Description,
		ImageURL:     req.ImageUrl,
		Points:       req.Points,
		Stock:        req.Stock,
		LimitPerUser: req.LimitPerUser,
		Status:       domain.PointsProductStatus(req.Status),
	}

	if err := s.app.CreateProduct(ctx, product); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create product: %v", err))
	}

	return &pb.CreateProductResponse{
		Product: convertProductToProto(product),
	}, nil
}

func (s *Server) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	var statusPtr *int
	if req.Status != -1 {
		st := int(req.Status)
		statusPtr = &st
	}

	products, total, err := s.app.ListProducts(ctx, statusPtr, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list products: %v", err))
	}

	pbProducts := make([]*pb.PointsProduct, len(products))
	for i, p := range products {
		pbProducts[i] = convertProductToProto(p)
	}

	return &pb.ListProductsResponse{
		Products:   pbProducts,
		TotalCount: total,
	}, nil
}

func (s *Server) ExchangeProduct(ctx context.Context, req *pb.ExchangeProductRequest) (*pb.ExchangeProductResponse, error) {
	order, err := s.app.ExchangeProduct(ctx, req.UserId, req.ProductId, req.Quantity, req.Address, req.Phone, req.Receiver)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to exchange product: %v", err))
	}

	return &pb.ExchangeProductResponse{
		Order: convertOrderToProto(order),
	}, nil
}

func (s *Server) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	var statusPtr *int
	if req.Status != -1 {
		st := int(req.Status)
		statusPtr = &st
	}

	orders, total, err := s.app.ListOrders(ctx, req.UserId, statusPtr, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list orders: %v", err))
	}

	pbOrders := make([]*pb.PointsOrder, len(orders))
	for i, o := range orders {
		pbOrders[i] = convertOrderToProto(o)
	}

	return &pb.ListOrdersResponse{
		Orders:     pbOrders,
		TotalCount: total,
	}, nil
}

func (s *Server) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	account, err := s.app.GetAccount(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get account: %v", err))
	}

	return &pb.GetAccountResponse{
		Account: convertAccountToProto(account),
	}, nil
}

func (s *Server) AddPoints(ctx context.Context, req *pb.AddPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.AddPoints(ctx, req.UserId, req.Points, req.Description, req.RefId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add points: %v", err))
	}
	return &emptypb.Empty{}, nil
}

func convertProductToProto(p *domain.PointsProduct) *pb.PointsProduct {
	if p == nil {
		return nil
	}
	return &pb.PointsProduct{
		Id:           uint64(p.ID),
		Name:         p.Name,
		Description:  p.Description,
		ImageUrl:     p.ImageURL,
		Points:       p.Points,
		Stock:        p.Stock,
		SoldCount:    p.SoldCount,
		LimitPerUser: p.LimitPerUser,
		Status:       int32(p.Status),
		CreatedAt:    timestamppb.New(p.CreatedAt),
		UpdatedAt:    timestamppb.New(p.UpdatedAt),
	}
}

func convertOrderToProto(o *domain.PointsOrder) *pb.PointsOrder {
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
		Id:          uint64(o.ID),
		OrderNo:     o.OrderNo,
		UserId:      o.UserID,
		ProductId:   o.ProductID,
		ProductName: o.ProductName,
		Quantity:    o.Quantity,
		Points:      o.Points,
		TotalPoints: o.TotalPoints,
		Status:      int32(o.Status),
		Address:     o.Address,
		Phone:       o.Phone,
		Receiver:    o.Receiver,
		ShippedAt:   shippedAt,
		CompletedAt: completedAt,
		CreatedAt:   timestamppb.New(o.CreatedAt),
		UpdatedAt:   timestamppb.New(o.UpdatedAt),
	}
}

func convertAccountToProto(a *domain.PointsAccount) *pb.PointsAccount {
	if a == nil {
		return nil
	}
	return &pb.PointsAccount{
		Id:          uint64(a.ID),
		UserId:      a.UserID,
		TotalPoints: a.TotalPoints,
		UsedPoints:  a.UsedPoints,
		CreatedAt:   timestamppb.New(a.CreatedAt),
		UpdatedAt:   timestamppb.New(a.UpdatedAt),
	}
}

// 手动添加此函数，因为它不在查看的文件中，但可能需要或有用。
// 如果它没被使用，也不会有什么坏处，但保险起见还是加上。
// 实际上，timestamppb.New 处理的是 time.Time。
// 等等，timestamppb.New 接受 *time.Time 吗？不，它通常直接接受 time.Time。
// 查阅 protobuf 文档... New 接受 time.Time。
// 所以如果 CreatedAt 是 time.Time (gorm.Model)，那就没问题。
// 但如果 ShippedAt 是 *time.Time，我们需要在传递给 New 之前解引用，但要先检查 nil。
// 我已经在 convertOrderToProto 中处理了 nil 检查。

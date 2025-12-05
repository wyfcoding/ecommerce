package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。

	pb "github.com/wyfcoding/ecommerce/go-api/settlement/v1"           // 导入结算模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/settlement/application"   // 导入结算模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/settlement/domain/entity" // 导入结算模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 SettlementService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedSettlementServiceServer                                // 嵌入生成的UnimplementedSettlementServiceServer，确保前向兼容性。
	app                                     *application.SettlementService // 依赖Settlement应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Settlement gRPC 服务端实例。
func NewServer(app *application.SettlementService) *Server {
	return &Server{app: app}
}

// CreateSettlement 处理创建结算单的gRPC请求。
// req: 包含商户ID、结算周期、开始/结束日期等信息的请求体。
// 返回创建成功的结算单响应和可能发生的gRPC错误。
func (s *Server) CreateSettlement(ctx context.Context, req *pb.CreateSettlementRequest) (*pb.CreateSettlementResponse, error) {
	settlement, err := s.app.CreateSettlement(ctx, req.MerchantId, req.Cycle, req.StartDate.AsTime(), req.EndDate.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create settlement: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.CreateSettlementResponse{
		Settlement: convertSettlementToProto(settlement),
	}, nil
}

// AddOrderToSettlement 处理添加订单到结算单的gRPC请求。
// req: 包含结算单ID、订单ID、订单号和金额的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) AddOrderToSettlement(ctx context.Context, req *pb.AddOrderToSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.AddOrderToSettlement(ctx, req.SettlementId, req.OrderId, req.OrderNo, req.Amount); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add order to settlement: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ProcessSettlement 处理结算单的gRPC请求。
// req: 包含结算单ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) ProcessSettlement(ctx context.Context, req *pb.ProcessSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.ProcessSettlement(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to process settlement: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// CompleteSettlement 处理完成结算单的gRPC请求。
// req: 包含结算单ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) CompleteSettlement(ctx context.Context, req *pb.CompleteSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.CompleteSettlement(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to complete settlement: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListSettlements 处理列出结算单的gRPC请求。
// req: 包含商户ID、状态过滤和分页参数的请求体。
// 返回结算单列表响应和可能发生的gRPC错误。
func (s *Server) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
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

	// 调用应用服务层获取结算单列表。
	settlements, total, err := s.app.ListSettlements(ctx, req.MerchantId, statusPtr, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list settlements: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbSettlements := make([]*pb.Settlement, len(settlements))
	for i, s := range settlements {
		pbSettlements[i] = convertSettlementToProto(s)
	}

	return &pb.ListSettlementsResponse{
		Settlements: pbSettlements,
		TotalCount:  total, // 总记录数。
	}, nil
}

// GetMerchantAccount 处理获取商户账户信息的gRPC请求。
// req: 包含商户ID的请求体。
// 返回商户账户响应和可能发生的gRPC错误。
func (s *Server) GetMerchantAccount(ctx context.Context, req *pb.GetMerchantAccountRequest) (*pb.GetMerchantAccountResponse, error) {
	account, err := s.app.GetMerchantAccount(ctx, req.MerchantId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get merchant account: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.GetMerchantAccountResponse{
		Account: convertAccountToProto(account),
	}, nil
}

// convertSettlementToProto 是一个辅助函数，将领域层的 Settlement 实体转换为 protobuf 的 Settlement 消息。
func convertSettlementToProto(s *entity.Settlement) *pb.Settlement {
	if s == nil {
		return nil
	}
	// 转换可选的结算时间字段。
	var settledAt *timestamppb.Timestamp
	if s.SettledAt != nil {
		settledAt = timestamppb.New(*s.SettledAt)
	}

	return &pb.Settlement{
		Id:               uint64(s.ID),                 // ID。
		SettlementNo:     s.SettlementNo,               // 结算单号。
		MerchantId:       s.MerchantID,                 // 商户ID。
		Cycle:            string(s.Cycle),              // 结算周期。
		StartDate:        timestamppb.New(s.StartDate), // 开始日期。
		EndDate:          timestamppb.New(s.EndDate),   // 结束日期。
		OrderCount:       s.OrderCount,                 // 订单数量。
		TotalAmount:      s.TotalAmount,                // 总金额。
		PlatformFee:      s.PlatformFee,                // 平台手续费。
		SettlementAmount: s.SettlementAmount,           // 结算金额。
		Status:           int32(s.Status),              // 状态。
		SettledAt:        settledAt,                    // 结算时间。
		FailReason:       s.FailReason,                 // 失败原因。
		CreatedAt:        timestamppb.New(s.CreatedAt), // 创建时间。
		UpdatedAt:        timestamppb.New(s.UpdatedAt), // 更新时间。
	}
}

// convertAccountToProto 是一个辅助函数，将领域层的 MerchantAccount 实体转换为 protobuf 的 MerchantAccount 消息。
func convertAccountToProto(a *entity.MerchantAccount) *pb.MerchantAccount {
	if a == nil {
		return nil
	}
	return &pb.MerchantAccount{
		Id:            uint64(a.ID),                 // ID。
		MerchantId:    a.MerchantID,                 // 商户ID。
		Balance:       a.Balance,                    // 余额。
		FrozenBalance: a.FrozenBalance,              // 冻结金额。
		TotalIncome:   a.TotalIncome,                // 总收入。
		TotalWithdraw: a.TotalWithdraw,              // 总提现。
		FeeRate:       a.FeeRate,                    // 费率。
		CreatedAt:     timestamppb.New(a.CreatedAt), // 创建时间。
		UpdatedAt:     timestamppb.New(a.UpdatedAt), // 更新时间。
	}
}

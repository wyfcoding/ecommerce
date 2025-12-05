package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/financial_settlement/v1"           // 导入财务结算模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/application"   // 导入财务结算模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain/entity" // 导入财务结算模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 FinancialSettlementService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedFinancialSettlementServiceServer                                         // 嵌入生成的UnimplementedFinancialSettlementServiceServer，确保前向兼容性。
	app                                              *application.FinancialSettlementService // 依赖FinancialSettlement应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 FinancialSettlement gRPC 服务端实例。
func NewServer(app *application.FinancialSettlementService) *Server {
	return &Server{app: app}
}

// CreateSettlement 处理创建结算单的gRPC请求。
// req: 包含卖家ID、结算周期、开始日期和结束日期的请求体。
// 返回创建成功的结算单响应和可能发生的gRPC错误。
func (s *Server) CreateSettlement(ctx context.Context, req *pb.CreateSettlementRequest) (*pb.CreateSettlementResponse, error) {
	// 调用应用服务层创建结算单。
	settlement, err := s.app.CreateSettlement(ctx, req.SellerId, req.Period, req.StartDate.AsTime(), req.EndDate.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create settlement: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.CreateSettlementResponse{
		Settlement: convertSettlementToProto(settlement),
	}, nil
}

// ApproveSettlement 处理批准结算单的gRPC请求。
// req: 包含结算单ID和批准人信息的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) ApproveSettlement(ctx context.Context, req *pb.ApproveSettlementRequest) (*emptypb.Empty, error) {
	// 调用应用服务层批准结算单。
	if err := s.app.ApproveSettlement(ctx, req.Id, req.ApprovedBy); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to approve settlement: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// RejectSettlement 处理拒绝结算单的gRPC请求。
// req: 包含结算单ID和拒绝原因的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) RejectSettlement(ctx context.Context, req *pb.RejectSettlementRequest) (*emptypb.Empty, error) {
	// 调用应用服务层拒绝结算单。
	if err := s.app.RejectSettlement(ctx, req.Id, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to reject settlement: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// GetSettlement 处理获取结算单详情的gRPC请求。
// req: 包含结算单ID的请求体。
// 返回结算单响应和可能发生的gRPC错误。
func (s *Server) GetSettlement(ctx context.Context, req *pb.GetSettlementRequest) (*pb.GetSettlementResponse, error) {
	settlement, err := s.app.GetSettlement(ctx, req.Id)
	if err != nil {
		// 如果结算单未找到，返回NotFound状态码。
		return nil, status.Error(codes.NotFound, fmt.Sprintf("settlement not found: %v", err))
	}
	return &pb.GetSettlementResponse{
		Settlement: convertSettlementToProto(settlement),
	}, nil
}

// ListSettlements 处理列出结算单的gRPC请求。
// req: 包含卖家ID和分页参数的请求体。
// 返回结算单列表响应和可能发生的gRPC错误。
func (s *Server) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取结算单列表。
	settlements, total, err := s.app.ListSettlements(ctx, req.SellerId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list settlements: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbSettlements := make([]*pb.Settlement, len(settlements))
	for i, st := range settlements {
		pbSettlements[i] = convertSettlementToProto(st)
	}

	return &pb.ListSettlementsResponse{
		Settlements: pbSettlements,
		TotalCount:  uint64(total), // 总记录数。
	}, nil
}

// ProcessPayment 处理结算单支付的gRPC请求。
// req: 包含结算单ID的请求体。
// 返回支付记录响应和可能发生的gRPC错误。
func (s *Server) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	payment, err := s.app.ProcessPayment(ctx, req.SettlementId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to process payment: %v", err))
	}

	return &pb.ProcessPaymentResponse{
		Payment: convertPaymentToProto(payment),
	}, nil
}

// convertSettlementToProto 是一个辅助函数，将领域层的 Settlement 实体转换为 protobuf 的 Settlement 消息。
func convertSettlementToProto(s *entity.Settlement) *pb.Settlement {
	if s == nil {
		return nil
	}
	resp := &pb.Settlement{
		Id:               uint64(s.ID),                 // 结算单ID。
		SellerId:         s.SellerID,                   // 卖家ID。
		Period:           s.Period,                     // 结算周期。
		StartDate:        timestamppb.New(s.StartDate), // 开始日期。
		EndDate:          timestamppb.New(s.EndDate),   // 结束日期。
		TotalSalesAmount: s.TotalSalesAmount,           // 总销售额。
		CommissionAmount: s.CommissionAmount,           // 佣金金额。
		RebateAmount:     s.RebateAmount,               // 返利金额。
		OtherFees:        s.OtherFees,                  // 其他费用。
		FinalAmount:      s.FinalAmount,                // 最终结算金额。
		Status:           string(s.Status),             // 状态。
		ApprovedBy:       s.ApprovedBy,                 // 审核人。
		RejectionReason:  s.RejectionReason,            // 拒绝原因。
		CreatedAt:        timestamppb.New(s.CreatedAt), // 创建时间。
	}
	if s.ApprovedAt != nil {
		resp.ApprovedAt = timestamppb.New(*s.ApprovedAt) // 审核时间。
	}
	// Proto中还包含 UpdatedAt, DeletedAt 等字段，但实体中没有或未映射。
	return resp
}

// convertPaymentToProto 是一个辅助函数，将领域层的 SettlementPayment 实体转换为 protobuf 的 SettlementPayment 消息。
func convertPaymentToProto(p *entity.SettlementPayment) *pb.SettlementPayment {
	if p == nil {
		return nil
	}
	resp := &pb.SettlementPayment{
		Id:            uint64(p.ID),                 // 支付记录ID。
		SettlementId:  p.SettlementID,               // 结算单ID。
		SellerId:      p.SellerID,                   // 卖家ID。
		Amount:        p.Amount,                     // 支付金额。
		Status:        string(p.Status),             // 支付状态。
		TransactionId: p.TransactionID,              // 交易流水号。
		CreatedAt:     timestamppb.New(p.CreatedAt), // 创建时间。
	}
	if p.CompletedAt != nil {
		resp.CompletedAt = timestamppb.New(*p.CompletedAt) // 完成时间。
	}
	// Proto中还包含 UpdatedAt, DeletedAt 等字段，但实体中没有或未映射。
	return resp
}

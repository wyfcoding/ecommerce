package grpc

import (
	"context" // 导入标准错误处理包。
	"fmt"     // 导入格式化包。

	pb "github.com/wyfcoding/ecommerce/goapi/aftersales/v1"                // 导入售后模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/aftersales/application"       // 导入售后模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/entity"     // 导入售后模块的领域实体。
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/repository" // 导入售后模块的仓储层查询对象。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 AftersalesService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedAftersalesServiceServer                                // 嵌入生成的UnimplementedAftersalesServiceServer，确保前向兼容性。
	app                                     *application.AfterSalesService // 依赖AfterSales应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 AfterSales gRPC 服务端实例。
func NewServer(app *application.AfterSalesService) *Server {
	return &Server{app: app}
}

// CreateReturnRequest 处理创建退货（售后）申请的gRPC请求。
// req: 包含创建退货申请所需信息的请求体。
// 返回created successfully的退货申请响应和可能发生的gRPC错误。
func (s *Server) CreateReturnRequest(ctx context.Context, req *pb.CreateReturnRequestRequest) (*pb.ReturnRequestResponse, error) {
	// 将protobuf定义的售后请求类型映射到领域实体定义的售后类型。
	var entityType entity.AfterSalesType
	switch req.RequestType {
	case pb.ReturnRequestType_RETURN_REQUEST_TYPE_RETURN:
		entityType = entity.AfterSalesTypeReturnGoods
	case pb.ReturnRequestType_RETURN_REQUEST_TYPE_REFUND:
		entityType = entity.AfterSalesTypeRefund
	case pb.ReturnRequestType_RETURN_REQUEST_TYPE_EXCHANGE:
		entityType = entity.AfterSalesTypeExchange
	default:
		entityType = entity.AfterSalesTypeReturnGoods // 默认处理。
	}

	// 注意：Proto请求中缺少详细的订单商品信息 (product_id, sku_id等) 和 orderNo。
	// 当前实现传递空商品列表和 "UNKNOWN" 作为订单号。
	// 在实际系统中，这部分信息可能需要从其他服务获取或通过更完善的Proto定义传递。
	items := []*entity.AfterSalesItem{}

	// 调用应用服务层创建售后申请。
	as, err := s.app.CreateAfterSales(ctx, req.OrderId, "UNKNOWN", req.UserId, entityType, req.Reason, req.GetDescription(), req.ImageUrls, items)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create return request: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.ReturnRequestResponse{
		Request: s.toProto(as),
	}, nil
}

// GetReturnRequest 处理获取单个退货（售后）申请的gRPC请求。
// req: 包含售后申请ID的请求体。
// 返回售后申请响应和可能发生的gRPC错误。
func (s *Server) GetReturnRequest(ctx context.Context, req *pb.GetReturnRequestRequest) (*pb.ReturnRequestResponse, error) {
	as, err := s.app.GetDetails(ctx, req.Id)
	if err != nil {
		// 如果售后记录未找到，返回NotFound状态码。
		return nil, status.Error(codes.NotFound, fmt.Sprintf("return request not found: %v", err))
	}
	return &pb.ReturnRequestResponse{
		Request: s.toProto(as),
	}, nil
}

// UpdateReturnRequestStatus 处理更新退货（售后）申请状态的gRPC请求。
// req: 包含要更新的状态和其他信息的请求体。
// 返回更新后的售后申请响应和可能发生的gRPC错误。
func (s *Server) UpdateReturnRequestStatus(ctx context.Context, req *pb.UpdateReturnRequestStatusRequest) (*pb.ReturnRequestResponse, error) {
	// 根据请求的状态调用应用服务层的 Approve 或 Reject 方法。
	switch req.Status {
	case pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_APPROVED:
		// 如果是批准操作，需要获取退款金额（Proto中以元为单位，转换为分）。
		amount := int64(req.GetRefundAmount() * 100)
		if err := s.app.Approve(ctx, req.Id, "admin", amount); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to approve return request: %v", err))
		}
	case pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_REJECTED:
		// 如果是拒绝操作，需要获取拒绝原因。
		reason := req.GetAdminNote()
		if reason == "" {
			reason = "Rejected by admin" // 提供默认拒绝原因。
		}
		if err := s.app.Reject(ctx, req.Id, "admin", reason); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to reject return request: %v", err))
		}
	default:
		// 暂时只支持批准和拒绝操作。
		return nil, status.Error(codes.Unimplemented, "Only Approve and Reject are supported via this API for now")
	}

	// 获取更新后的售后申请详情，以便在响应中返回最新状态。
	as, err := s.app.GetDetails(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get updated return request details: %v", err))
	}

	return &pb.ReturnRequestResponse{
		Request: s.toProto(as),
	}, nil
}

// ListReturnRequests 处理列出退货（售后）申请列表的gRPC请求，支持分页和过滤。
// req: 包含分页和过滤参数的请求体。
// 返回售后申请列表响应和可能发生的gRPC错误。
func (s *Server) ListReturnRequests(ctx context.Context, req *pb.ListReturnRequestsRequest) (*pb.ListReturnRequestsResponse, error) {
	// 将protobuf请求中的分页和过滤参数映射到应用服务层使用的查询对象。
	query := &repository.AfterSalesQuery{
		Page:     int(req.PageToken),
		PageSize: int(req.PageSize),
	}
	// 处理可选的用户ID过滤。
	if req.UserId != nil {
		query.UserID = *req.UserId
	}
	// 处理可选的订单ID过滤。
	if req.OrderId != nil {
		query.OrderID = *req.OrderId
	}

	// 确保分页参数有效。
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 10
	}

	// 映射protobuf的售后状态到领域实体定义的售后状态，用于过滤。
	if req.Status != nil {
		// 注意：此处强制类型转换，需要确保protobuf和entity的状态枚举值是一致的。
		st := entity.AfterSalesStatus(*req.Status)
		query.Status = st
	}

	// 调用应用服务层获取售后申请列表。
	list, total, err := s.app.List(ctx, query)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list return requests: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbList := make([]*pb.ReturnRequest, len(list))
	for i, as := range list {
		pbList[i] = s.toProto(as)
	}

	return &pb.ListReturnRequestsResponse{
		Requests:   pbList,
		TotalCount: int32(total), // 总记录数。
	}, nil
}

// ProcessRefund 处理退款流程的gRPC请求。
func (s *Server) ProcessRefund(ctx context.Context, req *pb.ProcessRefundRequest) (*pb.RefundResponse, error) {
	if err := s.app.ProcessRefund(ctx, req.ReturnRequestId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to process refund: %v", err))
	}
	return &pb.RefundResponse{
		ReturnRequestId: req.ReturnRequestId,
		Status:          "SUCCESS",
		// 金额、币种等字段可以从应用服务响应中填充（如果可用）
	}, nil
}

// ProcessExchange 处理换货流程的gRPC请求。
func (s *Server) ProcessExchange(ctx context.Context, req *pb.ProcessExchangeRequest) (*pb.ExchangeResponse, error) {
	if err := s.app.ProcessExchange(ctx, req.ReturnRequestId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to process exchange: %v", err))
	}
	return &pb.ExchangeResponse{
		ReturnRequestId: req.ReturnRequestId,
		Status:          "SUCCESS",
	}, nil
}

// --- 客服工单方法 ---

// CreateSupportTicket 创建客服工单。
func (s *Server) CreateSupportTicket(ctx context.Context, req *pb.CreateSupportTicketRequest) (*pb.SupportTicketResponse, error) {
	// 将 proto 枚举映射为 int8
	var priority int8 = 2 // 默认中等
	if req.Priority != nil {
		switch *req.Priority {
		case pb.SupportTicketPriority_SUPPORT_TICKET_PRIORITY_LOW:
			priority = 1
		case pb.SupportTicketPriority_SUPPORT_TICKET_PRIORITY_MEDIUM:
			priority = 2
		case pb.SupportTicketPriority_SUPPORT_TICKET_PRIORITY_HIGH:
			priority = 3
		case pb.SupportTicketPriority_SUPPORT_TICKET_PRIORITY_URGENT:
			priority = 4
		}
	}

	orderID := uint64(0)
	if req.OrderId != nil {
		orderID = *req.OrderId
	}

	// 应用服务需要 Description，将 InitialMessage 映射给它。
	// proto 中缺少 Category，传递空值。
	ticket, err := s.app.CreateSupportTicket(ctx, req.UserId, orderID, req.Subject, req.InitialMessage, "", priority)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create support ticket: %v", err))
	}
	return &pb.SupportTicketResponse{
		Ticket: s.toSupportTicketProto(ticket),
	}, nil
}

// GetSupportTicket 获取客服工单详情。
func (s *Server) GetSupportTicket(ctx context.Context, req *pb.GetSupportTicketRequest) (*pb.SupportTicketResponse, error) {
	ticket, err := s.app.GetSupportTicket(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get support ticket: %v", err))
	}
	if ticket == nil {
		return nil, status.Error(codes.NotFound, "support ticket not found")
	}
	return &pb.SupportTicketResponse{
		Ticket: s.toSupportTicketProto(ticket),
	}, nil
}

// UpdateSupportTicketStatus 更新客服工单状态。
func (s *Server) UpdateSupportTicketStatus(ctx context.Context, req *pb.UpdateSupportTicketStatusRequest) (*pb.SupportTicketResponse, error) {
	var st entity.SupportTicketStatus
	switch req.Status {
	case pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_OPEN:
		st = entity.SupportTicketStatusOpen
	case pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_IN_PROGRESS:
		st = entity.SupportTicketStatusPending // 将 InProgress 映射为 Pending 或类似状态
	case pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_RESOLVED:
		st = entity.SupportTicketStatusResolved
	case pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_CLOSED:
		st = entity.SupportTicketStatusClosed
	default:
		st = entity.SupportTicketStatusOpen
	}

	if err := s.app.UpdateSupportTicketStatus(ctx, req.Id, st); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update support ticket status: %v", err))
	}
	// 返回更新后的工单
	ticket, err := s.app.GetSupportTicket(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get updated ticket: %v", err))
	}
	return &pb.SupportTicketResponse{
		Ticket: s.toSupportTicketProto(ticket),
	}, nil
}

// AddSupportTicketMessage 为客服工单添加消息。
func (s *Server) AddSupportTicketMessage(ctx context.Context, req *pb.AddSupportTicketMessageRequest) (*pb.SupportTicketMessageResponse, error) {
	senderType := "User"
	if req.IsAdminSender {
		senderType = "Agent"
	}
	msg, err := s.app.CreateSupportTicketMessage(ctx, req.TicketId, req.SenderId, senderType, req.Content)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add message: %v", err))
	}
	return &pb.SupportTicketMessageResponse{
		Message: s.toSupportTicketMessageProto(msg),
	}, nil
}

// ListSupportTickets 列出客服工单。
func (s *Server) ListSupportTickets(ctx context.Context, req *pb.ListSupportTicketsRequest) (*pb.ListSupportTicketsResponse, error) {
	page := max(int(req.PageToken), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}
	// 状态过滤逻辑
	var statusPtr *int
	if req.Status != nil && *req.Status != pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_UNSPECIFIED {
		// 将 proto 状态映射为实体状态 int
		var sVal int
		switch *req.Status {
		case pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_OPEN:
			sVal = int(entity.SupportTicketStatusOpen)
		case pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_IN_PROGRESS:
			sVal = int(entity.SupportTicketStatusPending)
		case pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_RESOLVED:
			sVal = int(entity.SupportTicketStatusResolved)
		case pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_CLOSED:
			sVal = int(entity.SupportTicketStatusClosed)
		default:
			sVal = int(entity.SupportTicketStatusOpen)
		}
		statusPtr = &sVal
	}

	userID := uint64(0)
	if req.UserId != nil {
		userID = *req.UserId
	}

	list, total, err := s.app.ListSupportTickets(ctx, userID, statusPtr, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list tickets: %v", err))
	}

	pbList := make([]*pb.SupportTicket, len(list))
	for i, t := range list {
		pbList[i] = s.toSupportTicketProto(t)
	}

	return &pb.ListSupportTicketsResponse{
		Tickets:    pbList,
		TotalCount: int32(total),
	}, nil
}

// ListSupportTicketMessages 列出客服工单消息。
func (s *Server) ListSupportTicketMessages(ctx context.Context, req *pb.ListSupportTicketMessagesRequest) (*pb.ListSupportTicketMessagesResponse, error) {
	list, err := s.app.ListSupportTicketMessages(ctx, req.TicketId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list messages: %v", err))
	}

	pbList := make([]*pb.SupportTicketMessage, len(list))
	for i, m := range list {
		pbList[i] = s.toSupportTicketMessageProto(m)
	}

	return &pb.ListSupportTicketMessagesResponse{
		Messages: pbList,
	}, nil
}

// GetAftersalesConfig 获取售后配置。
func (s *Server) GetAftersalesConfig(ctx context.Context, req *pb.GetAftersalesConfigRequest) (*pb.AftersalesConfigResponse, error) {
	config, err := s.app.GetConfig(ctx, req.Key)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get config: %v", err))
	}
	if config == nil {
		return nil, status.Error(codes.NotFound, "config not found")
	}
	return &pb.AftersalesConfigResponse{
		Config: &pb.AftersalesConfig{
			Key:         config.Key,
			Value:       config.Value,
			Description: config.Description,
		},
	}, nil
}

// SetAftersalesConfig 设置售后配置。
func (s *Server) SetAftersalesConfig(ctx context.Context, req *pb.SetAftersalesConfigRequest) (*pb.AftersalesConfigResponse, error) {
	config, err := s.app.SetConfig(ctx, req.Key, req.Value, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to set config: %v", err))
	}
	return &pb.AftersalesConfigResponse{
		Config: &pb.AftersalesConfig{
			Key:         config.Key,
			Value:       config.Value,
			Description: config.Description,
		},
	}, nil
}

// --- 模块分段 ---

// toProto 是一个辅助函数，将领域层的 AfterSales 实体转换为 protobuf 的 ReturnRequest 消息。
func (s *Server) toProto(as *entity.AfterSales) *pb.ReturnRequest {
	// 映射领域实体状态到protobuf状态。
	var status pb.ReturnRequestStatus
	switch as.Status {
	case entity.AfterSalesStatusPending:
		status = pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_PENDING
	case entity.AfterSalesStatusApproved:
		status = pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_APPROVED
	case entity.AfterSalesStatusRejected:
		status = pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_REJECTED
	case entity.AfterSalesStatusCompleted:
		// 注意：Proto中没有Completed，这里映射到REFUNDED，可能需要调整。
		status = pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_REFUNDED
	case entity.AfterSalesStatusCancelled:
		status = pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_CLOSED
	default:
		status = pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_UNSPECIFIED
	}

	// 映射领域实体类型到protobuf类型。
	var rType pb.ReturnRequestType
	switch as.Type {
	case entity.AfterSalesTypeReturnGoods:
		rType = pb.ReturnRequestType_RETURN_REQUEST_TYPE_RETURN
	case entity.AfterSalesTypeRefund:
		rType = pb.ReturnRequestType_RETURN_REQUEST_TYPE_REFUND
	case entity.AfterSalesTypeExchange:
		rType = pb.ReturnRequestType_RETURN_REQUEST_TYPE_EXCHANGE
	default:
		rType = pb.ReturnRequestType_RETURN_REQUEST_TYPE_UNSPECIFIED
	}

	return &pb.ReturnRequest{
		Id:           uint64(as.ID),                    // 售后申请ID。
		UserId:       as.UserID,                        // 用户ID。
		OrderId:      as.OrderID,                       // 订单ID。
		RequestType:  rType,                            // 售后请求类型。
		Status:       status,                           // 售后请求状态。
		Reason:       as.Reason,                        // 申请原因。
		Description:  as.Description,                   // 详细描述。
		ImageUrls:    as.Images,                        // 凭证图片URL列表。
		RefundAmount: float64(as.RefundAmount) / 100.0, // 退款金额（分转元）。
		CreatedAt:    timestamppb.New(as.CreatedAt),    // 创建时间。
		UpdatedAt:    timestamppb.New(as.UpdatedAt),    // 更新时间。
		// Proto中还包含一些其他字段如 AdminNote, TrackingNumber 等，但实体中没有对应，此处未映射。
	}
}

func (s *Server) toSupportTicketProto(t *entity.SupportTicket) *pb.SupportTicket {
	if t == nil {
		return nil
	}
	var status pb.SupportTicketStatus
	switch t.Status {
	case entity.SupportTicketStatusOpen:
		status = pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_OPEN
	case entity.SupportTicketStatusPending:
		status = pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_IN_PROGRESS // 将 Pending 映射为 InProgress
	case entity.SupportTicketStatusResolved:
		status = pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_RESOLVED
	case entity.SupportTicketStatusClosed:
		status = pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_CLOSED
	default:
		status = pb.SupportTicketStatus_SUPPORT_TICKET_STATUS_UNSPECIFIED
	}

	var priority pb.SupportTicketPriority
	switch t.Priority {
	case 1:
		priority = pb.SupportTicketPriority_SUPPORT_TICKET_PRIORITY_LOW
	case 2:
		priority = pb.SupportTicketPriority_SUPPORT_TICKET_PRIORITY_MEDIUM
	case 3:
		priority = pb.SupportTicketPriority_SUPPORT_TICKET_PRIORITY_HIGH
	case 4:
		priority = pb.SupportTicketPriority_SUPPORT_TICKET_PRIORITY_URGENT
	default:
		priority = pb.SupportTicketPriority_SUPPORT_TICKET_PRIORITY_MEDIUM
	}

	orderID := uint64(0)
	if t.OrderID > 0 {
		orderID = t.OrderID
	}

	// 排除 TicketNo, Description, Category，因为它们不在 proto 中
	return &pb.SupportTicket{
		Id:        uint64(t.ID),
		UserId:    t.UserID,
		OrderId:   &orderID,
		Subject:   t.Subject,
		Status:    status,
		Priority:  priority,
		CreatedAt: timestamppb.New(t.CreatedAt),
		UpdatedAt: timestamppb.New(t.UpdatedAt),
	}
}

func (s *Server) toSupportTicketMessageProto(m *entity.SupportTicketMessage) *pb.SupportTicketMessage {
	if m == nil {
		return nil
	}
	isAdmin := false
	if m.SenderType != "User" {
		isAdmin = true
	}
	return &pb.SupportTicketMessage{
		Id:            uint64(m.ID),
		TicketId:      m.TicketID,
		SenderId:      m.SenderID,
		IsAdminSender: isAdmin,
		Content:       m.Content,
		CreatedAt:     timestamppb.New(m.CreatedAt),
	}
}

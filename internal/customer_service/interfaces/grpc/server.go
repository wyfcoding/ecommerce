package grpc

import (
	"context"
	"fmt"     // 导入格式化包，用于错误信息。
	"strings" // 导入字符串操作包。

	pb "github.com/wyfcoding/ecommerce/go-api/customer_service/v1"           // 导入客户服务模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/customer_service/application"   // 导入客户服务模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/customer_service/domain/entity" // 导入客户服务模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 CustomerService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedCustomerServiceServer                              // 嵌入生成的UnimplementedCustomerServiceServer，确保前向兼容性。
	app                                   *application.CustomerService // 依赖CustomerService应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 CustomerService gRPC 服务端实例。
func NewServer(app *application.CustomerService) *Server {
	return &Server{app: app}
}

// CreateTicket 处理创建工单的gRPC请求。
// req: 包含用户ID、主题、描述的请求体。
// 返回创建成功的工单响应和可能发生的gRPC错误。
func (s *Server) CreateTicket(ctx context.Context, req *pb.CreateTicketRequest) (*pb.TicketResponse, error) {
	// 应用服务层的 CreateTicket 方法需要 category 和 priority 字段。
	// Proto请求中缺少这些字段。这里使用默认值 "general" 作为类别，TicketPriorityMedium 作为优先级。
	ticket, err := s.app.CreateTicket(ctx, req.UserId, req.Subject, req.Description, "general", entity.TicketPriorityMedium)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create ticket: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.TicketResponse{
		Ticket: convertTicketToProto(ticket),
	}, nil
}

// GetTicketByID 处理根据ID获取工单信息的gRPC请求。
// req: 包含工单ID的请求体。
// 返回工单响应和可能发生的gRPC错误。
func (s *Server) GetTicketByID(ctx context.Context, req *pb.GetTicketByIDRequest) (*pb.TicketResponse, error) {
	ticket, err := s.app.GetTicket(ctx, req.TicketId)
	if err != nil {
		// 如果工单未找到，返回NotFound状态码。
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ticket not found: %v", err))
	}
	return &pb.TicketResponse{
		Ticket: convertTicketToProto(ticket),
	}, nil
}

// UpdateTicketStatus 处理更新工单状态的gRPC请求。
// req: 包含工单ID和新状态的请求体。
// 返回更新后的工单响应和可能发生的gRPC错误。
func (s *Server) UpdateTicketStatus(ctx context.Context, req *pb.UpdateTicketStatusRequest) (*pb.TicketResponse, error) {
	// 应用服务层目前仅显式暴露 CloseTicket 和 ResolveTicket 方法。
	// 这里根据Proto请求的状态字符串映射到这些方法。
	st := strings.ToUpper(req.Status)
	var err error

	switch st {
	case "CLOSED":
		err = s.app.CloseTicket(ctx, req.TicketId)
	case "RESOLVED":
		err = s.app.ResolveTicket(ctx, req.TicketId)
	default:
		// 对于Proto中可能存在的其他状态（例如 "IN_PROGRESS"），
		// 应用服务层可能没有直接的过渡方法，需要额外实现。
		return nil, status.Errorf(codes.Unimplemented, "status transition to %s not supported via gRPC yet", req.Status)
	}

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update ticket status: %v", err))
	}

	// 获取更新后的工单详情，以便在响应中返回最新状态。
	ticket, err := s.app.GetTicket(ctx, req.TicketId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to fetch updated ticket: %v", err))
	}

	return &pb.TicketResponse{
		Ticket: convertTicketToProto(ticket),
	}, nil
}

// AddMessageToTicket 处理向工单添加消息的gRPC请求。
// req: 包含工单ID、发送者信息和消息内容的请求体。
// 返回添加成功的消息响应和可能发生的gRPC错误。
func (s *Server) AddMessageToTicket(ctx context.Context, req *pb.AddMessageToTicketRequest) (*pb.TicketMessageResponse, error) {
	// 应用服务层的 ReplyTicket 方法需要 msgType 字段。
	// Proto请求中缺少 msgType。这里默认使用 MessageTypeText。
	msg, err := s.app.ReplyTicket(ctx, req.TicketId, req.SenderId, req.SenderType, req.MessageBody, entity.MessageTypeText)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add message to ticket: %v", err))
	}

	return &pb.TicketMessageResponse{
		Message: convertMessageToProto(msg),
	}, nil
}

// GetTicketMessages 处理获取工单消息列表的gRPC请求。
// req: 包含工单ID的请求体。
// 返回工单消息列表响应和可能发生的gRPC错误。
func (s *Server) GetTicketMessages(ctx context.Context, req *pb.GetTicketMessagesRequest) (*pb.GetTicketMessagesResponse, error) {
	// Proto请求中缺少分页字段 (page, pageSize)。
	// 这里使用默认值1页100条消息。
	msgs, _, err := s.app.ListMessages(ctx, req.TicketId, 1, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list ticket messages: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbMsgs := make([]*pb.TicketMessage, len(msgs))
	for i, m := range msgs {
		pbMsgs[i] = convertMessageToProto(m)
	}

	return &pb.GetTicketMessagesResponse{
		Messages: pbMsgs,
	}, nil
}

// convertTicketToProto 是一个辅助函数，将领域层的 Ticket 实体转换为 protobuf 的 TicketInfo 消息。
func convertTicketToProto(t *entity.Ticket) *pb.TicketInfo {
	if t == nil {
		return nil
	}

	// 映射领域实体状态到protobuf状态字符串。
	statusStr := "UNKNOWN"
	switch t.Status {
	case entity.TicketStatusOpen:
		statusStr = "OPEN"
	case entity.TicketStatusInProgress:
		statusStr = "IN_PROGRESS"
	case entity.TicketStatusResolved:
		statusStr = "RESOLVED"
	case entity.TicketStatusClosed:
		statusStr = "CLOSED"
	}

	return &pb.TicketInfo{
		TicketId:    uint64(t.ID),                 // 工单ID。
		UserId:      t.UserID,                     // 用户ID。
		Subject:     t.Subject,                    // 主题。
		Description: t.Description,                // 描述。
		Status:      statusStr,                    // 状态。
		CreatedAt:   timestamppb.New(t.CreatedAt), // 创建时间。
		UpdatedAt:   timestamppb.New(t.UpdatedAt), // 更新时间。
		// Proto中还包含一些其他字段如 Category, Priority, AssigneeId, ResolvedAt, ClosedAt 等，但实体中没有或未映射。
	}
}

// convertMessageToProto 是一个辅助函数，将领域层的 Message 实体转换为 protobuf 的 TicketMessage 消息。
func convertMessageToProto(m *entity.Message) *pb.TicketMessage {
	if m == nil {
		return nil
	}
	return &pb.TicketMessage{
		MessageId:   uint64(m.ID),                 // 消息ID。
		TicketId:    m.TicketID,                   // 工单ID。
		SenderId:    m.SenderID,                   // 发送者ID。
		SenderType:  m.SenderType,                 // 发送者类型。
		MessageBody: m.Content,                    // 消息内容。
		SentAt:      timestamppb.New(m.CreatedAt), // 发送时间。
		// Proto中还包含 Type, IsInternal 等字段，但实体中没有或未映射。
	}
}

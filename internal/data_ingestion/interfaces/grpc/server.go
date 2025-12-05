package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/data_ingestion/v1" // 导入数据摄取模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/application"
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb" // 导入空消息类型。
)

// Server 结构体实现了 DataIngestionService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
// Server 结构体实现了 DataIngestionService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedDataIngestionServer                                   // 嵌入生成的UnimplementedDataIngestionServer，确保前向兼容性。
	app                                 *application.DataIngestionService // 依赖DataIngestion应用服务
}

// NewServer 创建并返回一个新的 DataIngestion gRPC 服务端实例。
// app: 依赖DataIngestion应用服务
func NewServer(app *application.DataIngestionService) *Server {
	return &Server{
		app: app,
	}
}

// IngestEvent 处理摄取单个事件的gRPC请求。
// req: 包含事件类型、事件数据、时间戳和来源的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) IngestEvent(ctx context.Context, req *pb.IngestEventRequest) (*emptypb.Empty, error) {
	// 将map转换为JSON字符串存储
	// 注意：这里简单处理，实际可能需要更复杂的序列化
	eventData := fmt.Sprintf("%v", req.EventData)

	if err := s.app.IngestEvent(ctx, req.EventType, eventData, req.GetSource(), req.Timestamp.AsTime()); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to ingest event: %v", err))
	}

	return &emptypb.Empty{}, nil
}

// IngestBatch 处理批量摄取事件的gRPC请求。
// req: 包含多个事件信息的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) IngestBatch(ctx context.Context, req *pb.IngestBatchRequest) (*emptypb.Empty, error) {
	var events []*entity.IngestedEvent
	for _, reqEvent := range req.Events {
		eventData := fmt.Sprintf("%v", reqEvent.EventData)
		events = append(events, entity.NewIngestedEvent(
			reqEvent.EventType,
			eventData,
			reqEvent.GetSource(),
			reqEvent.Timestamp.AsTime(),
		))
	}

	if err := s.app.IngestBatch(ctx, events); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to batch ingest events: %v", err))
	}

	return &emptypb.Empty{}, nil
}

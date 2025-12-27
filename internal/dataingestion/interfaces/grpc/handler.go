package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/dataingestion/v1"
	"github.com/wyfcoding/ecommerce/internal/dataingestion/application"
	"github.com/wyfcoding/ecommerce/internal/dataingestion/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server 结构体实现了 DataIngestionService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedDataIngestionServer
	app *application.DataIngestionService
}

// NewServer 创建并返回一个新的 DataIngestion gRPC 服务端实例。
func NewServer(app *application.DataIngestionService) *Server {
	return &Server{
		app: app,
	}
}

// IngestEvent 处理摄取单个事件的gRPC请求。
func (s *Server) IngestEvent(ctx context.Context, req *pb.IngestEventRequest) (*emptypb.Empty, error) {
	eventData := fmt.Sprintf("%v", req.EventData)

	if err := s.app.IngestEvent(ctx, req.EventType, eventData, req.GetSource(), req.Timestamp.AsTime()); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to ingest event: %v", err))
	}

	return &emptypb.Empty{}, nil
}

// IngestBatch 处理批量摄取事件的gRPC请求。
func (s *Server) IngestBatch(ctx context.Context, req *pb.IngestBatchRequest) (*emptypb.Empty, error) {
	var events []*domain.IngestedEvent
	for _, reqEvent := range req.Events {
		eventData := fmt.Sprintf("%v", reqEvent.EventData)
		events = append(events, domain.NewIngestedEvent(
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

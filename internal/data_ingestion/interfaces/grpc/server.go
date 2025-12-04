package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/api/data_ingestion/v1" // 导入数据摄取模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/application"

	"google.golang.org/protobuf/types/known/emptypb" // 导入空消息类型。
)

// Server 结构体实现了 DataIngestionService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedDataIngestionServer // 嵌入生成的UnimplementedDataIngestionServer，确保前向兼容性。
	// app *application.DataIngestionService // 依赖DataIngestion应用服务，但当前方法未直接使用。
}

// NewServer 创建并返回一个新的 DataIngestion gRPC 服务端实例。
// app: 依赖DataIngestion应用服务，但当前方法未直接使用。
func NewServer(app *application.DataIngestionService) *Server {
	return &Server{
		// app: app, // 当前方法未直接调用 app 服务。
	}
}

// IngestEvent 处理摄取单个事件的gRPC请求。
// req: 包含事件类型、事件数据、时间戳和来源的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) IngestEvent(ctx context.Context, req *pb.IngestEventRequest) (*emptypb.Empty, error) {
	// 注意：当前应用服务层主要支持“拉取”模型（即注册数据源 -> 触发摄取任务）。
	// 而Proto接口定义了“推送”模型（即直接将事件数据推送过来）。
	// 当前实现仅返回空响应以满足接口，并未将事件数据集成到应用服务层。
	// TODO: 在实际场景中，可能需要扩展应用服务层，添加一个 PushIngestion 或类似的同步摄取方法来处理这些事件。

	// 示例：可以记录事件到日志或转发到消息队列。
	// s.app.PushEvent(ctx, req.EventType, req.EventData, req.Timestamp.AsTime(), req.Source)

	return &emptypb.Empty{}, nil
}

// IngestBatch 处理批量摄取事件的gRPC请求。
// req: 包含多个事件信息的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) IngestBatch(ctx context.Context, req *pb.IngestBatchRequest) (*emptypb.Empty, error) {
	// 同 IngestEvent，此处也仅返回空响应，未实际集成到应用服务层。
	// TODO: 需要扩展应用服务层，处理批量事件的摄取。
	return &emptypb.Empty{}, nil
}

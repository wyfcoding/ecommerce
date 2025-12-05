package grpc

import (
	"context"
	"fmt"     // 导入格式化包，用于错误信息。
	"strconv" // 导入字符串和数字转换工具。

	pb "github.com/wyfcoding/ecommerce/go-api/data_processing/v1"         // 导入数据处理模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/data_processing/application" // 导入数据处理模块的应用服务。

	"google.golang.org/grpc/codes"  // gRPC状态码。
	"google.golang.org/grpc/status" // gRPC状态处理。
)

// Server 结构体实现了 DataProcessingService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedDataProcessingServer                                    // 嵌入生成的UnimplementedDataProcessingServer，确保前向兼容性。
	app                                  *application.DataProcessingService // 依赖DataProcessing应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 DataProcessing gRPC 服务端实例。
func NewServer(app *application.DataProcessingService) *Server {
	return &Server{app: app}
}

// ProcessData 处理数据处理gRPC请求，提交一个数据处理任务。
// req: 包含数据ID、处理类型和处理参数的请求体。
// 返回任务提交响应和可能发生的gRPC错误。
func (s *Server) ProcessData(ctx context.Context, req *pb.ProcessDataRequest) (*pb.ProcessDataResponse, error) {
	// 将Proto请求映射到应用服务层的 SubmitTask 方法。
	// name: 使用 "Process-" + data_id 作为任务名称。
	// taskType: 直接使用 req.ProcessingType。
	// config: 如果需要，可以序列化 req.ProcessingParams。当前简化为字符串。
	// workflowID: 默认使用0，表示不属于任何特定工作流。
	name := "Process-" + req.DataId
	config := "" // TODO: 根据req.ProcessingParams序列化配置信息。

	// 调用应用服务层提交任务。
	task, err := s.app.SubmitTask(ctx, name, req.ProcessingType, config, 0)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to submit data processing task: %v", err))
	}

	// 返回任务提交成功的响应。
	return &pb.ProcessDataResponse{
		ProcessingId: strconv.FormatUint(uint64(task.ID), 10), // 返回任务ID的字符串形式。
		Status:       "SUBMITTED",                             // 表示任务已提交。
	}, nil
}

// GetProcessingStatus 处理获取数据处理任务状态的gRPC请求。
// req: 包含处理ID的请求体。
// 返回任务状态响应和可能发生的gRPC错误。
func (s *Server) GetProcessingStatus(ctx context.Context, req *pb.GetProcessingStatusRequest) (*pb.ProcessingStatusResponse, error) {
	// 注意：应用服务层当前未直接暴露根据ID获取单个任务的方法（GetTask）。
	// ListTasks方法需要workflowID作为过滤条件，不适合直接通过ProcessingId查询。
	// 理想情况下，应用服务层应该提供一个公共的GetTask方法。
	// 此处返回Unimplemented错误。
	return nil, status.Error(codes.Unimplemented, "GetProcessingStatus not implemented yet")
}

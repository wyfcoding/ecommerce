package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/file/v1"           // 导入文件模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/file/application"   // 导入文件模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/file/domain/entity" // 导入文件模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 FileService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedFileServiceServer                          // 嵌入生成的UnimplementedFileServiceServer，确保前向兼容性。
	app                               *application.FileService // 依赖File应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 File gRPC 服务端实例。
func NewServer(app *application.FileService) *Server {
	return &Server{app: app}
}

// UploadFile 处理上传文件的gRPC请求。
// req: 包含文件名称、大小、类型和内容的请求体。
// 返回上传成功的文件元数据响应和可能发生的gRPC错误。
func (s *Server) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.UploadFileResponse, error) {
	// 将protobuf的文件类型字符串转换为领域实体定义的 FileType。
	// 注意：这里进行了直接转换，如果req.Type是未知类型，可能导致错误或默认值。
	// 实际应用中，可能需要增加验证逻辑或映射函数。
	fType := entity.FileType(req.Type)

	// 调用应用服务层上传文件（模拟）。
	file, err := s.app.UploadFile(ctx, req.Name, req.Size, fType, req.Content)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to upload file: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.UploadFileResponse{
		File: convertFileToProto(file),
	}, nil
}

// GetFile 处理获取文件元数据的gRPC请求。
// req: 包含文件ID的请求体。
// 返回文件元数据响应和可能发生的gRPC错误。
func (s *Server) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileResponse, error) {
	file, err := s.app.GetFile(ctx, req.Id)
	if err != nil {
		// 如果文件未找到，返回NotFound状态码。
		return nil, status.Error(codes.NotFound, fmt.Sprintf("file not found: %v", err))
	}
	return &pb.GetFileResponse{
		File: convertFileToProto(file),
	}, nil
}

// DeleteFile 处理删除文件的gRPC请求。
// req: 包含文件ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteFile(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete file: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListFiles 处理列出文件元数据的gRPC请求。
// req: 包含分页参数的请求体。
// 返回文件元数据列表响应和可能发生的gRPC错误。
func (s *Server) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取文件列表。
	files, total, err := s.app.ListFiles(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list files: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbFiles := make([]*pb.FileMetadata, len(files))
	for i, f := range files {
		pbFiles[i] = convertFileToProto(f)
	}

	return &pb.ListFilesResponse{
		Files:      pbFiles,
		TotalCount: uint64(total), // 总记录数。
	}, nil
}

// convertFileToProto 是一个辅助函数，将领域层的 FileMetadata 实体转换为 protobuf 的 FileMetadata 消息。
func convertFileToProto(f *entity.FileMetadata) *pb.FileMetadata {
	if f == nil {
		return nil
	}
	return &pb.FileMetadata{
		Id:        uint64(f.ID),                 // 文件ID。
		Name:      f.Name,                       // 文件名。
		Size:      f.Size,                       // 文件大小。
		Type:      string(f.Type),               // 文件类型。
		Url:       f.URL,                        // 访问URL。
		Bucket:    f.Bucket,                     // 存储桶。
		Checksum:  f.Checksum,                   // 校验和。
		CreatedAt: timestamppb.New(f.CreatedAt), // 创建时间。
		// Proto中还包含 UpdatedAt, DeletedAt, Path 等字段，但实体中没有或未映射。
	}
}

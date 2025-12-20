package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/file/v1"
	"github.com/wyfcoding/ecommerce/internal/file/application"
	"github.com/wyfcoding/ecommerce/internal/file/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 FileService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedFileServiceServer
	app *application.FileService
}

// NewServer 创建并返回一个新的 File gRPC 服务端实例。
func NewServer(app *application.FileService) *Server {
	return &Server{app: app}
}

// UploadFile 处理上传文件的gRPC请求。
func (s *Server) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.UploadFileResponse, error) {
	fType := domain.FileType(req.Type)

	file, err := s.app.UploadFile(ctx, req.Name, req.Size, fType, req.Content)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to upload file: %v", err))
	}

	return &pb.UploadFileResponse{
		File: convertFileToProto(file),
	}, nil
}

// GetFile 处理获取文件元数据的gRPC请求。
func (s *Server) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileResponse, error) {
	file, err := s.app.GetFile(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("file not found: %v", err))
	}
	return &pb.GetFileResponse{
		File: convertFileToProto(file),
	}, nil
}

// DeleteFile 处理删除文件的gRPC请求。
func (s *Server) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteFile(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete file: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListFiles 处理列出文件元数据的gRPC请求。
func (s *Server) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	files, total, err := s.app.ListFiles(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list files: %v", err))
	}

	pbFiles := make([]*pb.FileMetadata, len(files))
	for i, f := range files {
		pbFiles[i] = convertFileToProto(f)
	}

	return &pb.ListFilesResponse{
		Files:      pbFiles,
		TotalCount: uint64(total),
	}, nil
}

func convertFileToProto(f *domain.FileMetadata) *pb.FileMetadata {
	if f == nil {
		return nil
	}
	return &pb.FileMetadata{
		Id:        uint64(f.ID),
		Name:      f.Name,
		Size:      f.Size,
		Type:      string(f.Type),
		Url:       f.URL,
		Bucket:    f.Bucket,
		Checksum:  f.Checksum,
		CreatedAt: timestamppb.New(f.CreatedAt),
	}
}

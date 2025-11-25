package grpc

import (
	"context"
	pb "github.com/wyfcoding/ecommerce/api/file/v1"
	"github.com/wyfcoding/ecommerce/internal/file/application"
	"github.com/wyfcoding/ecommerce/internal/file/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedFileServiceServer
	app *application.FileService
}

func NewServer(app *application.FileService) *Server {
	return &Server{app: app}
}

func (s *Server) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.UploadFileResponse, error) {
	// Service UploadFile(ctx, name, size, fileType, content)

	fType := entity.FileType(req.Type)
	// Default to OTHER if unknown? Service doesn't validate.

	file, err := s.app.UploadFile(ctx, req.Name, req.Size, fType, req.Content)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UploadFileResponse{
		File: convertFileToProto(file),
	}, nil
}

func (s *Server) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileResponse, error) {
	file, err := s.app.GetFile(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetFileResponse{
		File: convertFileToProto(file),
	}, nil
}

func (s *Server) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteFile(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

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
		return nil, status.Error(codes.Internal, err.Error())
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

func convertFileToProto(f *entity.FileMetadata) *pb.FileMetadata {
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

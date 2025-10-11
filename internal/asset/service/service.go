package service

import (
	"bytes" // For converting []byte to io.Reader
	"context"
	"strconv" // For getUserIDFromContext

	v1 "ecommerce/api/asset/v1"
	"ecommerce/internal/asset/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AssetService is the gRPC service implementation for asset management.
type AssetService struct {
	v1.UnimplementedAssetServiceServer
	uc *biz.AssetUsecase
}

// NewAssetService creates a new AssetService.
func NewAssetService(uc *biz.AssetUsecase) *AssetService {
	return &AssetService{uc: uc}
}

// getUserIDFromContext 从 gRPC 上下文的 metadata 中提取用户ID。
func getUserIDFromContext(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "无法获取元数据")
	}
	// 兼容 gRPC-Gateway 在 HTTP 请求时注入的用户ID
	values := md.Get("x-md-global-user-id")
	if len(values) == 0 {
		// 兼容直接 gRPC 调用时注入的用户ID
		values = md.Get("x-user-id")
		if len(values) == 0 {
			return 0, status.Errorf(codes.Unauthenticated, "请求头中缺少 x-user-id 信息")
		}
	}
	userID, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "x-user-id 格式无效")
	}
	return userID, nil
}

// UploadFile implements the UploadFile RPC。
func (s *AssetService) UploadFile(ctx context.Context, req *v1.UploadFileRequest) (*v1.UploadFileResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.FileName == "" || len(req.Content) == 0 || req.ContentType == "" || req.BucketName == "" {
		return nil, status.Error(codes.InvalidArgument, "file_name, content, content_type, and bucket_name are required")
	}

	fileSize := int64(len(req.Content))
	fileContent := bytes.NewReader(req.Content)

	uploadedFile, err := s.uc.UploadFile(ctx, req.FileName, req.ContentType, req.BucketName, fileSize, userID, fileContent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upload file: %v", err)
	}

	return &v1.UploadFileResponse{
		FileUrl: uploadedFile.FilePath,
		FileId:  uploadedFile.FileID,
	}, nil
}

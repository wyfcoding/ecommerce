package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"time"

	v1 "ecommerce/api/asset/v1"
	"ecommerce/internal/asset/model"
	"ecommerce/internal/asset/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// --- 错误定义 ---
var (
	ErrFileNotFound         = errors.New("file not found")
	ErrUploadFailed         = errors.New("file upload failed")
	ErrDeleteFailed         = errors.New("file delete failed")
	ErrInvalidFileType      = errors.New("invalid file type")
	ErrPresignedURLFailed   = errors.New("failed to generate presigned URL")
	ErrImageProcessFailed   = errors.New("image processing failed")
	ErrCDNPurgeFailed       = errors.New("CDN purge failed")
	ErrUnauthorized         = errors.New("unauthorized access")
	ErrPermissionDenied     = errors.New("permission denied")
)

// ObjectStorageClient 定义了与对象存储服务交互的接口。
type ObjectStorageClient interface {
	UploadFile(ctx context.Context, bucket, path, filename, mimeType string, content io.Reader, size int64) (string, error)
	DeleteFile(ctx context.Context, bucket, path string) error
	GetFileDownloadURL(ctx context.Context, bucket, path string) (string, error)
	GeneratePresignedUploadURL(ctx context.Context, bucket, path, mimeType string, expiresIn time.Duration) (string, error)
	GeneratePresignedDownloadURL(ctx context.Context, bucket, path string, expiresIn time.Duration) (string, error)
	// TODO: Add image processing capabilities if object storage supports it (e.g., S3 Lambda, image transforms)
}

// CDNClient 定义了与 CDN 服务交互的接口。
type CDNClient interface {
	PurgeCache(ctx context.Context, urls []string, purgeAll bool) error
}

// AssetService 封装了资产管理相关的业务逻辑。
type AssetService struct {
	v1.UnimplementedAssetServiceServer
	fileMetadataRepo    repository.FileMetadataRepo
	objectStorageClient ObjectStorageClient
	cdnClient           CDNClient
	defaultBucket       string // 默认存储桶
	baseCDNURL          string // CDN基础URL
}

// NewAssetService 是 AssetService 的构造函数。
func NewAssetService(
	fileMetadataRepo repository.FileMetadataRepo,
	objectStorageClient ObjectStorageClient,
	cdnClient CDNClient,
) *AssetService {
	// TODO: 从配置中获取 defaultBucket 和 baseCDNURL
	return &AssetService{
		fileMetadataRepo:    fileMetadataRepo,
		objectStorageClient: objectStorageClient,
		cdnClient:           cdnClient,
		defaultBucket:       "ecommerce-assets", // 示例默认值
		baseCDNURL:          "https://cdn.example.com", // 示例默认值
	}
}

// --- 文件上传与管理 ---

// UploadFile 上传一个文件到存储系统。
func (s *AssetService) UploadFile(ctx context.Context, req *v1.UploadFileRequest) (*v1.UploadFileResponse, error) {
	// 权限检查：只有认证用户或管理员可以上传文件
	uploaderID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Filename == "" || len(req.Content) == 0 {
		zap.S().Warn("upload file request missing filename or content")
		return nil, status.Errorf(codes.InvalidArgument, "filename and content are required")
	}

	// 生成唯一文件ID和存储路径
	fileID := uuid.New().String()
	ext := filepath.Ext(req.Filename)
	storagePath := fmt.Sprintf("%s/%s%s", req.GetFolder(), fileID, ext)

	// 确定 MIME 类型
	mimeType := getMimeType(ext, req.FileType)

	// 上传文件到对象存储
	_, err = s.objectStorageClient.UploadFile(ctx, s.defaultBucket, storagePath, req.Filename, mimeType, bytes.NewReader(req.Content), int64(len(req.Content)))
	if err != nil {
		zap.S().Errorf("failed to upload file %s to object storage: %v", req.Filename, err)
		return nil, status.Errorf(codes.Internal, "file upload failed")
	}

	// 获取文件可访问URL
	fileURL := fmt.Sprintf("%s/%s", s.baseCDNURL, storagePath) // 假设CDN路径与存储路径一致

	// 存储文件元数据到数据库
	metadata := &model.FileMetadata{
		ID:             fileID,
		Filename:       req.Filename,
		Bucket:         s.defaultBucket,
		Path:           storagePath,
		URL:            fileURL,
		FileType:       model.FileType(req.FileType),
		MimeType:       mimeType,
		Size:           int64(len(req.Content)),
		UploaderID:     uploaderID,
		UploadedAt:     time.Now(),
		CustomMetadata: req.CustomMetadata,
	}
	createdMetadata, err := s.fileMetadataRepo.CreateFileMetadata(ctx, metadata)
	if err != nil {
		zap.S().Errorf("failed to save file metadata for %s: %v", fileID, err)
		// 即使元数据保存失败，文件也已上传，需要人工介入处理数据不一致
		return nil, status.Errorf(codes.Internal, "failed to save file metadata")
	}

	zap.S().Infof("file %s uploaded and metadata saved, ID: %s", req.Filename, fileID)
	return &v1.UploadFileResponse{
		FileMetadata: bizFileMetadataToProto(createdMetadata),
	}, nil
}

// GetFileMetadata 获取文件的元数据。
func (s *AssetService) GetFileMetadata(ctx context.Context, req *v1.GetFileMetadataRequest) (*v1.FileMetadataResponse, error) {
	// 权限检查：只有管理员或文件上传者可以查看元数据
	currentUserID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	metadata, err := s.fileMetadataRepo.GetFileMetadataByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get file metadata by id %s: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to get file metadata")
	}
	if metadata == nil {
		return nil, status.Errorf(codes.NotFound, "file metadata with id %s not found", req.Id)
	}

	if !isAdmin(ctx) && metadata.UploaderID != currentUserID {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied to view this file metadata")
	}

	return &v1.FileMetadataResponse{
		FileMetadata: bizFileMetadataToProto(metadata),
	}, nil
}

// DeleteFile 删除存储系统中的文件。
func (s *AssetService) DeleteFile(ctx context.Context, req *v1.DeleteFileRequest) (*emptypb.Empty, error) {
	// 权限检查：只有管理员或文件上传者可以删除文件
	currentUserID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	metadata, err := s.fileMetadataRepo.GetFileMetadataByID(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to get file metadata by id %s for deletion: %v", req.Id, err)
		return nil, status.Errorf(codes.Internal, "failed to get file metadata")
	}
	if metadata == nil {
		return nil, status.Errorf(codes.NotFound, "file metadata with id %s not found", req.Id)
	}

	if !isAdmin(ctx) && metadata.UploaderID != currentUserID {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied to delete this file")
	}

	// 从对象存储中删除文件
	err = s.objectStorageClient.DeleteFile(ctx, metadata.Bucket, metadata.Path)
	if err != nil {
		zap.S().Errorf("failed to delete file %s from object storage: %v", metadata.Path, err)
		return nil, status.Errorf(codes.Internal, "file delete failed")
	}

	// 从数据库中删除元数据
	err = s.fileMetadataRepo.DeleteFileMetadata(ctx, req.Id)
	if err != nil {
		zap.S().Errorf("failed to delete file metadata for %s: %v", req.Id, err)
		// 即使元数据删除失败，文件也已删除，需要人工介入处理数据不一致
		return nil, status.Errorf(codes.Internal, "failed to delete file metadata")
	}

	// 清除CDN缓存 (如果配置了CDN)
	if s.cdnClient != nil {
		err = s.cdnClient.PurgeCache(ctx, []string{metadata.URL}, false)
		if err != nil {
			zap.S().Errorf("failed to purge CDN cache for URL %s: %v", metadata.URL, err)
			// CDN缓存清除失败不影响文件删除的核心逻辑，但会影响用户体验
		}
	}

	zap.S().Infof("file %s deleted successfully", req.Id)
	return &emptypb.Empty{}, nil
}

// ListFiles 获取文件列表。
func (s *AssetService) ListFiles(ctx context.Context, req *v1.ListFilesRequest) (*v1.ListFilesResponse, error) {
	// 权限检查：只有管理员可以查看所有文件，普通用户只能查看自己上传的文件
	currentUserID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if !isAdmin(ctx) {
		// 如果不是管理员，则只能查询自己上传的文件
		if req.HasUploaderId() && req.GetUploaderId() != currentUserID {
			return nil, status.Errorf(codes.PermissionDenied, "permission denied to list other users' files")
		}
		req.UploaderId = currentUserID // 强制设置为当前用户ID
	}

	files, totalCount, err := s.fileMetadataRepo.ListFileMetadata(
		ctx,
		model.FileType(req.GetFileType()),
		req.GetFilenameKeyword(),
		req.GetUploaderId(),
		timestamppbToTime(req.GetStartDate()),
		timestamppbToTime(req.GetEndDate()),
		req.PageSize,
		req.PageToken,
	)
	if err != nil {
		zap.S().Errorf("failed to list files: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list files")
	}

	protoFiles := make([]*v1.FileMetadata, len(files))
	for i, file := range files {
		protoFiles[i] = bizFileMetadataToProto(file)
	}

	return &v1.ListFilesResponse{
		Files:      protoFiles,
		TotalCount: totalCount,
		NextPageToken: req.PageToken + 1,
	}, nil
}

// GeneratePresignedURL 生成一个预签名URL，用于直接上传或下载文件。
func (s *AssetService) GeneratePresignedURL(ctx context.Context, req *v1.GeneratePresignedURLRequest) (*v1.GeneratePresignedURLResponse, error) {
	// 权限检查：生成预签名URL通常需要认证
	currentUserID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Filename == "" {
		zap.S().Warn("generate presigned URL request missing filename")
		return nil, status.Errorf(codes.InvalidArgument, "filename is required")
	}

	expiresIn := time.Duration(req.GetExpiresInSeconds()) * time.Second
	if expiresIn == 0 {
		expiresIn = 1 * time.Hour // 默认1小时
	}

	var presignedURL string
	var fileID string
	var storagePath string

	switch req.Type {
	case v1.PresignedURLType_PRESIGNED_URL_TYPE_UPLOAD:
		fileID = uuid.New().String()
		ext := filepath.Ext(req.Filename)
		storagePath = fmt.Sprintf("uploads/%d/%s%s", currentUserID, fileID, ext)
		presignedURL, err = s.objectStorageClient.GeneratePresignedUploadURL(ctx, s.defaultBucket, storagePath, req.GetContentType(), expiresIn)
		if err != nil {
			zap.S().Errorf("failed to generate presigned upload URL for %s: %v", req.Filename, err)
			return nil, status.Errorf(codes.Internal, "failed to generate presigned upload URL")
		}
		// 预上传成功后，需要先保存元数据，状态为 PENDING 或 UPLOADING
		metadata := &model.FileMetadata{
			ID:             fileID,
			Filename:       req.Filename,
			Bucket:         s.defaultBucket,
			Path:           storagePath,
			FileType:       model.FileTypeUnspecified, // 待上传，类型未知
			MimeType:       req.GetContentType(),
			UploaderID:     currentUserID,
			UploadedAt:     time.Now(),
			CustomMetadata: map[string]string{"status": "pending_upload"},
		}
		_, err = s.fileMetadataRepo.CreateFileMetadata(ctx, metadata)
		if err != nil {
			zap.S().Errorf("failed to save pending upload file metadata for %s: %v", fileID, err)
			// 元数据保存失败，但预签名URL已生成，需要人工介入
			return nil, status.Errorf(codes.Internal, "failed to save pending upload file metadata")
		}

	case v1.PresignedURLType_PRESIGNED_URL_TYPE_DOWNLOAD:
		if req.FileId == "" {
			return nil, status.Errorf(codes.InvalidArgument, "file_id is required for download presigned URL")
		}
		metadata, err := s.fileMetadataRepo.GetFileMetadataByID(ctx, req.FileId)
		if err != nil || metadata == nil {
			zap.S().Errorf("failed to get file metadata for download %s: %v", req.FileId, err)
			return nil, status.Errorf(codes.NotFound, "file not found")
		}
		// 权限检查：只有管理员或文件上传者可以生成下载URL
		if !isAdmin(ctx) && metadata.UploaderID != currentUserID {
			return nil, status.Errorf(codes.PermissionDenied, "permission denied to generate download URL for this file")
		}
		presignedURL, err = s.objectStorageClient.GeneratePresignedDownloadURL(ctx, metadata.Bucket, metadata.Path, expiresIn)
		if err != nil {
			zap.S().Errorf("failed to generate presigned download URL for %s: %v", req.FileId, err)
			return nil, status.Errorf(codes.Internal, "failed to generate presigned download URL")
		}
		fileID = req.FileId

	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid presigned URL type")
	}

	zap.S().Infof("generated presigned URL for %s, type: %s", fileID, req.Type.String())
	return &v1.GeneratePresignedURLResponse{
		Url:    presignedURL,
		FileId: fileID,
	}, nil
}

// --- 图片处理 ---

// GetImageURL 获取指定尺寸和格式的图片URL。
func (s *AssetService) GetImageURL(ctx context.Context, req *v1.GetImageURLRequest) (*v1.ImageURLResponse, error) {
	// 权限检查：图片URL获取通常是公开的，但可以根据需要添加认证

	if req.FileId == "" {
		zap.S().Warn("get image URL request missing file ID")
		return nil, status.Errorf(codes.InvalidArgument, "file_id is required")
	}

	metadata, err := s.fileMetadataRepo.GetFileMetadataByID(ctx, req.FileId)
	if err != nil || metadata == nil {
		zap.S().Errorf("failed to get file metadata for image %s: %v", req.FileId, err)
		return nil, status.Errorf(codes.NotFound, "image file not found")
	}

	if metadata.FileType != model.FileTypeImage {
		return nil, status.Errorf(codes.InvalidArgument, "file is not an image")
	}

	// TODO: 根据请求参数 (width, height, format, crop) 生成处理后的图片URL
	// 这通常通过CDN或对象存储的图片处理服务完成
	processedURL := metadata.URL // 默认返回原始URL

	if req.Width > 0 || req.Height > 0 || req.Format != v1.ImageFormat_IMAGE_FORMAT_UNSPECIFIED || req.Crop {
		// 模拟生成处理后的URL
		processedURL = fmt.Sprintf("%s?width=%d&height=%d&format=%s&crop=%t", metadata.URL, req.Width, req.Height, req.Format.String(), req.Crop)
	}

	zap.S().Debugf("generated image URL for file %s: %s", req.FileId, processedURL)
	return &v1.ImageURLResponse{
		Url: processedURL,
	}, nil
}

// ProcessImage 对图片进行处理 (例如: 缩放, 裁剪, 添加水印)。
func (s *AssetService) ProcessImage(ctx context.Context, req *v1.ProcessImageRequest) (*v1.ProcessImageResponse, error) {
	// 权限检查：只有管理员或文件上传者可以处理图片
	currentUserID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	metadata, err := s.fileMetadataRepo.GetFileMetadataByID(ctx, req.FileId)
	if err != nil || metadata == nil {
		zap.S().Errorf("failed to get file metadata for image processing %s: %v", req.FileId, err)
		return nil, status.Errorf(codes.NotFound, "image file not found")
	}

	if !isAdmin(ctx) && metadata.UploaderID != currentUserID {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied to process this image")
	}

	if metadata.FileType != model.FileTypeImage {
		return nil, status.Errorf(codes.InvalidArgument, "file is not an image")
	}

	// TODO: 调用外部图片处理服务或内部图片处理库
	// 这是一个复杂的业务逻辑，可能涉及图像处理库或云服务API
	zap.S().Infof("processing image %s with operation %s and params %v", req.FileId, req.Operation.String(), req.Params)

	// 模拟处理后的URL和文件ID
	processedImageURL := fmt.Sprintf("%s_processed_%s", metadata.URL, req.Operation.String())
	processedFileID := uuid.New().String() // 假设处理后生成新文件

	// 记录新的文件元数据 (如果生成了新文件)
	newMetadata := &model.FileMetadata{
		ID:             processedFileID,
		Filename:       fmt.Sprintf("processed_%s", metadata.Filename),
		Bucket:         metadata.Bucket,
		Path:           fmt.Sprintf("processed/%s", metadata.Path), // 示例路径
		URL:            processedImageURL,
		FileType:       model.FileTypeImage,
		MimeType:       metadata.MimeType,
		Size:           metadata.Size, // 假设大小不变
		UploaderID:     currentUserID,
		UploadedAt:     time.Now(),
		CustomMetadata: map[string]string{"original_file_id": req.FileId, "operation": req.Operation.String()},
	}
	_, err = s.fileMetadataRepo.CreateFileMetadata(ctx, newMetadata)
	if err != nil {
		zap.S().Errorf("failed to save processed image metadata for %s: %v", processedFileID, err)
		// 即使元数据保存失败，图片也已处理，需要人工介入
	}

	return &v1.ProcessImageResponse{
		ProcessedImageUrl: processedImageURL,
		ProcessedFileId:   processedFileID,
	}, nil
}

// --- CDN集成 ---

// PurgeCDNCache 清除CDN缓存。
func (s *AssetService) PurgeCDNCache(ctx context.Context, req *v1.PurgeCDNCacheRequest) (*emptypb.Empty, error) {
	// 权限检查：只有管理员可以清除CDN缓存
	if !isAdmin(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied to purge CDN cache")
	}

	if s.cdnClient == nil {
		zap.S().Warn("CDN client not configured, cannot purge cache")
		return nil, status.Errorf(codes.FailedPrecondition, "CDN client not configured")
	}

	if len(req.Urls) == 0 && !req.PurgeAll {
		zap.S().Warn("purge CDN cache request missing URLs or purge_all flag")
		return nil, status.Errorf(codes.InvalidArgument, "URLs or purge_all flag is required")
	}

	err := s.cdnClient.PurgeCache(ctx, req.Urls, req.PurgeAll)
	if err != nil {
		zap.S().Errorf("failed to purge CDN cache: %v", err)
		return nil, status.Errorf(codes.Internal, "CDN purge failed")
	}

	zap.S().Infof("CDN cache purged successfully for URLs: %v, purgeAll: %t", req.Urls, req.PurgeAll)
	return &emptypb.Empty{}, nil
}

// --- 辅助函数 ---

// getUserIDFromContext 从 gRPC 上下文的 metadata 中提取用户ID。
func getUserIDFromContext(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "cannot get metadata from context")
	}
	values := md.Get("x-user-id")
	if len(values) == 0 {
		// 尝试从 x-admin-user-id 获取，如果当前是管理员操作
		adminValues := md.Get("x-admin-user-id")
		if len(adminValues) > 0 {
			adminUserID, err := strconv.ParseUint(adminValues[0], 10, 64)
			if err != nil {
				return 0, status.Errorf(codes.Unauthenticated, "invalid x-admin-user-id format")
			}
			return adminUserID, nil
		}
		return 0, status.Errorf(codes.Unauthenticated, "missing user ID in request header")
	}
	userID, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "invalid x-user-id format")
	}
	return userID, nil
}

// isAdmin 从 gRPC 上下文的 metadata 中判断当前请求是否由管理员发起。
func isAdmin(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	// 检查是否存在 x-admin-user-id 头部，并且 x-is-admin 头部为 "true"
	adminIDValues := md.Get("x-admin-user-id")
	isAdminValues := md.Get("x-is-admin")
	return len(adminIDValues) > 0 && len(isAdminValues) > 0 && isAdminValues[0] == "true"
}

// timestamppbToTime 将 protobuf 的 Timestamp 转换为 Go 的 time.Time。
func timestamppbToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// getMimeType 根据文件扩展名和文件类型猜测 MIME 类型。
func getMimeType(ext string, fileType v1.FileType) string {
	switch fileType {
	case v1.FileType_FILE_TYPE_IMAGE:
		switch strings.ToLower(ext) {
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".png":
			return "image/png"
		case ".gif":
			return "image/gif"
		case ".webp":
			return "image/webp"
		}
	case v1.FileType_FILE_TYPE_VIDEO:
		switch strings.ToLower(ext) {
		case ".mp4":
			return "video/mp4"
		case ".webm":
			return "video/webm"
		}
	case v1.FileType_FILE_TYPE_DOCUMENT:
		switch strings.ToLower(ext) {
		case ".pdf":
			return "application/pdf"
		case ".doc", ".docx":
			return "application/msword"
		}
	}
	// 默认或未知类型
	return "application/octet-stream"
}

// bizFileMetadataToProto 将 model.FileMetadata 业务领域模型转换为 v1.FileMetadata API 模型。
func bizFileMetadataToProto(metadata *model.FileMetadata) *v1.FileMetadata {
	if metadata == nil {
		return nil
	}
	return &v1.FileMetadata{
		Id:             metadata.ID,
		Filename:       metadata.Filename,
		Bucket:         metadata.Bucket,
		Path:           metadata.Path,
		Url:            metadata.URL,
		FileType:       v1.FileType(metadata.FileType),
		MimeType:       metadata.MimeType,
		Size:           metadata.Size,
		UploaderId:     metadata.UploaderID,
		UploadedAt:     timestamppb.New(metadata.UploadedAt),
		CustomMetadata: metadata.CustomMetadata,
	}
}
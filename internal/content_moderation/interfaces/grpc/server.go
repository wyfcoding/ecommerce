package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/content_moderation/v1"           // 导入内容审核模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/content_moderation/application"   // 导入内容审核模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain/entity" // 导入内容审核模块的领域实体。

	"google.golang.org/grpc/codes"  // gRPC状态码。
	"google.golang.org/grpc/status" // gRPC状态处理。
)

// Server 结构体实现了 ContentModerationService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedContentModerationServer                                // 嵌入生成的UnimplementedContentModerationServer，确保前向兼容性。
	app                                     *application.ModerationService // 依赖Moderation应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 ContentModeration gRPC 服务端实例。
func NewServer(app *application.ModerationService) *Server {
	return &Server{app: app}
}

// ModerateText 处理文本内容审核的gRPC请求。
// req: 包含待审核文本内容和用户ID的请求体。
// 返回审核结果响应和可能发生的gRPC错误。
func (s *Server) ModerateText(ctx context.Context, req *pb.ModerateTextRequest) (*pb.ModerateTextResponse, error) {
	// 调用应用服务层提交文本内容进行审核。
	// 这里将 contentID 设为0，表示当前请求主要关注文本审核本身，而非特定内容实体的审核。
	record, err := s.app.SubmitContent(ctx, entity.ContentTypeText, 0, req.Text, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to submit text for moderation: %v", err))
	}

	isSafe := false
	var rejectionReason *string // 指向拒绝原因字符串的指针。

	// 根据审核记录的状态映射为protobuf的is_safe字段和拒绝原因。
	switch record.Status {
	case entity.ModerationStatusApproved:
		isSafe = true
	case entity.ModerationStatusRejected:
		isSafe = false
		r := record.RejectReason
		rejectionReason = &r // 如果被拒绝，记录拒绝原因。
	case entity.ModerationStatusPending:
		// 如果状态是待审核，根据Proto期望的is_safe布尔值，暂时将其视为不安全。
		// 实际应用中，可能需要返回一个特定的状态码或等待审核结果。
		isSafe = false
		r := "Pending Review" // 提供待审核的理由。
		rejectionReason = &r
	}

	return &pb.ModerateTextResponse{
		IsSafe:           isSafe,          // 是否安全。
		ModerationLabels: record.AITags,   // AI检测到的标签。
		RejectionReason:  rejectionReason, // 拒绝原因。
	}, nil
}

// ModerateImage 处理图片内容审核的gRPC请求。
// req: 包含待审核图片字节和用户ID的请求体。
// 返回图片审核结果响应和可能发生的gRPC错误。
func (s *Server) ModerateImage(ctx context.Context, req *pb.ModerateImageRequest) (*pb.ModerateImageResponse, error) {
	// 注意：当前应用服务层的 SubmitContent 方法期望接收字符串类型的内容（如图片URL），
	// 而此gRPC请求直接提供了图片的原始字节数据 (raw_image_bytes)。
	// 在没有实现图片上传和转换为URL的功能前，此方法无法直接调用应用服务。
	// TODO: 在调用应用服务之前，需要实现将原始图片字节上传到对象存储（如MinIO）并获取其URL的功能。
	return nil, status.Error(codes.Unimplemented, "ModerateImage not implemented for raw bytes directly; image upload to URL needed first.")
}

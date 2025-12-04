package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain/entity"     // 导入内容审核领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain/repository" // 导入内容审核领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// ModerationService 结构体定义了内容审核相关的应用服务。
// 它协调领域层和基础设施层，处理内容提交审核、人工审核、敏感词管理等业务逻辑。
type ModerationService struct {
	repo   repository.ModerationRepository // 依赖ModerationRepository接口，用于数据持久化操作。
	logger *slog.Logger                    // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewModerationService 创建并返回一个新的 ModerationService 实例。
func NewModerationService(repo repository.ModerationRepository, logger *slog.Logger) *ModerationService {
	return &ModerationService{
		repo:   repo,
		logger: logger,
	}
}

// SubmitContent 提交内容进行审核。
// ctx: 上下文。
// contentType: 待审核内容的类型（例如，“评论”、“商品描述”）。
// contentID: 待审核内容的唯一标识符。
// content: 待审核的实际内容字符串。
// userID: 提交内容的用户ID。
// 返回创建成功的ModerationRecord实体和可能发生的错误。
func (s *ModerationService) SubmitContent(ctx context.Context, contentType entity.ContentType, contentID uint64, content string, userID uint64) (*entity.ModerationRecord, error) {
	record := entity.NewModerationRecord(contentType, contentID, content, userID) // 创建ModerationRecord实体。

	// TODO: 调用AI服务进行预审核。
	// 当前实现是模拟AI审核结果。
	// AI服务可以对内容进行初步分类，例如“安全”、“风险”、“违规”。
	record.SetAIResult(0.1, []string{"safe"}) // 模拟AI结果：置信度0.1，标签“safe”。

	// 通过仓储接口将审核记录保存到数据库。
	if err := s.repo.CreateRecord(ctx, record); err != nil {
		s.logger.ErrorContext(ctx, "failed to create moderation record", "content_type", contentType, "content_id", contentID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "moderation record created successfully", "record_id", record.ID, "content_type", contentType, "content_id", contentID)
	return record, nil
}

// ReviewContent 对内容进行人工审核。
// ctx: 上下文。
// id: 审核记录的ID。
// moderatorID: 执行人工审核的管理员ID。
// approved: 布尔值，表示是否批准内容（true为批准，false为拒绝）。
// reason: 拒绝内容的理由。
// 返回可能发生的错误。
func (s *ModerationService) ReviewContent(ctx context.Context, id uint64, moderatorID uint64, approved bool, reason string) error {
	// 获取审核记录。
	record, err := s.repo.GetRecord(ctx, id)
	if err != nil {
		return err
	}

	// 根据审核结果调用实体方法更新记录状态。
	if approved {
		record.Approve(moderatorID)
	} else {
		record.Reject(moderatorID, reason)
	}

	// 通过仓储接口更新数据库中的审核记录。
	return s.repo.UpdateRecord(ctx, record)
}

// ListPendingRecords 获取所有待人工审核的内容记录列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回待审核记录列表、总数和可能发生的错误。
func (s *ModerationService) ListPendingRecords(ctx context.Context, page, pageSize int) ([]*entity.ModerationRecord, int64, error) {
	offset := (page - 1) * pageSize
	// 查询状态为Pending的审核记录。
	return s.repo.ListRecords(ctx, entity.ModerationStatusPending, offset, pageSize)
}

// AddSensitiveWord 添加一个敏感词到系统。
// ctx: 上下文。
// word: 敏感词字符串。
// category: 敏感词所属的类别。
// level: 敏感词的敏感级别。
// 返回创建成功的SensitiveWord实体和可能发生的错误。
func (s *ModerationService) AddSensitiveWord(ctx context.Context, word, category string, level int8) (*entity.SensitiveWord, error) {
	sensitiveWord := entity.NewSensitiveWord(word, category, level) // 创建SensitiveWord实体。
	if err := s.repo.CreateWord(ctx, sensitiveWord); err != nil {
		s.logger.ErrorContext(ctx, "failed to create sensitive word", "word", word, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "sensitive word created successfully", "word_id", sensitiveWord.ID, "word", word)
	return sensitiveWord, nil
}

// ListSensitiveWords 获取敏感词列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回敏感词列表、总数和可能发生的错误。
func (s *ModerationService) ListSensitiveWords(ctx context.Context, page, pageSize int) ([]*entity.SensitiveWord, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListWords(ctx, offset, pageSize)
}

// DeleteSensitiveWord 根据ID删除一个敏感词。
// ctx: 上下文。
// id: 敏感词的ID。
// 返回可能发生的错误。
func (s *ModerationService) DeleteSensitiveWord(ctx context.Context, id uint64) error {
	return s.repo.DeleteWord(ctx, id)
}

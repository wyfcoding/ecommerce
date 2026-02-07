package application

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	aimodelv1 "github.com/wyfcoding/ecommerce/goapi/aimodel/v1"
	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/structures"
)

// ModerationCommandService 处理内容审核的写操作。
type ModerationCommandService struct {
	repo          domain.ModerationRepository
	publisher     domain.EventPublisher
	logger        *slog.Logger
	sensitiveTrie *algorithm.Trie[*domain.SensitiveWord]
	aimodelCli    aimodelv1.AIModelServiceClient
}

// NewModerationCommandService 创建并返回一个新的 ModerationCommandService 实例。
func NewModerationCommandService(repo domain.ModerationRepository, publisher domain.EventPublisher, logger *slog.Logger, aimodelCli aimodelv1.AIModelServiceClient) *ModerationCommandService {
	return &ModerationCommandService{
		repo:          repo,
		publisher:     publisher,
		logger:        logger,
		sensitiveTrie: algorithm.NewTrie[*domain.SensitiveWord](),
		aimodelCli:    aimodelCli,
	}
}

// SubmitContent 提交内容进行审核。
func (m *ModerationCommandService) SubmitContent(ctx context.Context, contentType domain.ContentType, contentID uint64, content string, userID uint64) (*domain.ModerationRecord, error) {
	record := domain.NewModerationRecord(contentType, contentID, content, userID)

	// 1. 使用 Trie 进行敏感词检测 (简单的分词匹配)
	sensitiveWords := m.CheckSensitiveWords(content)

	if len(sensitiveWords) > 0 {
		// 命中敏感词，直接标记为高风险
		record.SetAIResult(0.95, append([]string{"sensitive_word_detected"}, sensitiveWords...))
	} else if m.aimodelCli != nil {
		// 2. 敏感词未命中，调用 AI 模型进行深度情感/风险分析
		resp, err := m.aimodelCli.AnalyzeReviewSentiment(ctx, &aimodelv1.AnalyzeReviewSentimentRequest{
			ReviewText: content,
		})
		if err != nil {
			m.logger.WarnContext(ctx, "AI model analysis failed, falling back to manual review", "error", err)
			record.SetAIResult(0.5, []string{"ai_analysis_failed"})
		} else {
			// 根据情感得分判定风险：如果是负面 (NEGATIVE)，风险较高
			riskScore := resp.Score
			if resp.Sentiment == aimodelv1.Sentiment_SENTIMENT_NEGATIVE {
				// 转换为风险分，假设 0.8 以上需要审核
				record.SetAIResult(riskScore, []string{"negative_sentiment_detected"})
			} else {
				record.SetAIResult(1.0-riskScore, []string{"safe_sentiment"})
			}
		}
	} else {
		// 未命中且无 AI 客户端，保守起见设为待定
		record.SetAIResult(0.5, []string{"no_ai_service"})
	}

	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveRecordInTx(ctx, tx, record); err != nil {
			m.logger.ErrorContext(ctx, "failed to create moderation record", "content_type", contentType, "content_id", contentID, "error", err)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.ModerationRecordCreatedEvent{
			RecordID:    uint64(record.ID),
			ContentType: record.ContentType,
			ContentID:   record.ContentID,
			UserID:      record.UserID,
			Status:      record.Status,
			Timestamp:   time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.ModerationRecordCreatedEventType, fmt.Sprintf("%d", record.ID), event)
	}); err != nil {
		return nil, err
	}
	m.logger.InfoContext(ctx, "moderation record created successfully", "record_id", record.ID, "content_type", contentType, "content_id", contentID)
	return record, nil
}

// ReviewContent 对内容进行人工审核。
func (m *ModerationCommandService) ReviewContent(ctx context.Context, id uint64, moderatorID uint64, approved bool, reason string) error {
	record, err := m.repo.GetRecord(ctx, id)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("moderation record not found")
	}

	if approved {
		record.Approve(moderatorID)
	} else {
		record.Reject(moderatorID, reason)
	}

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveRecordInTx(ctx, tx, record); err != nil {
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.ModerationRecordUpdatedEvent{
			RecordID:    uint64(record.ID),
			Status:      record.Status,
			ModeratorID: record.ModeratorID,
			Timestamp:   time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.ModerationRecordUpdatedEventType, fmt.Sprintf("%d", record.ID), event)
	})
}

// AddSensitiveWord 添加一个敏感词到系统。
func (m *ModerationCommandService) AddSensitiveWord(ctx context.Context, word, category string, level int8) (*domain.SensitiveWord, error) {
	sensitiveWord := domain.NewSensitiveWord(word, category, level)
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveWordInTx(ctx, tx, sensitiveWord); err != nil {
			m.logger.ErrorContext(ctx, "failed to create sensitive word", "word", word, "error", err)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.SensitiveWordCreatedEvent{
			WordID:    uint64(sensitiveWord.ID),
			Word:      sensitiveWord.Word,
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.SensitiveWordCreatedEventType, fmt.Sprintf("%d", sensitiveWord.ID), event)
	}); err != nil {
		return nil, err
	}

	// 实时更新到内存 Trie（使用小写以匹配检查逻辑）
	m.sensitiveTrie.Insert(strings.ToLower(word), sensitiveWord)

	m.logger.InfoContext(ctx, "sensitive word created successfully", "word_id", sensitiveWord.ID, "word", word)
	return sensitiveWord, nil
}

// DeleteSensitiveWord 根据ID删除一个敏感词。
func (m *ModerationCommandService) DeleteSensitiveWord(ctx context.Context, id uint64) error {
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.DeleteWordInTx(ctx, tx, id); err != nil {
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.SensitiveWordDeletedEvent{
			WordID:    id,
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.SensitiveWordDeletedEventType, fmt.Sprintf("%d", id), event)
	}); err != nil {
		return err
	}

	// 删除后重载内存词库以保持一致
	if err := m.LoadSensitiveWords(ctx); err != nil {
		m.logger.WarnContext(ctx, "failed to reload sensitive words after delete", "error", err)
	}
	return nil
}

// LoadSensitiveWords 加载所有敏感词到内存 Trie 中。
func (m *ModerationCommandService) LoadSensitiveWords(ctx context.Context) error {
	words, _, err := m.repo.ListWords(ctx, 0, 10000) // 假设最多10000个
	if err != nil {
		return err
	}

	newTrie := algorithm.NewTrie[*domain.SensitiveWord]()
	for _, w := range words {
		newTrie.Insert(strings.ToLower(w.Word), w)
	}

	m.sensitiveTrie = newTrie
	return nil
}

var wordRegex = regexp.MustCompile(`\w+`)

// CheckSensitiveWords 检查内容中是否包含敏感词。
func (m *ModerationCommandService) CheckSensitiveWords(content string) []string {
	// 真实化分词逻辑：使用正则提取所有词元，并自动过滤标点符号
	tokens := wordRegex.FindAllString(strings.ToLower(content), -1)
	found := make([]string, 0)

	for _, token := range tokens {
		if _, ok := m.sensitiveTrie.Search(token); ok {
			found = append(found, token)
		}
	}
	return found
}

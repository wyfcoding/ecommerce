package application

import (
	"context"
	"log/slog"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// ModerationManager 处理内容审核的写操作。
type ModerationManager struct {
	repo          domain.ModerationRepository
	logger        *slog.Logger
	sensitiveTrie *algorithm.Trie
}

// NewModerationManager 创建并返回一个新的 ModerationManager 实例。
func NewModerationManager(repo domain.ModerationRepository, logger *slog.Logger) *ModerationManager {
	return &ModerationManager{
		repo:          repo,
		logger:        logger,
		sensitiveTrie: algorithm.NewTrie(),
	}
}

// SubmitContent 提交内容进行审核。
func (m *ModerationManager) SubmitContent(ctx context.Context, contentType domain.ContentType, contentID uint64, content string, userID uint64) (*domain.ModerationRecord, error) {
	record := domain.NewModerationRecord(contentType, contentID, content, userID)

	// 使用 Trie 进行敏感词检测 (简单的分词匹配)
	sensitiveWords := m.CheckSensitiveWords(content)

	if len(sensitiveWords) > 0 {
		// 命中敏感词，直接标记为高风险
		record.SetAIResult(0.95, append([]string{"sensitive_word_detected"}, sensitiveWords...))
	} else {
		// 未命中，模拟 AI 结果为安全
		record.SetAIResult(0.1, []string{"safe"})
	}

	if err := m.repo.CreateRecord(ctx, record); err != nil {
		m.logger.ErrorContext(ctx, "failed to create moderation record", "content_type", contentType, "content_id", contentID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "moderation record created successfully", "record_id", record.ID, "content_type", contentType, "content_id", contentID)
	return record, nil
}

// ReviewContent 对内容进行人工审核。
func (m *ModerationManager) ReviewContent(ctx context.Context, id uint64, moderatorID uint64, approved bool, reason string) error {
	record, err := m.repo.GetRecord(ctx, id)
	if err != nil {
		return err
	}

	if approved {
		record.Approve(moderatorID)
	} else {
		record.Reject(moderatorID, reason)
	}

	return m.repo.UpdateRecord(ctx, record)
}

// AddSensitiveWord 添加一个敏感词到系统。
func (m *ModerationManager) AddSensitiveWord(ctx context.Context, word, category string, level int8) (*domain.SensitiveWord, error) {
	sensitiveWord := domain.NewSensitiveWord(word, category, level)
	if err := m.repo.CreateWord(ctx, sensitiveWord); err != nil {
		m.logger.ErrorContext(ctx, "failed to create sensitive word", "word", word, "error", err)
		return nil, err
	}

	// 实时更新到内存 Trie
	m.sensitiveTrie.Insert(word, sensitiveWord)

	m.logger.InfoContext(ctx, "sensitive word created successfully", "word_id", sensitiveWord.ID, "word", word)
	return sensitiveWord, nil
}

// DeleteSensitiveWord 根据ID删除一个敏感词。
func (m *ModerationManager) DeleteSensitiveWord(ctx context.Context, id uint64) error {
	// 注意：当前 Trie 实现不支持删除，生产环境需支持动态重载或并发安全的 Map/Trie 替换
	return m.repo.DeleteWord(ctx, id)
}

// LoadSensitiveWords 加载所有敏感词到内存 Trie 中。
func (m *ModerationManager) LoadSensitiveWords(ctx context.Context) error {
	words, _, err := m.repo.ListWords(ctx, 0, 10000) // 假设最多10000个
	if err != nil {
		return err
	}

	newTrie := algorithm.NewTrie()
	for _, w := range words {
		newTrie.Insert(w.Word, w)
	}

	m.sensitiveTrie = newTrie
	return nil
}

// CheckSensitiveWords 检查内容中是否包含敏感词 (基于 Trie 的分词精确匹配)。
func (m *ModerationManager) CheckSensitiveWords(content string) []string {
	// 简单的分词逻辑：按空格和标点分割 (实际应根据语言使用更复杂的分词器)
	// 这里为了演示 Trie 用法，仅按空格分割
	tokens := strings.Fields(content)
	found := make([]string, 0)

	for _, token := range tokens {
		// 去除标点符号 (简化处理)
		cleanToken := strings.Trim(token, ".,!?;:\"'")
		if _, ok := m.sensitiveTrie.Search(cleanToken); ok {
			found = append(found, cleanToken)
		}
	}
	return found
}

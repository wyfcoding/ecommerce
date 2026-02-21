// 变更说明：
// 彻底回收内容审核服务(contentmoderation)，CMS内容系统自带自动审核防线。
package domain

import (
	"context"
	"time"
)

type ContentStatus string

const (
	StatusDraft     ContentStatus = "DRAFT"
	StatusPending   ContentStatus = "PENDING_REVIEW"
	StatusPublished ContentStatus = "PUBLISHED"
	StatusBlocked   ContentStatus = "BLOCKED"
)

// ContentItem CMS 中的通用内容实体（商品评价、资讯、社区图文等）。
type ContentItem struct {
	ID        uint64
	Type      string // 如图文: ARTICLE, 视频: VIDEO
	AuthorID  uint64
	Body      string   // 正文文本
	MediaURLs []string // 富媒体附件
	Status    ContentStatus
	AuditMsg  string // 审核驳回原因
	CreatedAt time.Time
}

// NLPAnalyzer 抽象了外置（原 contentmoderation）的 AI/NLP 鉴黄、暴恐与广告过滤。
type NLPAnalyzer interface {
	CheckText(ctx context.Context, text string) (bool, string, error)
	CheckMedia(ctx context.Context, urls []string) (bool, string, error)
}

// Publish 包含了审核逻辑的内容发布模型。
func (c *ContentItem) Publish(ctx context.Context, analyzer NLPAnalyzer) {
	c.Status = StatusPending

	// 进行同步短路拦截审核
	passText, msgText, _ := analyzer.CheckText(ctx, c.Body)
	passMedia, msgMedia, _ := analyzer.CheckMedia(ctx, c.MediaURLs)

	if !passText {
		c.Status = StatusBlocked
		c.AuditMsg = "Text violation: " + msgText
		return
	}
	if !passMedia {
		c.Status = StatusBlocked
		c.AuditMsg = "Media violation: " + msgMedia
		return
	}

	c.Status = StatusPublished
}

type ContentRepository interface {
	SaveContent(ctx context.Context, item *ContentItem) error
}

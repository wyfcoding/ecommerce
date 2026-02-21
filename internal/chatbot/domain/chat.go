// 变更说明：
// 1. 【架构升级】全面支持 RAG (Retrieval-Augmented Generation) 架构
// 2. 【功能增强】新增意图识别 (查订单、申请售后、商品咨询等)
// 3. 【功能增强】新增人工触点转接逻辑 (Handoff)
// 4. 【基础模型】支持通过 LLM Provider 插件接入 OpenAI/本地大模型
package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSessionNotFound = errors.New("chat session not found")
	ErrHandoffFailed   = errors.New("failed to handoff to human agent")
)

// SessionStatus 会话状态
type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "ACTIVE"   // 机器人接待中
	SessionStatusHandoff  SessionStatus = "HANDOFF"  // 转人工服务中
	SessionStatusResolved SessionStatus = "RESOLVED" // 已解决
	SessionStatusClosed   SessionStatus = "CLOSED"   // 已关闭
)

// IntentType 客户意图类型
type IntentType string

const (
	IntentTypeGreeting       IntentType = "GREETING"        // 寒暄
	IntentTypeOrderQuery     IntentType = "ORDER_QUERY"     // 订单查询
	IntentTypeLogistics      IntentType = "LOGISTICS"       // 物流查询
	IntentTypeAfterSales     IntentType = "AFTER_SALES"     // 售后申请/咨询
	IntentTypeProductConsult IntentType = "PRODUCT_CONSULT" // 商品咨询
	IntentTypeComplaint      IntentType = "COMPLAINT"       // 投诉建议
	IntentTypeUnknown        IntentType = "UNKNOWN"         // 未知/闲聊
)

// ChatSession 智能客服会话聚合根。
type ChatSession struct {
	ID           uint64        `json:"id"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	SessionNo    string        `json:"session_no"`
	UserID       uint64        `json:"user_id"`
	Status       SessionStatus `json:"status"`
	AgentID      uint64        `json:"agent_id"` // 转人工后的客服ID
	Topic        string        `json:"topic"`    // 会话主题摘要
	Summary      string        `json:"summary"`  // 会话内容摘要（大模型生成）
	MessageCount int           `json:"message_count"`
	LastMsgAt    time.Time     `json:"last_msg_at"`
	Messages     []*Message    `json:"messages"`
}

// Role 消息角色
type Role string

const (
	RoleSystem    Role = "system"    // 系统人设提示词
	RoleUser      Role = "user"      // 用户输入
	RoleAssistant Role = "assistant" // 大模型回复
	RoleFunction  Role = "function"  // 工具调用返回结果
)

// Message 聊天消息记录。
type Message struct {
	ID         uint64     `json:"id"`
	SessionID  uint64     `json:"session_id"`
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	Tokens     int        `json:"tokens"`      // 消耗的 Token 数
	Intent     IntentType `json:"intent"`      // (如果是用户消息) 识别出的意图
	IsReadable bool       `json:"is_readable"` // 是否对用户可见（系统内置的检索结果不可见）
	CreatedAt  time.Time  `json:"created_at"`
}

// DocumentChunk 知识库检索返回的切片块。
type DocumentChunk struct {
	DocID      string  `json:"doc_id"`
	Content    string  `json:"content"`
	Source     string  `json:"source"`
	Similarity float64 `json:"similarity"` // 向量相似度得分
}

// NewChatSession 创建新会话。
func NewChatSession(sessionNo string, userID uint64) *ChatSession {
	now := time.Now()
	return &ChatSession{
		SessionNo: sessionNo,
		UserID:    userID,
		Status:    SessionStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
		LastMsgAt: now,
		Messages:  make([]*Message, 0),
	}
}

// AddMessage 追加消息到会话。
func (s *ChatSession) AddMessage(role Role, content string, intent IntentType, isReadable bool) *Message {
	now := time.Now()
	msg := &Message{
		SessionID:  s.ID,
		Role:       role,
		Content:    content,
		Intent:     intent,
		IsReadable: isReadable,
		CreatedAt:  now,
	}
	s.Messages = append(s.Messages, msg)
	s.MessageCount++
	s.LastMsgAt = now
	s.UpdatedAt = now
	return msg
}

// TransferToAgent 转接人工客服。
func (s *ChatSession) TransferToAgent(agentID uint64) error {
	if s.Status == SessionStatusClosed || s.Status == SessionStatusResolved {
		return errors.New("cannot transfer closed session")
	}
	s.Status = SessionStatusHandoff
	s.AgentID = agentID
	s.UpdatedAt = time.Now()
	return nil
}

// ----------------- Domain Interfaces -----------------

// SessionRepository 会话存储库
type SessionRepository interface {
	Save(ctx context.Context, session *ChatSession) error
	GetByID(ctx context.Context, id uint64) (*ChatSession, error)
	GetActiveSession(ctx context.Context, userID uint64) (*ChatSession, error)
	FindRecentMessages(ctx context.Context, sessionID uint64, limit int) ([]*Message, error)
}

// VectorDB 向量数据库，用于 RAG 知识检索
type VectorDB interface {
	// Search 检索相似度最高的知识切片
	Search(ctx context.Context, query string, topK int) ([]*DocumentChunk, error)
}

// LLMProvider 大模型推理接口
type LLMProvider interface {
	// GenerateResponse 生成大模型回复
	GenerateResponse(ctx context.Context, history []*Message, contextChunks []*DocumentChunk) (string, int, error)
	// ExtractIntent 分析用户输入的意图
	ExtractIntent(ctx context.Context, query string) (IntentType, error)
}

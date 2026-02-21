// 变更说明：
// 1. 【完整性】实现包含 RAG (检索增强生成) 的对话流程：预处理、意图识别、向量召回、大模型生成、后处理
// 2. 【多轮对话】查询数据库获取上下文并合并历史消息
// 3. 【边界控制】提供转接人工 (Handoff) 逻辑处理
// 4. 【高并发】优化日志与性能追踪
package application

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/chatbot/domain"
	"github.com/wyfcoding/pkg/idgen"
)

// ChatService 智能客服服务，作为编排层。
type ChatService struct {
	llm      domain.LLMProvider
	vectorDB domain.VectorDB
	repo     domain.SessionRepository
	idGen    idgen.Generator
	logger   *slog.Logger
	// Configs
	maxHistory int // 保留多少轮历史对话参与生成
	topK       int // 向量检索命中数量
}

func NewChatService(
	llm domain.LLMProvider,
	vectorDB domain.VectorDB,
	repo domain.SessionRepository,
	idGen idgen.Generator,
	logger *slog.Logger,
) *ChatService {
	return &ChatService{
		llm:        llm,
		vectorDB:   vectorDB,
		repo:       repo,
		idGen:      idGen,
		logger:     logger,
		maxHistory: 10,
		topK:       3,
	}
}

// ChatRequest 用户发送聊天请求 DTO
type ChatRequest struct {
	UserID    uint64 `json:"user_id"`
	SessionID uint64 `json:"session_id"` // 可选，为空则自动创建
	Content   string `json:"content"`
}

// ChatResponse 机器人回复 DTO
type ChatResponse struct {
	SessionID uint64            `json:"session_id"`
	Content   string            `json:"content"`
	Intent    domain.IntentType `json:"intent"`
	Handoff   bool              `json:"handoff"` // 是否已转人工
}

// SendMessage 处理用户消息。
func (s *ChatService) SendMessage(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// 1. 获取或创建会话
	session, err := s.getOrCreateSession(ctx, req.UserID, req.SessionID)
	if err != nil {
		s.logger.Error("failed to get chat session", "user_id", req.UserID, "err", err)
		return nil, err
	}

	// 如果已经在人工队列中，直接将消息转发给人工 (此时不需要大模型处理)
	if session.Status == domain.SessionStatusHandoff {
		session.AddMessage(domain.RoleUser, req.Content, domain.IntentTypeUnknown, true)
		if err := s.repo.Save(ctx, session); err != nil {
			return nil, err
		}
		// TODO: push message to websocket/agent
		return &ChatResponse{
			SessionID: session.ID,
			Content:   "您的问题已转交人工客服，请稍候...",
			Handoff:   true,
		}, nil
	}

	// 2. 意图识别 (提取关键信息)
	intent, err := s.llm.ExtractIntent(ctx, req.Content)
	if err != nil {
		s.logger.Warn("intent extraction failed, fallback to UNKNOWN", "error", err)
		intent = domain.IntentTypeUnknown
	}

	// 记录用户消息
	userMsg := session.AddMessage(domain.RoleUser, req.Content, intent, true)

	// 如果用户明确要求转人工或机器遇到连续客诉，执行 Handoff
	if intent == domain.IntentTypeComplaint || strings.Contains(req.Content, "人工") {
		session.TransferToAgent(0) // 0 表示进入人工排队池
		s.repo.Save(ctx, session)
		return &ChatResponse{
			SessionID: session.ID,
			Content:   "由于您的问题比较复杂，我已为您转接人工客服，请耐心等待。",
			Handoff:   true,
			Intent:    intent,
		}, nil
	}

	// 3. RAG 知识检索 (VectorDB)
	var chunks []*domain.DocumentChunk
	// 根据意图，筛选不同类型的语料库（可优化成动态匹配）
	if intent == domain.IntentTypeProductConsult || intent == domain.IntentTypeAfterSales || intent == domain.IntentTypeLogistics {
		searchResults, err := s.vectorDB.Search(ctx, req.Content, s.topK)
		if err == nil {
			chunks = searchResults
		} else {
			s.logger.Error("vector db search failed", "err", err)
		}
	}

	// 4. 获取历史消息用于上下文
	history, err := s.repo.FindRecentMessages(ctx, session.ID, s.maxHistory)
	if err != nil {
		// 历史不影响主流程，仅打日志
		s.logger.Warn("failed to fetch chat history", "session_id", session.ID, "err", err)
	}
	// append 当前用户消息到 history 尾部
	history = append(history, userMsg)

	// 5. 大模型推理 (Prompt 组装和回答生成)
	responseContent, tokens, err := s.llm.GenerateResponse(ctx, history, chunks)
	if err != nil {
		s.logger.Error("LLM generation failed", "err", err)
		// 兜底回复
		responseContent = "抱歉，我现在脑子有点糊涂，请稍后再试或回复【人工】接入客服。"
	}

	// 6. 保存助手回复并持久化
	assistantMsg := session.AddMessage(domain.RoleAssistant, responseContent, intent, true)
	assistantMsg.Tokens = tokens

	if saveErr := s.repo.Save(ctx, session); saveErr != nil {
		s.logger.Error("failed to save session after assistant reply", "session_id", session.ID, "err", saveErr)
		return nil, saveErr
	}

	return &ChatResponse{
		SessionID: session.ID,
		Content:   responseContent,
		Intent:    intent,
		Handoff:   false,
	}, nil
}

// getOrCreateSession 寻找现有可用对话，或创建新对话
func (s *ChatService) getOrCreateSession(ctx context.Context, userID, sessionID uint64) (*domain.ChatSession, error) {
	if sessionID > 0 {
		return s.repo.GetByID(ctx, sessionID)
	}

	// 查找用户是否有未完结的对话
	active, err := s.repo.GetActiveSession(ctx, userID)
	if err == nil && active != nil {
		return active, nil
	}

	// 创建新对话
	sessionNo := fmt.Sprintf("CHAT-%d", s.idGen.Generate())
	newSession := domain.NewChatSession(sessionNo, userID)
	
	// 这里通常插入一条 System Prompt 作为对话约束，但不作为可见消息
	newSession.AddMessage(domain.RoleSystem, "You are a professional e-commerce AI assistant named 'WyfBot'. Be polite, concise, and helpful. Use the retrieved context documents to answer the user's questions if provided.", domain.IntentTypeUnknown, false)

	if err := s.repo.Save(ctx, newSession); err != nil {
		return nil, err
	}
	return newSession, nil
}

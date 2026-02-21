package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/chatbot/application"
	"github.com/wyfcoding/ecommerce/internal/chatbot/domain"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/server"
)

type simpleIDGen struct{ n atomic.Int64 }

func (g *simpleIDGen) Generate() int64 { return g.n.Add(1) }

type memorySessionRepo struct {
	mu       sync.RWMutex
	sessions map[uint64]*domain.ChatSession
	nextID   uint64
}

func newMemorySessionRepo() *memorySessionRepo {
	return &memorySessionRepo{sessions: make(map[uint64]*domain.ChatSession), nextID: 1}
}

func (r *memorySessionRepo) Save(_ context.Context, session *domain.ChatSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if session.ID == 0 {
		session.ID = r.nextID
		r.nextID++
	}
	session.UpdatedAt = time.Now()
	r.sessions[session.ID] = cloneSession(session)
	return nil
}

func (r *memorySessionRepo) GetByID(_ context.Context, id uint64) (*domain.ChatSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[id]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	return cloneSession(s), nil
}

func (r *memorySessionRepo) GetActiveSession(_ context.Context, userID uint64) (*domain.ChatSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.sessions {
		if s.UserID == userID && s.Status == domain.SessionStatusActive {
			return cloneSession(s), nil
		}
	}
	return nil, domain.ErrSessionNotFound
}

func (r *memorySessionRepo) FindRecentMessages(_ context.Context, sessionID uint64, limit int) ([]*domain.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[sessionID]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	if limit <= 0 || len(s.Messages) <= limit {
		return cloneMessages(s.Messages), nil
	}
	return cloneMessages(s.Messages[len(s.Messages)-limit:]), nil
}

type mockVectorDB struct{}

func (mockVectorDB) Search(_ context.Context, query string, topK int) ([]*domain.DocumentChunk, error) {
	if topK <= 0 {
		topK = 1
	}
	items := make([]*domain.DocumentChunk, 0, topK)
	for i := 0; i < topK; i++ {
		items = append(items, &domain.DocumentChunk{
			DocID:      fmt.Sprintf("doc-%d", i+1),
			Content:    "平台支持订单查询、物流追踪、退款与售后处理。",
			Source:     "kb://faq",
			Similarity: 0.9,
		})
	}
	_ = query
	return items, nil
}

type mockLLM struct{}

func (mockLLM) ExtractIntent(_ context.Context, query string) (domain.IntentType, error) {
	q := strings.ToLower(query)
	switch {
	case strings.Contains(q, "物流"):
		return domain.IntentTypeLogistics, nil
	case strings.Contains(q, "订单"):
		return domain.IntentTypeOrderQuery, nil
	case strings.Contains(q, "退款") || strings.Contains(q, "售后"):
		return domain.IntentTypeAfterSales, nil
	case strings.Contains(q, "投诉"):
		return domain.IntentTypeComplaint, nil
	default:
		return domain.IntentTypeUnknown, nil
	}
}

func (mockLLM) GenerateResponse(_ context.Context, history []*domain.Message, chunks []*domain.DocumentChunk) (string, int, error) {
	last := ""
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == domain.RoleUser {
			last = history[i].Content
			break
		}
	}
	ctxHint := ""
	if len(chunks) > 0 {
		ctxHint = " 我查到知识库信息，可继续提供订单号以便精确查询。"
	}
	return "已收到你的问题：" + last + "。" + ctxHint, 120, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	repo := newMemorySessionRepo()
	service := application.NewChatService(mockLLM{}, mockVectorDB{}, repo, &simpleIDGen{}, logger)

	engine := server.NewDefaultGinEngine(gin.Recovery())
	v1 := engine.Group("/api/v1/chatbot")
	{
		v1.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"status": "ok"})
		})

		v1.POST("/messages", func(c *gin.Context) {
			var req application.ChatRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", err.Error())
				return
			}
			if req.UserID == 0 || strings.TrimSpace(req.Content) == "" {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", "user_id/content are required")
				return
			}
			resp, err := service.SendMessage(c.Request.Context(), &req)
			if err != nil {
				response.Error(c, err)
				return
			}
			response.Success(c, resp)
		})

		v1.GET("/sessions/:id", func(c *gin.Context) {
			id, err := parseUint64(c.Param("id"))
			if err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", err.Error())
				return
			}
			s, err := repo.GetByID(c.Request.Context(), id)
			if err != nil {
				response.ErrorWithStatus(c, http.StatusNotFound, "not found", err.Error())
				return
			}
			response.Success(c, s)
		})
	}

	addr := envOrDefault("CHATBOT_HTTP_ADDR", ":9203")
	srv := server.NewGinServer(engine, addr, logger)
	go func() {
		if err := srv.Start(context.Background()); err != nil {
			slog.Error("server exit", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = srv.Stop(context.Background())
	slog.Info("service chatbot gracefully stopped")
}

func cloneSession(src *domain.ChatSession) *domain.ChatSession {
	if src == nil {
		return nil
	}
	cp := *src
	cp.Messages = cloneMessages(src.Messages)
	return &cp
}

func cloneMessages(src []*domain.Message) []*domain.Message {
	out := make([]*domain.Message, 0, len(src))
	for _, m := range src {
		cp := *m
		out = append(out, &cp)
	}
	return out
}

func parseUint64(raw string) (uint64, error) {
	var n uint64
	_, err := fmt.Sscanf(raw, "%d", &n)
	if err != nil || n == 0 {
		return 0, fmt.Errorf("id must be unsigned integer")
	}
	return n, nil
}

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

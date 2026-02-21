package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/server"
)

type socialRelation struct {
	UserID       string    `json:"user_id"`
	ParentID     string    `json:"parent_id"`
	InvitationID string    `json:"invitation_id"`
	Level        int       `json:"level"`
	CreatedAt    time.Time `json:"created_at"`
}

type socialStore struct {
	mu       sync.RWMutex
	byUser   map[string]*socialRelation
	children map[string][]*socialRelation
}

func newSocialStore() *socialStore {
	return &socialStore{byUser: make(map[string]*socialRelation), children: make(map[string][]*socialRelation)}
}

func (s *socialStore) bind(userID, parentID, invitationID string) (*socialRelation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byUser[userID]; exists {
		return nil, false
	}
	level := 1
	if parentID != "" {
		if parent := s.byUser[parentID]; parent != nil {
			level = parent.Level + 1
		}
	}
	rel := &socialRelation{UserID: userID, ParentID: parentID, InvitationID: invitationID, Level: level, CreatedAt: time.Now().UTC()}
	s.byUser[userID] = rel
	if parentID != "" {
		s.children[parentID] = append(s.children[parentID], rel)
	}
	return rel, true
}

func (s *socialStore) network(userID string) []*socialRelation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := s.children[userID]
	out := make([]*socialRelation, 0, len(list))
	for _, item := range list {
		cp := *item
		out = append(out, &cp)
	}
	return out
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	store := newSocialStore()
	engine := server.NewDefaultGinEngine(gin.Recovery())
	v1 := engine.Group("/api/v1/social")
	{
		v1.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"status": "ok"})
		})
		v1.POST("/bind", func(c *gin.Context) {
			var req struct {
				UserID       string `json:"user_id"`
				ParentID     string `json:"parent_id"`
				InvitationID string `json:"invitation_id"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", err.Error())
				return
			}
			if req.UserID == "" || req.InvitationID == "" {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", "user_id/invitation_id are required")
				return
			}
			rel, ok := store.bind(req.UserID, req.ParentID, req.InvitationID)
			if !ok {
				response.ErrorWithStatus(c, http.StatusConflict, "already bound", "user already has social relation")
				return
			}
			response.Success(c, rel)
		})
		v1.GET("/network/:user_id", func(c *gin.Context) {
			response.Success(c, store.network(c.Param("user_id")))
		})
	}

	addr := envOrDefault("SOCIAL_HTTP_ADDR", ":9208")
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
	slog.Info("service social gracefully stopped")
}

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

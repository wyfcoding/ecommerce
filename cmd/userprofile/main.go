package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/userprofile/domain"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/server"
)

type profileStore struct {
	mu       sync.RWMutex
	profiles map[uint64]*domain.UserProfile
}

func newProfileStore() *profileStore {
	return &profileStore{profiles: make(map[uint64]*domain.UserProfile)}
}

func (s *profileStore) getOrCreate(userID uint64) *domain.UserProfile {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.profiles[userID]; ok {
		return p
	}
	p := domain.NewUserProfile(userID)
	p.ID = idgen.GenID()
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	s.profiles[userID] = p
	return p
}

func (s *profileStore) get(userID uint64) (*domain.UserProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profiles[userID]
	if !ok {
		return nil, false
	}
	return p, true
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	store := newProfileStore()
	engine := server.NewDefaultGinEngine(gin.Recovery())
	v1 := engine.Group("/api/v1/userprofile")
	{
		v1.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"status": "ok"})
		})

		v1.POST("/profiles", func(c *gin.Context) {
			var req struct {
				UserID uint64 `json:"user_id"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", err.Error())
				return
			}
			if req.UserID == 0 {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user_id", "user_id is required")
				return
			}
			response.Success(c, store.getOrCreate(req.UserID))
		})

		v1.GET("/profiles/:user_id", func(c *gin.Context) {
			userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
			if err != nil || userID == 0 {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user_id", "user_id must be unsigned integer")
				return
			}
			p, ok := store.get(userID)
			if !ok {
				response.ErrorWithStatus(c, http.StatusNotFound, "not found", "profile not found")
				return
			}
			response.Success(c, p)
		})

		v1.POST("/profiles/:user_id/tags", func(c *gin.Context) {
			userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
			if err != nil || userID == 0 {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user_id", "user_id must be unsigned integer")
				return
			}
			var req struct {
				TagKey     string  `json:"tag_key"`
				TagValue   string  `json:"tag_value"`
				Category   int8    `json:"category"`
				Source     int8    `json:"source"`
				Confidence float64 `json:"confidence"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", err.Error())
				return
			}
			if req.TagKey == "" || req.TagValue == "" {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", "tag_key/tag_value are required")
				return
			}
			p := store.getOrCreate(userID)
			tag := domain.NewUserTag(p.ID, userID, req.TagKey, req.TagValue, domain.TagCategory(req.Category), domain.TagSource(req.Source))
			tag.ID = idgen.GenID()
			tag.SetConfidence(req.Confidence)
			_ = p.AddTag(tag)
			p.CalculateOverallScore()
			response.Success(c, tag)
		})

		v1.DELETE("/profiles/:user_id/tags/:tag_key", func(c *gin.Context) {
			userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
			if err != nil || userID == 0 {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user_id", "user_id must be unsigned integer")
				return
			}
			p, ok := store.get(userID)
			if !ok {
				response.ErrorWithStatus(c, http.StatusNotFound, "not found", "profile not found")
				return
			}
			_ = p.RemoveTag(c.Param("tag_key"))
			p.CalculateOverallScore()
			response.Success(c, gin.H{"success": true})
		})

		v1.GET("/profiles/:user_id/tags", func(c *gin.Context) {
			userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
			if err != nil || userID == 0 {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user_id", "user_id must be unsigned integer")
				return
			}
			p, ok := store.get(userID)
			if !ok {
				response.ErrorWithStatus(c, http.StatusNotFound, "not found", "profile not found")
				return
			}
			response.Success(c, p.Tags)
		})

		v1.GET("/profiles/:user_id/summary", func(c *gin.Context) {
			userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
			if err != nil || userID == 0 {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user_id", "user_id must be unsigned integer")
				return
			}
			p, ok := store.get(userID)
			if !ok {
				response.ErrorWithStatus(c, http.StatusNotFound, "not found", "profile not found")
				return
			}
			p.CalculateOverallScore()
			response.Success(c, gin.H{
				"user_id":              p.UserID,
				"overall_score":        p.OverallScore,
				"activity_score":       p.ActivityScore,
				"engagement_score":     p.EngagementScore,
				"value_score":          p.ValueScore,
				"profile_completeness": p.ProfileCompleteness,
				"segment":              p.Segment,
				"lifecycle_stage":      p.LifecycleStage,
				"tag_count":            p.TagCount,
				"status":               p.Status.String(),
			})
		})
	}

	addr := envOrDefault("USERPROFILE_HTTP_ADDR", ":9210")
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
	slog.Info("service userprofile gracefully stopped")
}

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

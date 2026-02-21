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
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/jwt"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/security"
	"github.com/wyfcoding/pkg/server"
)

type authUser struct {
	UserID       uint64
	Username     string
	PasswordHash string
	Roles        []string
}

type authStore struct {
	mu      sync.RWMutex
	byName  map[string]*authUser
	issuer  string
	secret  string
	expires time.Duration
}

func newAuthStore() *authStore {
	hash, _ := security.HashPassword("password")
	return &authStore{
		byName: map[string]*authUser{
			"admin": {
				UserID:       1,
				Username:     "admin",
				PasswordHash: hash,
				Roles:        []string{"ADMIN", "USER"},
			},
		},
		issuer:  "ecommerce-auth",
		secret:  nonEmpty(os.Getenv("AUTH_JWT_SECRET"), "change-me-in-production"),
		expires: 2 * time.Hour,
	}
}

func (s *authStore) login(username, password string) (string, int64, uint64, error) {
	s.mu.RLock()
	user, ok := s.byName[username]
	s.mu.RUnlock()
	if !ok || !security.CheckPassword(password, user.PasswordHash) {
		return "", 0, 0, http.ErrNoCookie
	}
	token, err := jwt.GenerateToken(user.UserID, user.Username, user.Roles, s.secret, s.issuer, s.expires)
	if err != nil {
		return "", 0, 0, err
	}
	return token, int64(s.expires.Seconds()), user.UserID, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	store := newAuthStore()
	engine := server.NewDefaultGinEngine(gin.Recovery())
	v1 := engine.Group("/api/v1/auth")
	{
		v1.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"status": "ok"})
		})

		v1.POST("/login", func(c *gin.Context) {
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", err.Error())
				return
			}
			if req.Username == "" || req.Password == "" {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid credentials", "username/password are required")
				return
			}
			token, expiresIn, userID, err := store.login(req.Username, req.Password)
			if err != nil {
				response.ErrorWithStatus(c, http.StatusUnauthorized, "unauthorized", "invalid username or password")
				return
			}
			response.Success(c, gin.H{
				"access_token": token,
				"expires_in":   expiresIn,
				"token_type":   "Bearer",
				"user_id":      userID,
			})
		})

		v1.POST("/users", func(c *gin.Context) {
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", err.Error())
				return
			}
			if req.Username == "" || req.Password == "" {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", "username/password are required")
				return
			}
			hash, err := security.HashPassword(req.Password)
			if err != nil {
				response.Error(c, err)
				return
			}
			store.mu.Lock()
			defer store.mu.Unlock()
			if _, exists := store.byName[req.Username]; exists {
				response.ErrorWithStatus(c, http.StatusConflict, "user exists", "username already exists")
				return
			}
			uid := idgen.GenID()
			store.byName[req.Username] = &authUser{UserID: uid, Username: req.Username, PasswordHash: hash, Roles: []string{"USER"}}
			response.Success(c, gin.H{"user_id": uid, "username": req.Username})
		})
	}

	addr := nonEmpty(os.Getenv("AUTH_HTTP_ADDR"), ":9201")
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
	slog.Info("service auth gracefully stopped")
}

func nonEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

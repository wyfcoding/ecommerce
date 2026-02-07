// 生成摘要：实现 Websocket HTTP 处理器，用于升级 HTTP 连接到 Websocket。
// 假设：使用 JWT 令牌进行身份验证，令牌可通过 query 参数或 Authorization 头传递。
package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/wyfcoding/pkg/jwt"
	"github.com/wyfcoding/pkg/server"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebsocketHandler 处理 Websocket 连接请求。
type WebsocketHandler struct {
	wsMgr     *server.WSManager
	jwtSecret string
	logger    *slog.Logger
}

// NewWebsocketHandler 创建 WebsocketHandler 实例。
func NewWebsocketHandler(wsMgr *server.WSManager, jwtSecret string, logger *slog.Logger) *WebsocketHandler {
	return &WebsocketHandler{
		wsMgr:     wsMgr,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// HandleWebsocket 处理 Websocket 升级请求。
func (h *WebsocketHandler) HandleWebsocket(c *gin.Context) {
	// 1. 获取 JWT 令牌
	token := c.Query("token")
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// 2. 解析 JWT 获取用户 ID
	claims, err := jwt.ParseToken(token, h.jwtSecret)
	if err != nil {
		h.logger.Error("failed to parse jwt token", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	userID := claims.UserID
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
		return
	}

	// 3. 升级 HTTP 连接到 Websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("failed to upgrade websocket", "error", err)
		return
	}

	// 4. 创建客户端并注册到 Manager
	client := &server.Client{
		Manager: h.wsMgr,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		UserID:  strconv.FormatUint(userID, 10),
		Topics:  make(map[string]struct{}),
	}

	h.wsMgr.Register(client)

	// 5. 启动读写协程
	go client.WritePump()
	go client.ReadPump()

	h.logger.Info("websocket connection established", "user_id", userID)
}

// RegisterRoutes 注册 Websocket 路由。
func (h *WebsocketHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/ws/notifications", h.HandleWebsocket)
}

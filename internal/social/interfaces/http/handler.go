package http

import (
	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/social/application"
	"github.com/wyfcoding/pkg/response"
)

// 生成摘要：实现 Social 服务的 HTTP 接口。
// 关键改动：请求参数绑定、调用应用服务并格式化返回结果。

type SocialHandler struct {
	appService *application.SocialAppService
}

func NewSocialHandler(app *application.SocialAppService) *SocialHandler {
	return &SocialHandler{appService: app}
}

type BindRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	ParentID     string `json:"parent_id"`
	InvitationID string `json:"invitation_id" binding:"required"`
}

// Bind 绑定社交关系。
func (h *SocialHandler) Bind(c *gin.Context) {
	var req BindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	err := h.appService.BindRelation(c.Request.Context(), req.UserID, req.ParentID, req.InvitationID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, "relation bound")
}

// GetNetwork 获取直属下级。
func (h *SocialHandler) GetNetwork(c *gin.Context) {
	userID := c.Param("user_id")
	list, err := h.appService.GetUserNetwork(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, list)
}

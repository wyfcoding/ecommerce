package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/groupbuy/service"
	"ecommerce/pkg/response"
)

// GroupBuyHandler 拼团HTTP处理器
type GroupBuyHandler struct {
	service service.GroupBuyService
	logger  *zap.Logger
}

// NewGroupBuyHandler 创建拼团HTTP处理器
func NewGroupBuyHandler(service service.GroupBuyService, logger *zap.Logger) *GroupBuyHandler {
	return &GroupBuyHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *GroupBuyHandler) RegisterRoutes(r *gin.RouterGroup) {
	groupbuy := r.Group("/groupbuy")
	{
		// 拼团活动
		groupbuy.GET("/activities", h.ListActivities)
		groupbuy.GET("/activities/:id", h.GetActivity)
		
		// 拼团
		groupbuy.POST("/groups", h.StartGroup)
		groupbuy.POST("/groups/:id/join", h.JoinGroup)
		groupbuy.GET("/groups/:id", h.GetGroup)
		groupbuy.GET("/groups", h.ListGroups)
		groupbuy.GET("/groups/:id/members", h.GetGroupMembers)
	}
}

// ListActivities 获取拼团活动列表
func (h *GroupBuyHandler) ListActivities(c *gin.Context) {
	status := c.Query("status")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))

	activities, total, err := h.service.ListActivities(c.Request.Context(), status, int32(pageSize), int32(pageNum))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取活动列表失败", err)
		return
	}

	response.SuccessWithPagination(c, activities, total, int32(pageNum), int32(pageSize))
}

// GetActivity 获取拼团活动详情
func (h *GroupBuyHandler) GetActivity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	activity, err := h.service.GetActivity(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "活动不存在", err)
		return
	}

	response.Success(c, activity)
}

// StartGroup 发起拼团
func (h *GroupBuyHandler) StartGroup(c *gin.Context) {
	var req struct {
		ActivityID uint64 `json:"activityId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	userID := c.GetUint64("userID")

	group, err := h.service.StartGroup(c.Request.Context(), userID, req.ActivityID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "发起拼团失败", err)
		return
	}

	response.Success(c, group)
}

// JoinGroup 参与拼团
func (h *GroupBuyHandler) JoinGroup(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	userID := c.GetUint64("userID")

	if err := h.service.JoinGroup(c.Request.Context(), userID, groupID); err != nil {
		response.Error(c, http.StatusInternalServerError, "参与拼团失败", err)
		return
	}

	response.Success(c, nil)
}

// GetGroup 获取拼团详情
func (h *GroupBuyHandler) GetGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	group, err := h.service.GetGroup(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "拼团不存在", err)
		return
	}

	response.Success(c, group)
}

// ListGroups 获取拼团列表
func (h *GroupBuyHandler) ListGroups(c *gin.Context) {
	activityID, _ := strconv.ParseUint(c.Query("activityId"), 10, 64)
	status := c.Query("status")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))

	groups, total, err := h.service.ListGroups(c.Request.Context(), activityID, status, int32(pageSize), int32(pageNum))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取拼团列表失败", err)
		return
	}

	response.SuccessWithPagination(c, groups, total, int32(pageNum), int32(pageSize))
}

// GetGroupMembers 获取拼团成员
func (h *GroupBuyHandler) GetGroupMembers(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	members, err := h.service.GetGroupMembers(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取成员列表失败", err)
		return
	}

	response.Success(c, members)
}

package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/content_moderation/application"   // 导入内容审核模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain/entity" // 导入内容审核模块的领域实体。
	"github.com/wyfcoding/ecommerce/pkg/response"                              // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了ContentModeration模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.ModerationService // 依赖Moderation应用服务，处理核心业务逻辑。
	logger  *slog.Logger                   // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 ContentModeration HTTP Handler 实例。
func NewHandler(service *application.ModerationService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// SubmitContent 处理提交内容审核的HTTP请求。
// Method: POST
// Path: /moderation/submit
func (h *Handler) SubmitContent(c *gin.Context) {
	// 定义请求体结构，用于接收待审核内容的信息。
	var req struct {
		ContentType string `json:"content_type" binding:"required"` // 内容类型，必填。
		ContentID   uint64 `json:"content_id" binding:"required"`   // 内容ID，必填。
		Content     string `json:"content" binding:"required"`      // 内容字符串，必填。
		UserID      uint64 `json:"user_id" binding:"required"`      // 提交用户ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层提交内容进行审核。
	record, err := h.service.SubmitContent(c.Request.Context(), entity.ContentType(req.ContentType), req.ContentID, req.Content, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to submit content", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to submit content", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Content submitted successfully", record)
}

// ReviewContent 处理人工审核内容的HTTP请求。
// Method: POST
// Path: /moderation/review/:id
func (h *Handler) ReviewContent(c *gin.Context) {
	// 从URL路径中解析审核记录ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收审核结果。
	var req struct {
		ModeratorID uint64 `json:"moderator_id" binding:"required"` // 审核人ID，必填。
		Approved    bool   `json:"approved"`                        // 是否通过审核。
		Reason      string `json:"reason"`                          // 拒绝理由，如果拒绝则必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层进行人工审核。
	err = h.service.ReviewContent(c.Request.Context(), id, req.ModeratorID, req.Approved, req.Reason)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to review content", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to review content", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Content reviewed successfully", nil)
}

// ListPendingRecords 处理获取待审核内容列表的HTTP请求。
// Method: GET
// Path: /moderation/pending
func (h *Handler) ListPendingRecords(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取待审核记录列表。
	list, total, err := h.service.ListPendingRecords(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list pending records", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list pending records", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Pending records listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AddSensitiveWord 处理添加敏感词的HTTP请求。
// Method: POST
// Path: /moderation/sensitive-words
func (h *Handler) AddSensitiveWord(c *gin.Context) {
	// 定义请求体结构，用于接收敏感词信息。
	var req struct {
		Word     string `json:"word" binding:"required"`     // 敏感词，必填。
		Category string `json:"category" binding:"required"` // 分类，必填。
		Level    int8   `json:"level" binding:"required"`    // 敏感等级，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层添加敏感词。
	word, err := h.service.AddSensitiveWord(c.Request.Context(), req.Word, req.Category, req.Level)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add sensitive word", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add sensitive word", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Sensitive word added successfully", word)
}

// ListSensitiveWords 处理获取敏感词列表的HTTP请求。
// Method: GET
// Path: /moderation/sensitive-words
func (h *Handler) ListSensitiveWords(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 调用应用服务层获取敏感词列表。
	list, total, err := h.service.ListSensitiveWords(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list sensitive words", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list sensitive words", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Sensitive words listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteSensitiveWord 处理删除敏感词的HTTP请求。
// Method: DELETE
// Path: /moderation/sensitive-words/:id
func (h *Handler) DeleteSensitiveWord(c *gin.Context) {
	// 从URL路径中解析敏感词ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层删除敏感词。
	err = h.service.DeleteSensitiveWord(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to delete sensitive word", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete sensitive word", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Sensitive word deleted successfully", nil)
}

// RegisterRoutes 在给定的Gin路由组中注册ContentModeration模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /moderation 路由组，用于所有内容审核相关接口。
	group := r.Group("/moderation")
	{
		// 内容审核记录接口。
		group.POST("/submit", h.SubmitContent)      // 提交内容审核。
		group.POST("/review/:id", h.ReviewContent)  // 人工审核内容。
		group.GET("/pending", h.ListPendingRecords) // 获取待审核列表。

		// 敏感词管理接口。
		group.POST("/sensitive-words", h.AddSensitiveWord)          // 添加敏感词。
		group.GET("/sensitive-words", h.ListSensitiveWords)         // 获取敏感词列表。
		group.DELETE("/sensitive-words/:id", h.DeleteSensitiveWord) // 删除敏感词。
		// TODO: 补充更新敏感词、获取敏感词详情的接口。
	}
}

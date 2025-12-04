package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/ai_model/application"       // 导入AI模型模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/ai_model/domain/entity"     // 导入AI模型模块的领域实体。
	"github.com/wyfcoding/ecommerce/internal/ai_model/domain/repository" // 导入AI模型模块的领域仓储查询对象。
	"github.com/wyfcoding/ecommerce/pkg/response"                        // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了AI模型模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.AIModelService // 依赖AIModel应用服务，处理核心业务逻辑。
	logger  *slog.Logger                // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 AIModel HTTP Handler 实例。
func NewHandler(service *application.AIModelService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateModel 处理创建AI模型的HTTP请求。
// Method: POST
// Path: /ai-models
func (h *Handler) CreateModel(c *gin.Context) {
	// 定义请求体结构，用于接收AI模型的创建信息。
	var req struct {
		Name        string `json:"name" binding:"required"`       // 模型名称，必填。
		Description string `json:"description"`                   // 模型描述，选填。
		Type        string `json:"type" binding:"required"`       // 模型类型，必填。
		Algorithm   string `json:"algorithm" binding:"required"`  // 使用算法，必填。
		CreatorID   uint64 `json:"creator_id" binding:"required"` // 创建者ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建AI模型。
	model, err := h.service.CreateModel(c.Request.Context(), req.Name, req.Description, req.Type, req.Algorithm, req.CreatorID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create model", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create model", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Model created successfully", model)
}

// ListModels 处理获取AI模型列表的HTTP请求。
// Method: GET
// Path: /ai-models
// 支持分页和基于状态、类型、算法、创建者ID的过滤。
func (h *Handler) ListModels(c *gin.Context) {
	// 从查询参数中获取分页和过滤条件，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.DefaultQuery("status", "")
	modelType := c.DefaultQuery("type", "")
	algorithm := c.DefaultQuery("algorithm", "")
	creatorID, _ := strconv.ParseUint(c.DefaultQuery("creator_id", "0"), 10, 64)

	// 构建查询对象。
	query := &repository.ModelQuery{
		Status:    entity.ModelStatus(status),
		Type:      modelType,
		Algorithm: algorithm,
		CreatorID: creatorID,
		Page:      page,
		PageSize:  pageSize,
	}

	// 调用应用服务层获取AI模型列表。
	list, total, err := h.service.ListModels(c.Request.Context(), query)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list models", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list models", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Models listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetModelDetails 处理获取AI模型详情的HTTP请求。
// Method: GET
// Path: /ai-models/:id
func (h *Handler) GetModelDetails(c *gin.Context) {
	// 从URL路径中解析模型ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层获取模型详情。
	model, err := h.service.GetModelDetails(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get model details", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get model details", err.Error())
		return
	}

	// 返回成功的响应，包含模型详情。
	response.SuccessWithStatus(c, http.StatusOK, "Model details retrieved successfully", model)
}

// StartTraining 处理启动AI模型训练的HTTP请求。
// Method: POST
// Path: /ai-models/:id/train
func (h *Handler) StartTraining(c *gin.Context) {
	// 从URL路径中解析模型ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层启动训练。
	if err := h.service.StartTraining(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to start training", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to start training", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Training started successfully", nil)
}

// Predict 处理AI模型预测的HTTP请求。
// Method: POST
// Path: /ai-models/:id/predict
func (h *Handler) Predict(c *gin.Context) {
	// 从URL路径中解析模型ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收预测输入数据和用户ID。
	var req struct {
		Input  string `json:"input" binding:"required"`   // 输入数据，必填。
		UserID uint64 `json:"user_id" binding:"required"` // 用户ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层进行预测。
	output, confidence, err := h.service.Predict(c.Request.Context(), id, req.Input, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to predict", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to predict", err.Error())
		return
	}

	// 返回成功的响应，包含预测结果和置信度。
	response.SuccessWithStatus(c, http.StatusOK, "Prediction successful", gin.H{
		"output":     output,
		"confidence": confidence,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册AI模型模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /ai-models 路由组，用于所有AI模型相关接口。
	group := r.Group("/ai-models")
	{
		group.POST("", h.CreateModel)             // 创建AI模型。
		group.GET("", h.ListModels)               // 获取AI模型列表。
		group.GET("/:id", h.GetModelDetails)      // 获取AI模型详情。
		group.POST("/:id/train", h.StartTraining) // 启动AI模型训练。
		group.POST("/:id/predict", h.Predict)     // 进行AI模型预测。
	}
}

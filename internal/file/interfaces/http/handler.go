package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/file/application"   // 导入文件模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/file/domain/entity" // 导入文件模块的领域实体。
	"github.com/wyfcoding/ecommerce/pkg/response"                // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了File模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.FileService // 依赖File应用服务，处理核心业务逻辑。
	logger  *slog.Logger             // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 File HTTP Handler 实例。
func NewHandler(service *application.FileService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// UploadFile 处理文件上传的HTTP请求。
// Method: POST
// Path: /files/upload
func (h *Handler) UploadFile(c *gin.Context) {
	// 从请求中获取上传的文件。
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "No file uploaded", err.Error())
		return
	}

	// 注意：当前实现为模拟上传，只读取了文件的元数据（名称、大小）。
	// 在生产环境中，需要读取文件内容并将其存储到实际的对象存储服务（如MinIO）。
	// fileHeader.Open() 可以获取文件内容的io.Reader。
	// simplified type detection: 简化的文件类型检测，默认为 Other。
	fileType := entity.FileTypeOther

	// 调用应用服务层上传文件元数据（模拟文件存储）。
	// 此处content参数为nil，表示未处理实际文件内容。
	metadata, err := h.service.UploadFile(c.Request.Context(), fileHeader.Filename, fileHeader.Size, fileType, nil)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to upload file", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to upload file", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "File uploaded successfully", metadata)
}

// GetFile 处理获取文件元数据信息的HTTP请求。
// Method: GET
// Path: /files/:id
func (h *Handler) GetFile(c *gin.Context) {
	// 从URL路径中解析文件ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层获取文件元数据。
	file, err := h.service.GetFile(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get file", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get file", err.Error())
		return
	}

	// 返回成功的响应，包含文件元数据。
	response.SuccessWithStatus(c, http.StatusOK, "File retrieved successfully", file)
}

// DeleteFile 处理删除文件的HTTP请求。
// Method: DELETE
// Path: /files/:id
func (h *Handler) DeleteFile(c *gin.Context) {
	// 从URL路径中解析文件ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层删除文件。
	if err := h.service.DeleteFile(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to delete file", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete file", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "File deleted successfully", nil)
}

// ListFiles 处理获取文件列表的HTTP请求。
// Method: GET
// Path: /files
func (h *Handler) ListFiles(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取文件列表。
	list, total, err := h.service.ListFiles(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list files", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list files", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Files listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册File模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /files 路由组，用于所有文件管理相关接口。
	group := r.Group("/files")
	{
		group.POST("/upload", h.UploadFile) // 上传文件。
		group.GET("/:id", h.GetFile)        // 获取文件元数据。
		group.DELETE("/:id", h.DeleteFile)  // 删除文件。
		group.GET("", h.ListFiles)          // 获取文件列表。
		// TODO: 补充更新文件元数据、下载文件等接口。
	}
}

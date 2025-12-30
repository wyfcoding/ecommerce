package http

import (
	"net/http"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/permission/application"
	"github.com/wyfcoding/pkg/response"
)

// Handler 处理 HTTP 或 gRPC 请求。
type Handler struct {
	app    *application.PermissionService
	logger *slog.Logger
}

// NewHandler 处理 HTTP 或 gRPC 请求。
func NewHandler(app *application.PermissionService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	roles := router.Group("/roles")
	{
		roles.POST("", h.CreateRole)
		roles.GET("/:id", h.GetRole)
		roles.GET("", h.ListRoles)
		roles.DELETE("/:id", h.DeleteRole)
	}

	permissions := router.Group("/permissions")
	{
		permissions.POST("", h.CreatePermission)
		permissions.GET("", h.ListPermissions)
	}

	users := router.Group("/users")
	{
		users.POST("/:id/roles", h.AssignRole)
		users.DELETE("/:id/roles", h.RevokeRole)
		users.GET("/:id/roles", h.GetUserRoles)
		users.GET("/:id/permissions/check", h.CheckPermission)
	}
}

type createRoleRequest struct {
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	PermissionIDs []uint64 `json:"permission_ids"`
}

func (h *Handler) CreateRole(c *gin.Context) {
	var req createRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body: "+err.Error(), "")
		return
	}

	role, err := h.app.CreateRole(c.Request.Context(), req.Name, req.Description, req.PermissionIDs)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create role", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to create role: "+err.Error(), "")
		return
	}

	response.Success(c, role)
}

func (h *Handler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid role id: "+err.Error(), "")
		return
	}

	role, err := h.app.GetRole(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get role", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to get role: "+err.Error(), "")
		return
	}
	if role == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "role not found", "")
		return
	}

	response.Success(c, role)
}

func (h *Handler) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	roles, total, err := h.app.ListRoles(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list roles", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to list roles: "+err.Error(), "")
		return
	}

	response.SuccessWithPagination(c, roles, total, int32(page), int32(pageSize))
}

func (h *Handler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid role id: "+err.Error(), "")
		return
	}

	if err := h.app.DeleteRole(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to delete role", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to delete role: "+err.Error(), "")
		return
	}

	response.Success(c, nil)
}

type createPermissionRequest struct {
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

func (h *Handler) CreatePermission(c *gin.Context) {
	var req createPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body: "+err.Error(), "")
		return
	}

	permission, err := h.app.CreatePermission(c.Request.Context(), req.Code, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create permission", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to create permission: "+err.Error(), "")
		return
	}

	response.Success(c, permission)
}

func (h *Handler) ListPermissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	permissions, total, err := h.app.ListPermissions(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list permissions", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to list permissions: "+err.Error(), "")
		return
	}

	response.SuccessWithPagination(c, permissions, total, int32(page), int32(pageSize))
}

type assignRoleRequest struct {
	RoleID uint64 `json:"role_id" binding:"required"`
}

func (h *Handler) AssignRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user id: "+err.Error(), "")
		return
	}

	var req assignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body: "+err.Error(), "")
		return
	}

	if err := h.app.AssignRole(c.Request.Context(), userID, req.RoleID); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to assign role", "user_id", userID, "role_id", req.RoleID, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to assign role: "+err.Error(), "")
		return
	}

	response.Success(c, nil)
}

func (h *Handler) RevokeRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user id: "+err.Error(), "")
		return
	}

	var req assignRoleRequest // Reuse struct as it has RoleID
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body: "+err.Error(), "")
		return
	}

	if err := h.app.RevokeRole(c.Request.Context(), userID, req.RoleID); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to revoke role", "user_id", userID, "role_id", req.RoleID, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to revoke role: "+err.Error(), "")
		return
	}

	response.Success(c, nil)
}

func (h *Handler) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user id: "+err.Error(), "")
		return
	}

	roles, err := h.app.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get user roles", "user_id", userID, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to get user roles: "+err.Error(), "")
		return
	}

	response.Success(c, roles)
}

func (h *Handler) CheckPermission(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user id: "+err.Error(), "")
		return
	}

	permissionCode := c.Query("code")
	if permissionCode == "" {
		response.ErrorWithStatus(c, http.StatusBadRequest, "permission code required", "")
		return
	}

	allowed, err := h.app.CheckPermission(c.Request.Context(), userID, permissionCode)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to check permission", "user_id", userID, "code", permissionCode, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to check permission: "+err.Error(), "")
		return
	}

	response.Success(c, gin.H{"allowed": allowed})
}

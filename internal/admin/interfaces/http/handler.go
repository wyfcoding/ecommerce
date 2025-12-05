package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/admin/application" // 导入Admin模块的应用服务。
	"github.com/wyfcoding/ecommerce/pkg/response"               // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了Admin模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.AdminService // 依赖Admin应用服务，处理核心业务逻辑。
	logger  *slog.Logger              // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Admin HTTP Handler 实例。
func NewHandler(service *application.AdminService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterAdmin 处理管理员注册的HTTP请求。
// Method: POST
// Path: /auth/register
func (h *Handler) RegisterAdmin(c *gin.Context) {
	// 定义请求体结构，使用 Gin 的 binding 标签进行参数绑定和验证。
	var req struct {
		Username string `json:"username" binding:"required"`       // 用户名，必填。
		Email    string `json:"email" binding:"required,email"`    // 邮箱，必填且格式为邮箱。
		Password string `json:"password" binding:"required,min=6"` // 密码，必填且最小长度为6。
		RealName string `json:"real_name"`                         // 真实姓名，选填。
		Phone    string `json:"phone"`                             // 手机号，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层进行管理员注册。
	admin, err := h.service.RegisterAdmin(c.Request.Context(), req.Username, req.Email, req.Password, req.RealName, req.Phone)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to register admin", "error", err)
		// 根据错误类型返回不同的HTTP状态码，例如，如果用户名或邮箱已存在，可以返回409 Conflict。
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register admin", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Admin registered successfully", admin)
}

// Login 处理管理员登录的HTTP请求。
// Method: POST
// Path: /auth/login
func (h *Handler) Login(c *gin.Context) {
	// 定义请求体结构，用于用户名和密码。
	var req struct {
		Username string `json:"username" binding:"required"` // 用户名，必填。
		Password string `json:"password" binding:"required"` // 密码，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层进行登录，并获取JWT令牌。
	token, err := h.service.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP())
	if err != nil {
		h.logger.WarnContext(c.Request.Context(), "Login failed", "username", req.Username, "error", err)
		// 登录失败通常返回401 Unauthorized。
		response.ErrorWithStatus(c, http.StatusUnauthorized, "Login failed", err.Error())
		return
	}

	// 返回成功的响应，包含JWT令牌。
	response.SuccessWithStatus(c, http.StatusOK, "Login successful", gin.H{"token": token})
}

// GetProfile 处理获取当前登录管理员个人信息的HTTP请求。
// Method: GET
// Path: /admin/profile
func (h *Handler) GetProfile(c *gin.Context) {
	// 假设前置中间件已经将用户ID存储在Gin上下文中，键名为 "userID"。
	// 例如，通过JWT验证中间件获取并存储。
	userID, exists := c.Get("userID")
	if !exists {
		response.ErrorWithStatus(c, http.StatusUnauthorized, "Unauthorized", "User ID not found in context")
		return
	}

	// 调用应用服务层获取管理员详情。
	admin, err := h.service.GetAdminProfile(c.Request.Context(), userID.(uint64))
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get profile", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get profile", err.Error())
		return
	}

	// 返回成功的响应，包含管理员个人信息。
	response.SuccessWithStatus(c, http.StatusOK, "Profile retrieved successfully", admin)
}

// ListAdmins 处理获取管理员列表的HTTP请求。
// Method: GET
// Path: /admin/list
// 支持分页查询，通过query参数 "page" 和 "page_size" 控制。
func (h *Handler) ListAdmins(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取管理员列表。
	admins, total, err := h.service.ListAdmins(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list admins", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list admins", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Admins listed successfully", gin.H{
		"data":      admins,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateRole 处理创建角色的HTTP请求。
// Method: POST
// Path: /admin/role
func (h *Handler) CreateRole(c *gin.Context) {
	// 定义请求体结构，用于角色名称、编码和描述。
	var req struct {
		Name        string `json:"name" binding:"required"` // 角色名称，必填。
		Code        string `json:"code" binding:"required"` // 角色编码，必填。
		Description string `json:"description"`             // 角色描述，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建角色。
	role, err := h.service.CreateRole(c.Request.Context(), req.Name, req.Code, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create role", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create role", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Role created successfully", role)
}

// AssignRole 处理为管理员分配角色的HTTP请求。
// Method: POST
// Path: /admin/assign-role
func (h *Handler) AssignRole(c *gin.Context) {
	// 定义请求体结构，用于管理员ID和角色ID。
	var req struct {
		AdminID uint64 `json:"admin_id" binding:"required"` // 管理员ID，必填。
		RoleID  uint64 `json:"role_id" binding:"required"`  // 角色ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层为管理员分配角色。
	if err := h.service.AssignRoleToAdmin(c.Request.Context(), req.AdminID, req.RoleID); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to assign role", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to assign role", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Role assigned successfully", nil)
}

// --- Additional Handlers ---

// UpdateAdmin 处理更新管理员信息的HTTP请求。
// Method: PUT
// Path: /admin/users/:id
func (h *Handler) UpdateAdmin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Email    string   `json:"email"`
		RealName string   `json:"real_name"`
		Phone    string   `json:"phone"`
		RoleIDs  []uint64 `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	admin, err := h.service.UpdateAdmin(c.Request.Context(), id, req.Email, req.RealName, req.Phone, req.RoleIDs)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update admin", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Admin updated successfully", admin)
}

// DeleteAdmin 处理删除管理员的HTTP请求。
// Method: DELETE
// Path: /admin/users/:id
func (h *Handler) DeleteAdmin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.DeleteAdmin(c.Request.Context(), id); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete admin", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Admin deleted successfully", nil)
}

// GetRole 处理获取角色详情的HTTP请求。
// Method: GET
// Path: /admin/roles/:id
func (h *Handler) GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	role, err := h.service.GetRole(c.Request.Context(), id)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get role", err.Error())
		return
	}
	if role == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Role not found", "role not found")
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Role retrieved successfully", role)
}

// UpdateRole 处理更新角色的HTTP请求。
// Method: PUT
// Path: /admin/roles/:id
func (h *Handler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		PermissionIDs []uint64 `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	role, err := h.service.UpdateRole(c.Request.Context(), id, req.Name, req.Description, req.PermissionIDs)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update role", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Role updated successfully", role)
}

// DeleteRole 处理删除角色的HTTP请求。
// Method: DELETE
// Path: /admin/roles/:id
func (h *Handler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.DeleteRole(c.Request.Context(), id); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete role", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Role deleted successfully", nil)
}

// ListRoles 处理列出角色的HTTP请求。
// Method: GET
// Path: /admin/roles
func (h *Handler) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	roles, total, err := h.service.ListRoles(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list roles", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Roles listed successfully", gin.H{
		"data":      roles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreatePermission 处理创建权限的HTTP请求。
// Method: POST
// Path: /admin/permissions
func (h *Handler) CreatePermission(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Code     string `json:"code" binding:"required"`
		Type     string `json:"type" binding:"required"`
		Path     string `json:"path"`
		Method   string `json:"method"`
		ParentID uint64 `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	perm, err := h.service.CreatePermission(c.Request.Context(), req.Name, req.Code, req.Type, req.Path, req.Method, req.ParentID)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create permission", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Permission created successfully", perm)
}

// ListPermissions 处理列出权限的HTTP请求。
// Method: GET
// Path: /admin/permissions
func (h *Handler) ListPermissions(c *gin.Context) {
	perms, err := h.service.ListPermissions(c.Request.Context())
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list permissions", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Permissions listed successfully", perms)
}

// ListAuditLogs 处理列出审计日志的HTTP请求。
// Method: GET
// Path: /admin/audit-logs
func (h *Handler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	adminID, _ := strconv.ParseUint(c.Query("admin_id"), 10, 64)

	logs, total, err := h.service.ListAuditLogs(c.Request.Context(), adminID, page, pageSize)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list audit logs", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Audit logs listed successfully", gin.H{
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetSystemSetting 处理获取系统设置的HTTP请求。
// Method: GET
// Path: /admin/settings/:key
func (h *Handler) GetSystemSetting(c *gin.Context) {
	key := c.Param("key")
	setting, err := h.service.GetSystemSetting(c.Request.Context(), key)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get setting", err.Error())
		return
	}
	if setting == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Setting not found", "setting not found")
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Setting retrieved successfully", setting)
}

// UpdateSystemSetting 处理更新系统设置的HTTP请求。
// Method: PUT
// Path: /admin/settings/:key
func (h *Handler) UpdateSystemSetting(c *gin.Context) {
	key := c.Param("key")
	var req struct {
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	setting, err := h.service.UpdateSystemSetting(c.Request.Context(), key, req.Value, req.Description)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update setting", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Setting updated successfully", setting)
}

// RegisterRoutes 在给定的Gin路由组中注册Admin模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /auth 路由组，用于认证相关接口。
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", h.RegisterAdmin) // 注册管理员。
		authGroup.POST("/login", h.Login)            // 管理员登录。
	}

	// /admin 路由组，用于管理员管理接口。
	adminGroup := r.Group("/admin")
	// TODO: 在这里添加认证中间件（例如JWTAuth），保护/admin组的接口。
	{
		adminGroup.GET("/profile", h.GetProfile)       // 获取当前登录管理员的个人信息。
		adminGroup.GET("/list", h.ListAdmins)          // 获取管理员列表。
		adminGroup.PUT("/users/:id", h.UpdateAdmin)    // 更新管理员信息。
		adminGroup.DELETE("/users/:id", h.DeleteAdmin) // 删除管理员。

		adminGroup.POST("/role", h.CreateRole)        // 创建角色。
		adminGroup.GET("/roles", h.ListRoles)         // 获取角色列表。
		adminGroup.GET("/roles/:id", h.GetRole)       // 获取角色详情。
		adminGroup.PUT("/roles/:id", h.UpdateRole)    // 更新角色。
		adminGroup.DELETE("/roles/:id", h.DeleteRole) // 删除角色。
		adminGroup.POST("/assign-role", h.AssignRole) // 分配角色给管理员。

		adminGroup.POST("/permissions", h.CreatePermission) // 创建权限。
		adminGroup.GET("/permissions", h.ListPermissions)   // 获取权限列表。

		adminGroup.GET("/audit-logs", h.ListAuditLogs) // 获取审计日志。

		adminGroup.GET("/settings/:key", h.GetSystemSetting)    // 获取系统设置。
		adminGroup.PUT("/settings/:key", h.UpdateSystemSetting) // 更新系统设置。
	}
}

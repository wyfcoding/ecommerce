package application

import (
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"github.com/wyfcoding/pkg/storage"
	"google.golang.org/grpc"
)

// AdminService 门面服务，聚合 Command (Manager) 和 Query (Query)
type AdminService struct {
	Manager *AdminManager
	Query   *AdminQuery
}

// NewAdminService 创建一个新的管理后台应用服务实例。
func NewAdminService(
	userRepo domain.AdminRepository,
	roleRepo domain.RoleRepository,
	auditRepo domain.AuditRepository,
	settingRepo domain.SettingRepository,
	approvalRepo domain.ApprovalRepository,
	opsDeps SystemOpsDependencies,
	logger *slog.Logger,
) *AdminService {
	return &AdminService{
		Manager: NewAdminManager(userRepo, roleRepo, auditRepo, settingRepo, approvalRepo, opsDeps, logger),
		Query:   NewAdminQuery(userRepo, roleRepo, auditRepo, settingRepo, approvalRepo),
	}
}

// SystemOpsDependencies 系统操作依赖的其他服务客户端与基础设施
type SystemOpsDependencies struct {
	OrderClient   *grpc.ClientConn
	PaymentClient *grpc.ClientConn
	UserClient    *grpc.ClientConn
	Storage       storage.Storage // 【优化】：纳入统一依赖管理
}

// --- DTO Definitions ---

// LoginRequest 定义了请求参数结构。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 定义了响应数据结构。
type LoginResponse struct {
	Token    string   `json:"token"`
	UserInfo UserInfo `json:"userInfo"`
}

// UserInfo 结构体定义。
type UserInfo struct {
	ID          uint     `json:"id"`
	Username    string   `json:"username"`
	FullName    string   `json:"fullName"`
	Roles       []string `json:"roles"` // Role Codes
	Permissions []string `json:"permissions"`
}

// CreateUserRequest 定义了请求参数结构。
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"fullName"`
	Roles    []uint `json:"roles"` // 角色 ID 列表
}

// ApprovalCreateRequest 定义了请求参数结构。
type ApprovalCreateRequest struct {
	ActionType  string `json:"actionType" binding:"required"`
	Description string `json:"description"`
	Payload     string `json:"payload" binding:"required"` // JSON 字符串
}

// ApprovalActionRequest 定义了请求参数结构。
type ApprovalActionRequest struct {
	Action  string `json:"action" binding:"required,oneof=approve reject"`
	Comment string `json:"comment"`
}

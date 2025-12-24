package grpc

import (
	"context"
	"fmt" // 用于格式化错误信息。

	pb "github.com/wyfcoding/ecommerce/goapi/admin/v1"          // 导入Admin模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/admin/application" // 导入Admin模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/admin/domain"

	// 导入Admin模块的领域实体。
	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 AdminService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedAdminServiceServer                           // 嵌入生成的UnimplementedAdminServiceServer，确保前向兼容性。
	app                                *application.AdminService // 依赖Admin应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Admin gRPC 服务端实例。
func NewServer(app *application.AdminService) *Server {
	return &Server{app: app}
}

// --- 模块分段 ---

// CreateAdminUser 处理创建管理员用户的gRPC请求。
// req: 包含创建管理员所需信息的请求体。
// 返回created successfully的管理员用户响应和可能发生的gRPC错误。
func (s *Server) CreateAdminUser(ctx context.Context, req *pb.CreateAdminUserRequest) (*pb.CreateAdminUserResponse, error) {
	// 调用应用服务层注册管理员用户。
	admin, err := s.app.RegisterAdmin(ctx, req.Username, req.Email, req.Password, req.Nickname, "")
	if err != nil {
		// 简单错误处理，生产环境应使用 status.Error 包装具体错误
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create admin user: %v", err))
	}

	// 如果请求中包含角色ID，则为新创建的管理员分配角色。
	if len(req.RoleIds) > 0 {
		for _, roleID := range req.RoleIds {
			if err := s.app.AssignRoleToAdmin(ctx, uint64(admin.ID), roleID); err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("failed to assign role to admin: %v", err))
			}
		}
		// 刷新数据
		if a, fetchErr := s.app.GetAdminProfile(ctx, uint64(admin.ID)); fetchErr == nil {
			admin = a
		}
	}

	return &pb.CreateAdminUserResponse{
		AdminUser: s.adminToProto(admin),
	}, nil
}

// GetAdminUser 处理获取单个管理员用户信息的gRPC请求。
// req: 包含管理员用户ID的请求体。
// 返回管理员用户响应和可能发生的gRPC错误。
func (s *Server) GetAdminUser(ctx context.Context, req *pb.GetAdminUserRequest) (*pb.GetAdminUserResponse, error) {
	admin, err := s.app.GetAdminProfile(ctx, req.Id)
	if err != nil {
		// 如果管理员未找到，返回NotFound状态码。
		return nil, status.Error(codes.NotFound, fmt.Sprintf("admin user not found: %v", err))
	}
	return &pb.GetAdminUserResponse{
		AdminUser: s.adminToProto(admin),
	}, nil
}

// UpdateAdminUser 处理更新管理员用户信息的gRPC请求。
// 此方法尚未实现。
// UpdateAdminUser 处理更新管理员用户信息的gRPC请求。
func (s *Server) UpdateAdminUser(ctx context.Context, req *pb.UpdateAdminUserRequest) (*pb.UpdateAdminUserResponse, error) {
	admin, err := s.app.UpdateAdmin(ctx, req.Id, req.Email, req.Nickname, "", req.RoleIds)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update admin user: %v", err))
	}
	return &pb.UpdateAdminUserResponse{
		AdminUser: s.adminToProto(admin),
	}, nil
}

// DeleteAdminUser 处理删除管理员用户的gRPC请求。
func (s *Server) DeleteAdminUser(ctx context.Context, req *pb.DeleteAdminUserRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteAdmin(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete admin user: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListAdminUsers 处理列出管理员用户的gRPC请求，支持分页。
// req: 包含分页参数的请求体。
// 返回管理员用户列表响应和可能发生的gRPC错误。
func (s *Server) ListAdminUsers(ctx context.Context, req *pb.ListAdminUsersRequest) (*pb.ListAdminUsersResponse, error) {
	// 将protobuf的分页Token转换为应用服务层的页码。
	page := max(int(req.PageToken), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取管理员列表。
	admins, total, err := s.app.ListAdmins(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list admin users: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbAdmins := make([]*pb.AdminUser, len(admins))
	for i, a := range admins {
		pbAdmins[i] = s.adminToProto(a)
	}

	return &pb.ListAdminUsersResponse{
		AdminUsers: pbAdmins,
		TotalCount: int32(total), // 总记录数。
	}, nil
}

// --- 角色 ---

// CreateRole 处理创建角色的gRPC请求。
// req: 包含创建角色所需信息的请求体。
// 返回创建成功的角色响应和可能发生的gRPC错误。
func (s *Server) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	// 调用应用服务层创建角色。
	// 注意：Proto中没有Code字段，这里暂时使用Name作为Code。
	role, err := s.app.CreateRole(ctx, req.Name, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create role: %v", err))
	}

	// 如果请求中包含权限ID，则为新创建的角色分配权限。
	if len(req.PermissionIds) > 0 {
		for _, permID := range req.PermissionIds {
			if err := s.app.AssignPermissionToRole(ctx, uint64(role.ID), permID); err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("failed to assign permission to role: %v", err))
			}
		}
	}

	return &pb.CreateRoleResponse{
		Role: s.roleToProto(role),
	}, nil
}

// GetRole 处理获取单个角色信息的gRPC请求。
func (s *Server) GetRole(ctx context.Context, req *pb.GetRoleRequest) (*pb.GetRoleResponse, error) {
	role, err := s.app.GetRole(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get role: %v", err))
	}
	if role == nil {
		return nil, status.Error(codes.NotFound, "role not found")
	}
	return &pb.GetRoleResponse{
		Role: s.roleToProto(role),
	}, nil
}

// UpdateRole 处理更新角色信息的gRPC请求。
func (s *Server) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	role, err := s.app.UpdateRole(ctx, req.Id, req.Name, req.Description, req.PermissionIds)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update role: %v", err))
	}
	return &pb.UpdateRoleResponse{
		Role: s.roleToProto(role),
	}, nil
}

// DeleteRole 处理删除角色的gRPC请求。
func (s *Server) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteRole(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete role: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListRoles 处理列出角色列表的gRPC请求，支持分页。
// req: 包含分页参数的请求体。
// 返回角色列表响应和可能发生的gRPC错误。
func (s *Server) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	// 将protobuf的分页Token转换为应用服务层的页码。
	page := max(int(req.PageToken), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取角色列表。
	roles, total, err := s.app.ListRoles(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list roles: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbRoles := make([]*pb.Role, len(roles))
	for i, r := range roles {
		pbRoles[i] = s.roleToProto(r)
	}

	return &pb.ListRolesResponse{
		Roles:      pbRoles,
		TotalCount: int32(total), // 总记录数。
	}, nil
}

// --- 权限 ---

// CreatePermission 处理创建权限的gRPC请求。
// req: 包含创建权限所需信息的请求体。
// 返回创建成功的权限响应和可能发生的gRPC错误。
func (s *Server) CreatePermission(ctx context.Context, req *pb.CreatePermissionRequest) (*pb.CreatePermissionResponse, error) {
	// 调用应用服务层创建权限。
	// 注意：Proto定义与实体定义之间存在字段缺失，这里使用占位符或默认值填充。
	perm, err := s.app.CreatePermission(ctx, req.Name, req.Name, "api", "", "", 0)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create permission: %v", err))
	}
	return &pb.CreatePermissionResponse{
		Permission: s.permissionToProto(perm),
	}, nil
}

// GetPermission 处理获取单个权限信息的gRPC请求。
func (s *Server) GetPermission(ctx context.Context, req *pb.GetPermissionRequest) (*pb.GetPermissionResponse, error) {
	perm, err := s.app.GetPermission(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get permission: %v", err))
	}
	if perm == nil {
		return nil, status.Error(codes.NotFound, "permission not found")
	}
	return &pb.GetPermissionResponse{
		Permission: s.permissionToProto(perm),
	}, nil
}

// ListPermissions 处理列出权限列表的gRPC请求。
// req: 空的请求体。
// 返回权限列表响应和可能发生的gRPC错误。
func (s *Server) ListPermissions(ctx context.Context, req *pb.ListPermissionsRequest) (*pb.ListPermissionsResponse, error) {
	perms, err := s.app.ListPermissions(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list permissions: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbPerms := make([]*pb.Permission, len(perms))
	for i, p := range perms {
		pbPerms[i] = s.permissionToProto(p)
	}

	return &pb.ListPermissionsResponse{
		Permissions: pbPerms,
		TotalCount:  int32(len(perms)), // 总记录数。
	}, nil
}

// --- 审计日志 ---

// ListAuditLogs 处理列出审计日志的gRPC请求。
func (s *Server) ListAuditLogs(ctx context.Context, req *pb.ListAuditLogsRequest) (*pb.ListAuditLogsResponse, error) {
	page := max(int(req.PageToken), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	logs, total, err := s.app.ListAuditLogs(ctx, req.AdminUserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list audit logs: %v", err))
	}

	pbLogs := make([]*pb.AuditLog, len(logs))
	for i, l := range logs {
		pbLogs[i] = &pb.AuditLog{
			Id:          uint64(l.ID),
			AdminUserId: uint64(l.UserID), // 领域层 UserID 类型为 uint
			Action:      l.Action,
			EntityType:  l.Resource, // 将 Resource 映射到 EntityType
			Details:     l.Payload,  // 将 Payload 映射到 Details
			IpAddress:   l.ClientIP,
			CreatedAt:   timestamppb.New(l.CreatedAt),
		}
	}

	return &pb.ListAuditLogsResponse{
		AuditLogs:  pbLogs,
		TotalCount: int32(total),
	}, nil
}

// --- 系统设置 ---

// GetSystemSetting 处理获取系统设置的gRPC请求。
func (s *Server) GetSystemSetting(ctx context.Context, req *pb.GetSystemSettingRequest) (*pb.GetSystemSettingResponse, error) {
	setting, err := s.app.GetSystemSetting(ctx, req.Key)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get system setting: %v", err))
	}
	if setting == nil {
		return nil, status.Error(codes.NotFound, "setting not found")
	}
	return &pb.GetSystemSettingResponse{
		Setting: &pb.SystemSetting{
			Key:         setting.Key,
			Value:       setting.Value,
			Description: setting.Description,
			UpdatedAt:   timestamppb.New(setting.UpdatedAt),
		},
	}, nil
}

// UpdateSystemSetting 处理更新系统设置的gRPC请求。
func (s *Server) UpdateSystemSetting(ctx context.Context, req *pb.UpdateSystemSettingRequest) (*pb.UpdateSystemSettingResponse, error) {
	setting, err := s.app.UpdateSystemSetting(ctx, req.Key, req.Value, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update system setting: %v", err))
	}
	return &pb.UpdateSystemSettingResponse{
		Setting: &pb.SystemSetting{
			Key:         setting.Key,
			Value:       setting.Value,
			Description: setting.Description,
			UpdatedAt:   timestamppb.New(setting.UpdatedAt),
		},
	}, nil
}

// --- 辅助函数 ---

// adminToProto 将领域层的 Admin 实体转换为 protobuf 的 AdminUser 消息。
func (s *Server) adminToProto(a *domain.AdminUser) *pb.AdminUser {
	// 提取管理员的角色名称列表。
	roles := make([]string, len(a.Roles))
	for i, r := range a.Roles {
		roles[i] = r.Name
	}
	// Proto 期望 IsActive 为 bool, 实体 Status 为 int。进行映射。
	isActive := a.Status == domain.UserStatusActive

	return &pb.AdminUser{
		Id:        uint64(a.ID),                 // 管理员ID。
		Username:  a.Username,                   // 用户名。
		Email:     a.Email,                      // 邮箱。
		Nickname:  a.FullName,                   // 昵称（映射为全名）。
		IsActive:  isActive,                     // 是否激活。
		Roles:     roles,                        // 角色名称列表。
		CreatedAt: timestamppb.New(a.CreatedAt), // 创建时间。
		UpdatedAt: timestamppb.New(a.UpdatedAt), // 更新时间。
	}
}

// roleToProto 将领域层的 Role 实体转换为 protobuf 的 Role 消息。
func (s *Server) roleToProto(r *domain.Role) *pb.Role {
	// 提取角色拥有的权限名称列表。
	perms := make([]string, len(r.Permissions))
	for i, p := range r.Permissions {
		perms[i] = p.Name // Proto中期望字符串权限（可能是Code或Name），这里使用Name。
	}
	return &pb.Role{
		Id:          uint64(r.ID),                 // 角色ID。
		Name:        r.Name,                       // 角色名称。
		Description: r.Description,                // 角色描述。
		Permissions: perms,                        // 权限名称列表。
		CreatedAt:   timestamppb.New(r.CreatedAt), // 创建时间。
		UpdatedAt:   timestamppb.New(r.UpdatedAt), // 更新时间。
	}
}

// permissionToProto 将领域层的 Permission 实体转换为 protobuf 的 Permission 消息。
func (s *Server) permissionToProto(p *domain.Permission) *pb.Permission {
	return &pb.Permission{
		Id:          uint64(p.ID), // 权限ID。
		Name:        p.Name,       // 权限名称。
		Description: p.Description,
	}
}

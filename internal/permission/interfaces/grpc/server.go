package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/permission/v1"              // 导入权限模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/permission/application"   // 导入权限模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/permission/domain/entity" // 导入权限模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 PermissionService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedPermissionServiceServer                                // 嵌入生成的UnimplementedPermissionServiceServer，确保前向兼容性。
	app                                     *application.PermissionService // 依赖Permission应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Permission gRPC 服务端实例。
func NewServer(app *application.PermissionService) *Server {
	return &Server{app: app}
}

// CreateRole 处理创建角色的gRPC请求。
// req: 包含角色名称、描述和权限ID列表的请求体。
// 返回创建成功的角色响应和可能发生的gRPC错误。
func (s *Server) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.Role, error) {
	role, err := s.app.CreateRole(ctx, req.Name, req.Description, req.PermissionIds)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create role: %v", err))
	}
	// 将领域实体转换为protobuf响应格式。
	return convertRoleToProto(role), nil
}

// GetRole 处理获取角色详情的gRPC请求。
// req: 包含角色ID的请求体。
// 返回角色响应和可能发生的gRPC错误。
func (s *Server) GetRole(ctx context.Context, req *pb.GetRoleRequest) (*pb.Role, error) {
	role, err := s.app.GetRole(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("role not found: %v", err))
	}
	return convertRoleToProto(role), nil
}

// ListRoles 处理列出角色的gRPC请求。
// req: 包含分页参数的请求体。
// 返回角色列表响应和可能发生的gRPC错误。
func (s *Server) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	// 获取分页参数。
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
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
		pbRoles[i] = convertRoleToProto(r)
	}

	return &pb.ListRolesResponse{
		Roles:      pbRoles,
		TotalCount: total, // 总记录数。
	}, nil
}

// DeleteRole 处理删除角色的gRPC请求。
// req: 包含角色ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteRole(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete role: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// CreatePermission 处理创建权限的gRPC请求。
// req: 包含权限代码和描述的请求体。
// 返回创建成功的权限响应和可能发生的gRPC错误。
func (s *Server) CreatePermission(ctx context.Context, req *pb.CreatePermissionRequest) (*pb.Permission, error) {
	perm, err := s.app.CreatePermission(ctx, req.Code, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create permission: %v", err))
	}
	// 将领域实体转换为protobuf响应格式。
	return convertPermissionToProto(perm), nil
}

// ListPermissions 处理列出权限的gRPC请求。
// req: 包含分页参数的请求体。
// 返回权限列表响应和可能发生的gRPC错误。
func (s *Server) ListPermissions(ctx context.Context, req *pb.ListPermissionsRequest) (*pb.ListPermissionsResponse, error) {
	// 获取分页参数。
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取权限列表。
	perms, total, err := s.app.ListPermissions(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list permissions: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbPerms := make([]*pb.Permission, len(perms))
	for i, p := range perms {
		pbPerms[i] = convertPermissionToProto(p)
	}

	return &pb.ListPermissionsResponse{
		Permissions: pbPerms,
		TotalCount:  total, // 总记录数。
	}, nil
}

// AssignRole 处理为用户分配角色的gRPC请求。
// req: 包含用户ID和角色ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) AssignRole(ctx context.Context, req *pb.AssignRoleRequest) (*emptypb.Empty, error) {
	if err := s.app.AssignRole(ctx, req.UserId, req.RoleId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to assign role: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// RevokeRole 处理撤销用户角色的gRPC请求。
// req: 包含用户ID和角色ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) RevokeRole(ctx context.Context, req *pb.RevokeRoleRequest) (*emptypb.Empty, error) {
	if err := s.app.RevokeRole(ctx, req.UserId, req.RoleId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to revoke role: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// GetUserRoles 处理获取用户所有角色的gRPC请求。
// req: 包含用户ID的请求体。
// 返回角色列表响应和可能发生的gRPC错误。
func (s *Server) GetUserRoles(ctx context.Context, req *pb.GetUserRolesRequest) (*pb.GetUserRolesResponse, error) {
	roles, err := s.app.GetUserRoles(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get user roles: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbRoles := make([]*pb.Role, len(roles))
	for i, r := range roles {
		pbRoles[i] = convertRoleToProto(r)
	}

	return &pb.GetUserRolesResponse{
		Roles: pbRoles,
	}, nil
}

// CheckPermission 处理检查用户是否拥有特定权限的gRPC请求。
// req: 包含用户ID和权限代码的请求体。
// 返回权限检查结果响应和可能发生的gRPC错误。
func (s *Server) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	allowed, err := s.app.CheckPermission(ctx, req.UserId, req.PermissionCode)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check permission: %v", err))
	}
	return &pb.CheckPermissionResponse{
		Allowed: allowed, // 是否允许。
	}, nil
}

// convertRoleToProto 是一个辅助函数，将领域层的 Role 实体转换为 protobuf 的 Role 消息。
func convertRoleToProto(r *entity.Role) *pb.Role {
	if r == nil {
		return nil
	}
	// 转换关联的 Permissions。
	pbPerms := make([]*pb.Permission, len(r.Permissions))
	for i, p := range r.Permissions {
		pbPerms[i] = convertPermissionToProto(p)
	}
	return &pb.Role{
		Id:          uint64(r.ID),                 // 角色ID。
		Name:        r.Name,                       // 名称。
		Description: r.Description,                // 描述。
		Permissions: pbPerms,                      // 权限列表。
		CreatedAt:   timestamppb.New(r.CreatedAt), // 创建时间。
		UpdatedAt:   timestamppb.New(r.UpdatedAt), // 更新时间。
	}
}

// convertPermissionToProto 是一个辅助函数，将领域层的 Permission 实体转换为 protobuf 的 Permission 消息。
func convertPermissionToProto(p *entity.Permission) *pb.Permission {
	if p == nil {
		return nil
	}
	return &pb.Permission{
		Id:          uint64(p.ID),                 // 权限ID。
		Code:        p.Code,                       // 代码。
		Description: p.Description,                // 描述。
		CreatedAt:   timestamppb.New(p.CreatedAt), // 创建时间。
		UpdatedAt:   timestamppb.New(p.UpdatedAt), // 更新时间。
	}
}

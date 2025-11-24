package grpc

import (
	"context"
	pb "ecommerce/api/admin/v1"
	"ecommerce/internal/admin/application"
	"ecommerce/internal/admin/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedAdminServiceServer
	app *application.AdminService
}

func NewServer(app *application.AdminService) *Server {
	return &Server{app: app}
}

// --- AdminUser ---

func (s *Server) CreateAdminUser(ctx context.Context, req *pb.CreateAdminUserRequest) (*pb.CreateAdminUserResponse, error) {
	// Service RegisterAdmin(ctx, username, email, password, realName, phone)
	// Proto: username, password, email, nickname, role_ids
	// Mapping nickname -> realName. Phone is missing in proto, passing empty.

	admin, err := s.app.RegisterAdmin(ctx, req.Username, req.Email, req.Password, req.Nickname, "")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Assign roles if provided
	if len(req.RoleIds) > 0 {
		for _, roleID := range req.RoleIds {
			if err := s.app.AssignRoleToAdmin(ctx, uint64(admin.ID), roleID); err != nil {
				// Log error but continue? Or fail?
				// For now, fail to ensure consistency
				return nil, status.Error(codes.Internal, "failed to assign role: "+err.Error())
			}
		}
		// Re-fetch admin to get roles? Or just return what we have (roles empty in returned admin from Register)
		// Service RegisterAdmin returns admin without preloading roles usually.
		// We might need to fetch it again if we want to return roles in response.
		// Let's fetch it.
		if a, err := s.app.GetAdminProfile(ctx, uint64(admin.ID)); err == nil {
			admin = a
		}
	}

	return &pb.CreateAdminUserResponse{
		AdminUser: s.adminToProto(admin),
	}, nil
}

func (s *Server) GetAdminUser(ctx context.Context, req *pb.GetAdminUserRequest) (*pb.GetAdminUserResponse, error) {
	admin, err := s.app.GetAdminProfile(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetAdminUserResponse{
		AdminUser: s.adminToProto(admin),
	}, nil
}

func (s *Server) UpdateAdminUser(ctx context.Context, req *pb.UpdateAdminUserRequest) (*pb.UpdateAdminUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UpdateAdminUser not implemented")
}

func (s *Server) DeleteAdminUser(ctx context.Context, req *pb.DeleteAdminUserRequest) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "DeleteAdminUser not implemented")
}

func (s *Server) ListAdminUsers(ctx context.Context, req *pb.ListAdminUsersRequest) (*pb.ListAdminUsersResponse, error) {
	// Service ListAdmins(ctx, page, pageSize)
	// Proto has page_token (int32) as page?
	page := int(req.PageToken)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	admins, total, err := s.app.ListAdmins(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbAdmins := make([]*pb.AdminUser, len(admins))
	for i, a := range admins {
		pbAdmins[i] = s.adminToProto(a)
	}

	return &pb.ListAdminUsersResponse{
		AdminUsers: pbAdmins,
		TotalCount: int32(total),
	}, nil
}

// --- Role ---

func (s *Server) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	// Service CreateRole(ctx, name, code, description)
	// Proto: name, description, permission_ids. Missing code.
	// We'll use name as code or generate one.
	// Let's use name as code for now.
	role, err := s.app.CreateRole(ctx, req.Name, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Assign permissions
	if len(req.PermissionIds) > 0 {
		for _, permID := range req.PermissionIds {
			if err := s.app.AssignPermissionToRole(ctx, uint64(role.ID), permID); err != nil {
				return nil, status.Error(codes.Internal, "failed to assign permission: "+err.Error())
			}
		}
	}

	return &pb.CreateRoleResponse{
		Role: s.roleToProto(role),
	}, nil
}

func (s *Server) GetRole(ctx context.Context, req *pb.GetRoleRequest) (*pb.GetRoleResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetRole not implemented")
}

func (s *Server) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UpdateRole not implemented")
}

func (s *Server) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "DeleteRole not implemented")
}

func (s *Server) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	page := int(req.PageToken)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	roles, total, err := s.app.ListRoles(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbRoles := make([]*pb.Role, len(roles))
	for i, r := range roles {
		pbRoles[i] = s.roleToProto(r)
	}

	return &pb.ListRolesResponse{
		Roles:      pbRoles,
		TotalCount: int32(total),
	}, nil
}

// --- Permission ---

func (s *Server) CreatePermission(ctx context.Context, req *pb.CreatePermissionRequest) (*pb.CreatePermissionResponse, error) {
	// Service CreatePermission(ctx, name, code, permType, path, method, parentID)
	// Proto: name, description. Missing code, type, path, method, parentID.
	// This is a big gap.
	// We'll use placeholders or defaults.
	// Code = name, Type = "api", Path = "", Method = "", ParentID = 0
	perm, err := s.app.CreatePermission(ctx, req.Name, req.Name, "api", "", "", 0)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreatePermissionResponse{
		Permission: s.permissionToProto(perm),
	}, nil
}

func (s *Server) GetPermission(ctx context.Context, req *pb.GetPermissionRequest) (*pb.GetPermissionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetPermission not implemented")
}

func (s *Server) ListPermissions(ctx context.Context, req *pb.ListPermissionsRequest) (*pb.ListPermissionsResponse, error) {
	perms, err := s.app.ListPermissions(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbPerms := make([]*pb.Permission, len(perms))
	for i, p := range perms {
		pbPerms[i] = s.permissionToProto(p)
	}

	return &pb.ListPermissionsResponse{
		Permissions: pbPerms,
		TotalCount:  int32(len(perms)),
	}, nil
}

// --- AuditLog ---

func (s *Server) ListAuditLogs(ctx context.Context, req *pb.ListAuditLogsRequest) (*pb.ListAuditLogsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListAuditLogs not implemented")
}

// --- SystemSetting ---

func (s *Server) GetSystemSetting(ctx context.Context, req *pb.GetSystemSettingRequest) (*pb.GetSystemSettingResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetSystemSetting not implemented")
}

func (s *Server) UpdateSystemSetting(ctx context.Context, req *pb.UpdateSystemSettingRequest) (*pb.UpdateSystemSettingResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UpdateSystemSetting not implemented")
}

// --- Helpers ---

func (s *Server) adminToProto(a *entity.Admin) *pb.AdminUser {
	roles := make([]string, len(a.Roles))
	for i, r := range a.Roles {
		roles[i] = r.Name
	}
	return &pb.AdminUser{
		Id:        uint64(a.ID),
		Username:  a.Username,
		Email:     a.Email,
		Nickname:  a.RealName,
		IsActive:  a.IsActive(),
		Roles:     roles,
		CreatedAt: timestamppb.New(a.CreatedAt),
		UpdatedAt: timestamppb.New(a.UpdatedAt),
	}
}

func (s *Server) roleToProto(r *entity.Role) *pb.Role {
	perms := make([]string, len(r.Permissions))
	for i, p := range r.Permissions {
		perms[i] = p.Name // Proto expects string permissions, maybe codes?
	}
	return &pb.Role{
		Id:          uint64(r.ID),
		Name:        r.Name,
		Description: r.Description,
		Permissions: perms,
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
	}
}

func (s *Server) permissionToProto(p *entity.Permission) *pb.Permission {
	return &pb.Permission{
		Id:          uint64(p.ID),
		Name:        p.Name,
		Description: "", // Entity doesn't have description? Check entity.
		// Entity Permission: Name, Code, Type, ParentID, Path, Method, Icon, Sort, Status.
		// No Description.
	}
}

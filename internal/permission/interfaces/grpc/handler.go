package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/permission/v1"
	"github.com/wyfcoding/ecommerce/internal/permission/application"
	"github.com/wyfcoding/ecommerce/internal/permission/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体定义。
type Server struct {
	pb.UnimplementedPermissionServiceServer
	app *application.PermissionService
}

// NewServer 函数。
func NewServer(app *application.PermissionService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.Role, error) {
	role, err := s.app.CreateRole(ctx, req.Name, req.Description, req.PermissionIds)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create role: %v", err))
	}
	return convertRoleToProto(role), nil
}

func (s *Server) GetRole(ctx context.Context, req *pb.GetRoleRequest) (*pb.Role, error) {
	role, err := s.app.GetRole(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("role not found: %v", err))
	}
	return convertRoleToProto(role), nil
}

func (s *Server) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	roles, total, err := s.app.ListRoles(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list roles: %v", err))
	}

	pbRoles := make([]*pb.Role, len(roles))
	for i, r := range roles {
		pbRoles[i] = convertRoleToProto(r)
	}

	return &pb.ListRolesResponse{
		Roles:      pbRoles,
		TotalCount: total,
	}, nil
}

func (s *Server) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteRole(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete role: %v", err))
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) CreatePermission(ctx context.Context, req *pb.CreatePermissionRequest) (*pb.Permission, error) {
	perm, err := s.app.CreatePermission(ctx, req.Code, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create permission: %v", err))
	}
	return convertPermissionToProto(perm), nil
}

func (s *Server) ListPermissions(ctx context.Context, req *pb.ListPermissionsRequest) (*pb.ListPermissionsResponse, error) {
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	perms, total, err := s.app.ListPermissions(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list permissions: %v", err))
	}

	pbPerms := make([]*pb.Permission, len(perms))
	for i, p := range perms {
		pbPerms[i] = convertPermissionToProto(p)
	}

	return &pb.ListPermissionsResponse{
		Permissions: pbPerms,
		TotalCount:  total,
	}, nil
}

func (s *Server) AssignRole(ctx context.Context, req *pb.AssignRoleRequest) (*emptypb.Empty, error) {
	if err := s.app.AssignRole(ctx, req.UserId, req.RoleId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to assign role: %v", err))
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) RevokeRole(ctx context.Context, req *pb.RevokeRoleRequest) (*emptypb.Empty, error) {
	if err := s.app.RevokeRole(ctx, req.UserId, req.RoleId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to revoke role: %v", err))
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetUserRoles(ctx context.Context, req *pb.GetUserRolesRequest) (*pb.GetUserRolesResponse, error) {
	roles, err := s.app.GetUserRoles(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get user roles: %v", err))
	}

	pbRoles := make([]*pb.Role, len(roles))
	for i, r := range roles {
		pbRoles[i] = convertRoleToProto(r)
	}

	return &pb.GetUserRolesResponse{
		Roles: pbRoles,
	}, nil
}

func (s *Server) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	allowed, err := s.app.CheckPermission(ctx, req.UserId, req.PermissionCode)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check permission: %v", err))
	}
	return &pb.CheckPermissionResponse{
		Allowed: allowed,
	}, nil
}

func convertRoleToProto(r *domain.Role) *pb.Role {
	if r == nil {
		return nil
	}
	pbPerms := make([]*pb.Permission, len(r.Permissions))
	for i, p := range r.Permissions {
		pbPerms[i] = convertPermissionToProto(p)
	}
	return &pb.Role{
		Id:          uint64(r.ID),
		Name:        r.Name,
		Description: r.Description,
		Permissions: pbPerms,
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
	}
}

func convertPermissionToProto(p *domain.Permission) *pb.Permission {
	if p == nil {
		return nil
	}
	return &pb.Permission{
		Id:          uint64(p.ID),
		Code:        p.Code,
		Description: p.Description,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

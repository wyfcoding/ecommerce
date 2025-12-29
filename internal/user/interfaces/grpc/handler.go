package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/user/v1"
	"github.com/wyfcoding/ecommerce/internal/user/application"
	"github.com/wyfcoding/ecommerce/internal/user/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 User gRPC 服务。
type Server struct {
	pb.UnimplementedUserServer
	app *application.UserService
}

// NewServer 创建并返回一个新的 User gRPC 服务端实例。
func NewServer(app *application.UserService) *Server {
	return &Server{app: app}
}

// RegisterByPassword 处理用户通过密码注册的gRPC请求。
func (s *Server) RegisterByPassword(ctx context.Context, req *pb.RegisterByPasswordRequest) (*pb.RegisterResponse, error) {
	// 构造 DTO
	createReq := &application.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Username, // 临时假设
		Phone:    "",
	}

	user, err := s.app.Manager.Register(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to register user: %v", err))
	}
	return &pb.RegisterResponse{UserId: uint64(user.ID)}, nil
}

// LoginByPassword 处理用户通过密码登录的gRPC请求。
func (s *Server) LoginByPassword(ctx context.Context, req *pb.LoginByPasswordRequest) (*pb.LoginByPasswordResponse, error) {
	token, expiresAt, err := s.app.Manager.Login(ctx, req.Username, req.Password, "127.0.0.1")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to login: %v", err))
	}
	return &pb.LoginByPasswordResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// GetUserByID 处理根据用户ID获取用户信息的gRPC请求。
func (s *Server) GetUserByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.UserResponse, error) {
	user, err := s.app.Query.GetUser(ctx, uint(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get user: %v", err))
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &pb.UserResponse{User: convertUserToProto(user)}, nil
}

// UpdateUserInfo 处理更新用户信息的gRPC请求。
func (s *Server) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UserResponse, error) {
	var birthday string
	if req.Birthday != nil {
		t := req.Birthday.AsTime()
		birthday = t.Format("2006-01-02")
	}

	nickname := ""
	if req.Nickname != nil {
		nickname = *req.Nickname
	}
	avatar := ""
	if req.Avatar != nil {
		avatar = *req.Avatar
	}
	gender := int8(-1)
	if req.Gender != nil {
		gender = int8(*req.Gender)
	}

	updateReq := &application.UpdateProfileRequest{
		Nickname: nickname,
		Avatar:   avatar,
		Gender:   gender,
		Birthday: birthday,
	}

	user, err := s.app.Manager.UpdateProfile(ctx, uint(req.UserId), updateReq)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update user info: %v", err))
	}
	return &pb.UserResponse{User: convertUserToProto(user)}, nil
}

// AddAddress 处理添加用户地址的gRPC请求。
func (s *Server) AddAddress(ctx context.Context, req *pb.AddAddressRequest) (*pb.Address, error) {
	isDefault := false
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}

	addrDTO := &application.AddressDTO{
		RecipientName:   req.Name,
		PhoneNumber:     req.Phone,
		Province:        req.Province,
		City:            req.City,
		District:        req.District,
		DetailedAddress: req.DetailedAddress,
		IsDefault:       isDefault,
	}

	addr, err := s.app.Manager.AddAddress(ctx, uint(req.UserId), addrDTO)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add address: %v", err))
	}
	return convertAddressToProto(addr), nil
}

// GetAddress 处理获取用户地址的gRPC请求。
func (s *Server) GetAddress(ctx context.Context, req *pb.GetAddressRequest) (*pb.Address, error) {
	addr, err := s.app.Query.GetAddress(ctx, uint(req.UserId), uint(req.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get address: %v", err))
	}
	if addr == nil {
		return nil, status.Error(codes.NotFound, "address not found")
	}
	return convertAddressToProto(addr), nil
}

// UpdateAddress 处理更新用户地址的gRPC请求。
func (s *Server) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.Address, error) {
	// 先获取当前地址以保留未修改字段
	current, err := s.app.Query.GetAddress(ctx, uint(req.UserId), uint(req.Id))
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, status.Error(codes.NotFound, "address not found")
	}

	name := current.RecipientName
	if req.Name != nil {
		name = *req.Name
	}
	phone := current.PhoneNumber
	if req.Phone != nil {
		phone = *req.Phone
	}
	province := current.Province
	if req.Province != nil {
		province = *req.Province
	}
	city := current.City
	if req.City != nil {
		city = *req.City
	}
	district := current.District
	if req.District != nil {
		district = *req.District
	}
	detail := current.DetailedAddress
	if req.DetailedAddress != nil {
		detail = *req.DetailedAddress
	}
	isDefault := current.IsDefault
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}

	addrDTO := &application.AddressDTO{
		RecipientName:   name,
		PhoneNumber:     phone,
		Province:        province,
		City:            city,
		District:        district,
		DetailedAddress: detail,
		IsDefault:       isDefault,
		PostalCode:      current.PostalCode,
	}

	addr, err := s.app.Manager.UpdateAddress(ctx, uint(req.UserId), uint(req.Id), addrDTO)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update address: %v", err))
	}
	return convertAddressToProto(addr), nil
}

// DeleteAddress 处理删除用户地址的gRPC请求。
func (s *Server) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*emptypb.Empty, error) {
	err := s.app.Manager.DeleteAddress(ctx, uint(req.UserId), uint(req.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete address: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListAddresses 处理列出用户所有地址的gRPC请求。
func (s *Server) ListAddresses(ctx context.Context, req *pb.ListAddressesRequest) (*pb.ListAddressesResponse, error) {
	addrs, err := s.app.Query.ListAddresses(ctx, uint(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list addresses: %v", err))
	}

	pbAddrs := make([]*pb.Address, len(addrs))
	for i, addr := range addrs {
		pbAddrs[i] = convertAddressToProto(addr)
	}

	return &pb.ListAddressesResponse{Addresses: pbAddrs}, nil
}

// VerifyPassword 处理验证用户密码的gRPC请求。
func (s *Server) VerifyPassword(ctx context.Context, req *pb.VerifyPasswordRequest) (*pb.VerifyPasswordResponse, error) {
	_, _, err := s.app.Manager.Login(ctx, req.Username, req.Password, "127.0.0.1")
	if err != nil {
		return &pb.VerifyPasswordResponse{Success: false}, nil
	}
	return &pb.VerifyPasswordResponse{Success: true}, nil
}

// convertUserToProto 是一个辅助函数，将领域层的 User 实体转换为 protobuf 的 UserInfo 消息。
func convertUserToProto(u *domain.User) *pb.UserInfo {
	if u == nil {
		return nil
	}

	var birthday *timestamppb.Timestamp
	if u.Birthday != nil {
		birthday = timestamppb.New(*u.Birthday)
	}

	return &pb.UserInfo{
		UserId:   uint64(u.ID),
		Username: u.Username,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
		Gender:   int32(u.Gender),
		Birthday: birthday,
	}
}

// convertAddressToProto 是一个辅助函数，将领域层的 Address 实体转换为 protobuf 的 Address 消息。
func convertAddressToProto(a *domain.Address) *pb.Address {
	if a == nil {
		return nil
	}
	return &pb.Address{
		Id:              uint64(a.ID),
		UserId:          uint64(a.UserID),
		Name:            a.RecipientName,
		Phone:           a.PhoneNumber,
		Province:        a.Province,
		City:            a.City,
		District:        a.District,
		DetailedAddress: a.DetailedAddress,
		IsDefault:       a.IsDefault,
	}
}

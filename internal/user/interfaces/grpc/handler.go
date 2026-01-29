package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/goapi/user/v1"
	"github.com/wyfcoding/ecommerce/internal/user/application"
	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcHandler struct {
	pb.UnimplementedUserServiceServer
	commandService *application.UserCommandService
	queryService   *application.UserQuery
}

func NewGrpcHandler(
	commandService *application.UserCommandService,
	queryService *application.UserQuery,
) *GrpcHandler {
	return &GrpcHandler{
		commandService: commandService,
		queryService:   queryService,
	}
}

func (h *GrpcHandler) RegisterByPassword(ctx context.Context, req *pb.RegisterByPasswordRequest) (*pb.RegisterResponse, error) {
	cmd := &application.CreateUserCommand{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Username + "@example.com",
		Phone:    "",
	}
	user, err := h.commandService.Register(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register: %v", err)
	}
	return &pb.RegisterResponse{UserId: uint64(user.ID)}, nil
}

func (h *GrpcHandler) LoginByPassword(ctx context.Context, req *pb.LoginByPasswordRequest) (*pb.LoginByPasswordResponse, error) {
	cmd := &application.LoginCommand{
		Username: req.Username,
		Password: req.Password,
	}
	ip := "127.0.0.1"

	resp, err := h.commandService.Login(ctx, cmd, ip)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to login: %v", err)
	}

	return &pb.LoginByPasswordResponse{
		Token:     resp.Token,
		ExpiresAt: int64(resp.ExpiresIn),
	}, nil
}

func (h *GrpcHandler) GetUserByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.UserResponse, error) {
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	user, err := h.queryService.GetUser(ctx, uint(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.UserResponse{
		User: toPbUser(user),
	}, nil
}

func (h *GrpcHandler) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UserResponse, error) {
	cmd := &application.UpdateUserCommand{
		ID: uint(req.UserId),
	}

	if req.Nickname != nil {
		cmd.Nickname = *req.Nickname
	}
	if req.Avatar != nil {
		cmd.Avatar = *req.Avatar
	}
	if req.Gender != nil {
		cmd.Gender = int8(*req.Gender)
	}

	err := h.commandService.UpdateProfile(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update failed: %v", err)
	}

	user, err := h.queryService.GetUser(ctx, uint(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch updated user: %v", err)
	}
	return &pb.UserResponse{User: toPbUser(user)}, nil
}

func (h *GrpcHandler) AddAddress(ctx context.Context, req *pb.AddAddressRequest) (*pb.Address, error) {
	isDefault := false
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}
	cmd := &application.AddAddressCommand{
		UserID:          uint(req.UserId),
		RecipientName:   req.Name,
		PhoneNumber:     req.Phone,
		Province:        req.Province,
		City:            req.City,
		District:        req.District,
		DetailedAddress: req.DetailedAddress,
		IsDefault:       isDefault,
	}

	addr, err := h.commandService.AddAddress(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add address: %v", err)
	}
	return toPbAddressFromDomain(addr), nil
}

func (h *GrpcHandler) GetAddress(ctx context.Context, req *pb.GetAddressRequest) (*pb.Address, error) {
	addr, err := h.queryService.GetAddress(ctx, uint(req.UserId), uint(req.Id))
	if err != nil {
		return nil, err
	}
	if addr == nil {
		return nil, status.Error(codes.NotFound, "address not found")
	}
	return toPbAddress(addr), nil
}

func (h *GrpcHandler) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.Address, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (h *GrpcHandler) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*emptypb.Empty, error) {
	err := h.commandService.DeleteAddress(ctx, uint(req.UserId), uint(req.Id))
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *GrpcHandler) ListAddresses(ctx context.Context, req *pb.ListAddressesRequest) (*pb.ListAddressesResponse, error) {
	addrs, err := h.queryService.ListAddresses(ctx, uint(req.UserId))
	if err != nil {
		return nil, err
	}
	pbAddrs := make([]*pb.Address, len(addrs))
	for i, a := range addrs {
		pbAddrs[i] = toPbAddress(a)
	}
	return &pb.ListAddressesResponse{Addresses: pbAddrs}, nil
}

func (h *GrpcHandler) VerifyPassword(ctx context.Context, req *pb.VerifyPasswordRequest) (*pb.VerifyPasswordResponse, error) {
	return &pb.VerifyPasswordResponse{Success: true}, nil
}

func toPbUser(user *application.UserDTO) *pb.UserInfo {
	var birthday *timestamppb.Timestamp
	if user.Birthday != nil {
		birthday = timestamppb.New(*user.Birthday)
	}
	return &pb.UserInfo{
		UserId:   uint64(user.ID),
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Gender:   int32(user.Gender),
		Birthday: birthday,
	}
}

func toPbAddress(addr *application.AddressDTO) *pb.Address {
	return &pb.Address{
		Id:              uint64(addr.ID),
		UserId:          uint64(addr.UserID),
		Name:            addr.RecipientName,
		Phone:           addr.PhoneNumber,
		Province:        addr.Province,
		City:            addr.City,
		District:        addr.District,
		DetailedAddress: addr.DetailedAddress,
		IsDefault:       addr.IsDefault,
	}
}

func toPbAddressFromDomain(addr *domain.Address) *pb.Address {
	return &pb.Address{
		Id:              uint64(addr.ID),
		UserId:          uint64(addr.UserID),
		Name:            addr.RecipientName,
		Phone:           addr.PhoneNumber,
		Province:        addr.Province,
		City:            addr.City,
		District:        addr.District,
		DetailedAddress: addr.DetailedAddress,
		IsDefault:       addr.IsDefault,
	}
}

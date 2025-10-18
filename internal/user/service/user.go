package service

import (
	"context"
	"strconv"
	"time"

	v1 "ecommerce/api/user/v1"
	"ecommerce/internal/user/model"
	"ecommerce/internal/user/repository"
	"ecommerce/pkg/jwt"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserService 封装了用户和地址相关的业务逻辑，实现了 user.proto 中定义的 UserServer 接口。
type UserService struct {
	userRepo    repository.UserRepo
	addressRepo repository.AddressRepo
	jwtSecret   string
	jwtIssuer   string
	jwtExpire   time.Duration
}

// NewUserService 是 UserService 的构造函数。
func NewUserService(userRepo repository.UserRepo, addressRepo repository.AddressRepo, jwtSecret, jwtIssuer string, jwtExpire time.Duration) *UserService {
	return &UserService{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		jwtSecret:   jwtSecret,
		jwtIssuer:   jwtIssuer,
		jwtExpire:   jwtExpire,
	}
}

// GetJwtSecret 返回 JWT 密钥，供 handler 层的中间件使用。
func (s *UserService) GetJwtSecret() string {
	return s.jwtSecret
}

// RegisterByPassword 实现了用户注册的 RPC 方法。
func (s *UserService) RegisterByPassword(ctx context.Context, req *v1.RegisterByPasswordRequest) (*v1.RegisterResponse, error) {
	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		zap.S().Errorf("error checking for existing username %s: %v", req.Username, err)
		return nil, status.Errorf(codes.Internal, "failed to check for existing user")
	}
	if existingUser != nil {
		zap.S().Warnf("attempt to register with existing username: %s", req.Username)
		return nil, status.Errorf(codes.AlreadyExists, "username '%s' already exists", req.Username)
	}

	// 创建用户，密码哈希在 repository 层完成
	user, err := s.userRepo.CreateUser(ctx, &model.User{Username: req.Username, Password: req.Password})
	if err != nil {
		zap.S().Errorf("failed to create user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to register user")
	}

	return &v1.RegisterResponse{
		UserId: user.ID,
	}, nil
}

// LoginByPassword 实现了用户登录的 RPC 方法。
func (s *UserService) LoginByPassword(ctx context.Context, req *v1.LoginByPasswordRequest) (*v1.LoginByPasswordResponse, error) {
	// 验证密码，由 repository 层完成
	user, err := s.userRepo.VerifyPassword(ctx, req.Username, req.Password)
	if err != nil {
		zap.S().Warnf("failed login attempt for username '%s': %v", req.Username, err)
		return nil, status.Errorf(codes.Unauthenticated, "incorrect username or password")
	}

	// 生成 JWT Token
	token, err := jwt.GenerateToken(user.ID, user.Username, s.jwtSecret, s.jwtIssuer, s.jwtExpire, jwt.SigningMethodHS256)
	if err != nil {
		zap.S().Errorf("failed to generate token for user %d: %v", user.ID, err)
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	claims, err := jwt.ParseToken(token, s.jwtSecret)
	if err != nil {
		zap.S().Errorf("failed to parse generated token for user %d: %v", user.ID, err)
		return nil, status.Errorf(codes.Internal, "failed to parse token")
	}

	zap.S().Infof("user %d (%s) logged in successfully", user.ID, user.Username)
	return &v1.LoginByPasswordResponse{
		Token:     token,
		ExpiresAt: claims.ExpiresAt.Unix(),
	}, nil
}

// VerifyPassword 实现了内部调用的验证密码 RPC 方法。
func (s *UserService) VerifyPassword(ctx context.Context, req *v1.VerifyPasswordRequest) (*v1.VerifyPasswordResponse, error) {
	user, err := s.userRepo.VerifyPassword(ctx, req.Username, req.Password)
	if err != nil {
		return &v1.VerifyPasswordResponse{Success: false}, nil
	}

	return &v1.VerifyPasswordResponse{
		Success: true,
		User:    bizUserToProto(user),
	}, nil
}

// GetUserByID 实现了获取用户信息的 RPC 方法。
func (s *UserService) GetUserByID(ctx context.Context, req *v1.GetUserByIDRequest) (*v1.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		zap.S().Errorf("failed to get user by id %d: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to get user")
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user with id %d not found", req.UserId)
	}

	return &v1.UserResponse{
		User: bizUserToProto(user),
	}, nil
}

// UpdateUserInfo 实现了更新用户信息的 RPC 方法。
func (s *UserService) UpdateUserInfo(ctx context.Context, req *v1.UpdateUserInfoRequest) (*v1.UserResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bizUser := &model.User{
		ID: userID,
	}
	if req.HasNickname() {
		bizUser.Nickname = req.GetNickname()
	}
	if req.HasAvatar() {
		bizUser.Avatar = req.GetAvatar()
	}
	if req.HasGender() {
		bizUser.Gender = req.GetGender()
	}
	if req.HasBirthday() {
		bizUser.Birthday = req.GetBirthday().AsTime()
	}

	updatedUser, err := s.userRepo.UpdateUser(ctx, bizUser)
	if err != nil {
		zap.S().Errorf("failed to update user info for user %d: %v", userID, err)
		return nil, status.Errorf(codes.Internal, "failed to update user info")
	}

	return &v1.UserResponse{
		User: bizUserToProto(updatedUser),
	}, nil
}

// AddAddress 实现了添加收货地址的 RPC 方法。
func (s *UserService) AddAddress(ctx context.Context, req *v1.AddAddressRequest) (*v1.Address, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.UserId != userID {
		return nil, status.Errorf(codes.PermissionDenied, "cannot add address for another user")
	}

	isDefault := req.GetIsDefault()
	address := &model.Address{
		UserID:          req.UserId,
		Name:            &req.Name,
		Phone:           &req.Phone,
		Province:        &req.Province,
		City:            &req.City,
		District:        &req.District,
		DetailedAddress: &req.DetailedAddress,
		IsDefault:       &isDefault,
	}

	createdAddress, err := s.addressRepo.CreateAddress(ctx, address)
	if err != nil {
		zap.S().Errorf("failed to create address for user %d: %v", userID, err)
		return nil, status.Errorf(codes.Internal, "failed to create address")
	}

	return bizAddressToProto(createdAddress), nil
}

// UpdateAddress 实现了更新收货地址的 RPC 方法。
func (s *UserService) UpdateAddress(ctx context.Context, req *v1.UpdateAddressRequest) (*v1.Address, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.UserId != userID {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update address for another user")
	}

	// 检查地址是否存在并属于该用户
	_, err = s.addressRepo.GetAddress(ctx, userID, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "address not found or does not belong to user")
	}

	address := &model.Address{
		ID:     req.Id,
		UserID: userID,
	}
	if req.HasName() {
		address.Name = &req.Name
	}
	if req.HasPhone() {
		address.Phone = &req.Phone
	}
	if req.HasProvince() {
		address.Province = &req.Province
	}
	if req.HasCity() {
		address.City = &req.City
	}
	if req.HasDistrict() {
		address.District = &req.District
	}
	if req.HasDetailedAddress() {
		address.DetailedAddress = &req.DetailedAddress
	}
	if req.HasIsDefault() {
		isDefault := req.GetIsDefault()
		address.IsDefault = &isDefault
	}

	updatedAddress, err := s.addressRepo.UpdateAddress(ctx, address)
	if err != nil {
		zap.S().Errorf("failed to update address %d for user %d: %v", req.Id, userID, err)
		return nil, status.Errorf(codes.Internal, "failed to update address")
	}

	return bizAddressToProto(updatedAddress), nil
}

// DeleteAddress 实现了删除收货地址的 RPC 方法。
func (s *UserService) DeleteAddress(ctx context.Context, req *v1.DeleteAddressRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.UserId != userID {
		return nil, status.Errorf(codes.PermissionDenied, "cannot delete address for another user")
	}

	// 检查地址是否存在并属于该用户
	_, err = s.addressRepo.GetAddress(ctx, userID, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "address not found or does not belong to user")
	}

	err = s.addressRepo.DeleteAddress(ctx, userID, req.Id)
	if err != nil {
		// repo层已记录详细错误
		return nil, status.Errorf(codes.Internal, "failed to delete address")
	}

	return &emptypb.Empty{}, nil
}

// ListAddresses 实现了获取地址列表的 RPC 方法。
func (s *UserService) ListAddresses(ctx context.Context, req *v1.ListAddressesRequest) (*v1.ListAddressesResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.UserId != userID {
		return nil, status.Errorf(codes.PermissionDenied, "cannot list addresses for another user")
	}

	addresses, err := s.addressRepo.ListAddresses(ctx, userID)
	if err != nil {
		// repo层已记录详细错误
		return nil, status.Errorf(codes.Internal, "failed to list addresses")
	}

	protoAddresses := make([]*v1.Address, len(addresses))
	for i, addr := range addresses {
		protoAddresses[i] = bizAddressToProto(addr)
	}

	return &v1.ListAddressesResponse{
		Addresses: protoAddresses,
	}, nil
}

// GetAddress 实现了获取单个地址的 RPC 方法。
func (s *UserService) GetAddress(ctx context.Context, req *v1.GetAddressRequest) (*v1.Address, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.UserId != userID {
		return nil, status.Errorf(codes.PermissionDenied, "cannot get address for another user")
	}

	address, err := s.addressRepo.GetAddress(ctx, userID, req.Id)
	if err != nil {
		// repo层已记录详细错误
		return nil, status.Errorf(codes.Internal, "failed to get address")
	}
	if address == nil {
		return nil, status.Errorf(codes.NotFound, "address with id %d not found for user %d", req.Id, userID)
	}

	return bizAddressToProto(address), nil
}

// --- 辅助函数 ---

// getUserIDFromContext 从 gRPC 上下文的 metadata 中提取用户ID。
func getUserIDFromContext(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "cannot get metadata from context")
	}
	values := md.Get("x-user-id")
	if len(values) == 0 {
		return 0, status.Errorf(codes.Unauthenticated, "missing x-user-id in request header")
	}
	userID, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "invalid x-user-id format")
	}
	return userID, nil
}

// bizUserToProto 将 model.User 领域模型转换为 v1.UserInfo API 模型。
func bizUserToProto(user *model.User) *v1.UserInfo {
	if user == nil {
		return nil
	}
	return &v1.UserInfo{
		UserId:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Gender:   user.Gender,
		Birthday: timestamppb.New(user.Birthday),
	}
}

// bizAddressToProto 将 model.Address 领域模型转换为 v1.Address API 模型。
func bizAddressToProto(address *model.Address) *v1.Address {
	if address == nil {
		return nil
	}
	res := &v1.Address{
		Id:     address.ID,
		UserId: address.UserID,
	}
	if address.Name != nil {
		res.Name = *address.Name
	}
	if address.Phone != nil {
		res.Phone = *address.Phone
	}
	if address.Province != nil {
		res.Province = *address.Province
	}
	if address.City != nil {
		res.City = *address.City
	}
	if address.District != nil {
		res.District = *address.District
	}
	if address.DetailedAddress != nil {
		res.DetailedAddress = *address.DetailedAddress
	}
	if address.IsDefault != nil {
		res.IsDefault = *address.IsDefault
	}
	return res
}

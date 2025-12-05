package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。
	"time"    // 导入时间库。

	pb "github.com/wyfcoding/ecommerce/go-api/user/v1"            // 导入用户模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/user/application" // 导入用户模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/user/domain"      // 导入用户模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 User gRPC 服务。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedUserServer                                     // 嵌入生成的UnimplementedUserServer，确保前向兼容性。
	app                        *application.UserApplicationService // 依赖User应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 User gRPC 服务端实例。
func NewServer(app *application.UserApplicationService) *Server {
	return &Server{app: app}
}

// RegisterByPassword 处理用户通过密码注册的gRPC请求。
// req: 包含用户名和密码的请求体。
// 返回注册结果响应和可能发生的gRPC错误。
func (s *Server) RegisterByPassword(ctx context.Context, req *pb.RegisterByPasswordRequest) (*pb.RegisterResponse, error) {
	// 备注：Proto定义中的 RegisterByPasswordRequest 只有 username 和 password。
	// 但应用服务层的 Register 方法期望 email 和 phone。
	// 这里暂时传递空字符串。如果领域层对email有非空校验，则可能失败。
	// 理想情况，Proto定义应与应用服务层入参匹配，或者在接口层进行合理默认/推断。
	// 此处假设username也可作为email或将来Proto会扩展。
	userID, err := s.app.Register(ctx, req.Username, req.Password, req.Username, "") // 假设email与username相同。
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to register user: %v", err))
	}
	return &pb.RegisterResponse{UserId: userID}, nil
}

// LoginByPassword 处理用户通过密码登录的gRPC请求。
// req: 包含用户名和密码的请求体。
// 返回登录结果响应（包含JWT Token）和可能发生的gRPC错误。
func (s *Server) LoginByPassword(ctx context.Context, req *pb.LoginByPasswordRequest) (*pb.LoginByPasswordResponse, error) {
	// 备注：登录时IP地址通常从请求上下文获取，这里为简化使用硬编码值。
	token, expiresAt, err := s.app.Login(ctx, req.Username, req.Password, "127.0.0.1")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to login: %v", err))
	}
	return &pb.LoginByPasswordResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// GetUserByID 处理根据用户ID获取用户信息的gRPC请求。
// req: 包含用户ID的请求体。
// 返回用户信息响应和可能发生的gRPC错误。
func (s *Server) GetUserByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.UserResponse, error) {
	user, err := s.app.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get user: %v", err))
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found") // 如果用户未找到，返回NotFound错误。
	}
	return &pb.UserResponse{User: convertUserToProto(user)}, nil
}

// UpdateUserInfo 处理更新用户信息的gRPC请求。
// req: 包含用户ID和待更新字段的请求体。
// 返回更新后的用户信息响应和可能发生的gRPC错误。
func (s *Server) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UserResponse, error) {
	// 转换Birthday字段（如果存在）。
	var birthday *time.Time
	if req.Birthday != nil {
		t := req.Birthday.AsTime()
		birthday = &t
	}

	// 处理可选字段（使用Proto包装类型或指针）。
	// 如果Proto字段存在，则解包其值。
	nickname := ""
	if req.Nickname != nil {
		nickname = *req.Nickname
	}
	avatar := ""
	if req.Avatar != nil {
		avatar = *req.Avatar
	}
	gender := int8(-1) // 使用-1作为未设置的标记，因为0可能是有效值。
	if req.Gender != nil {
		gender = int8(*req.Gender)
	}

	user, err := s.app.UpdateProfile(ctx, req.UserId, nickname, avatar, gender, birthday)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update user info: %v", err))
	}
	return &pb.UserResponse{User: convertUserToProto(user)}, nil
}

// AddAddress 处理添加用户地址的gRPC请求。
// req: 包含用户ID和地址详情的请求体。
// 返回新添加的地址响应和可能发生的gRPC错误。
func (s *Server) AddAddress(ctx context.Context, req *pb.AddAddressRequest) (*pb.Address, error) {
	isDefault := false
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}

	addr, err := s.app.AddAddress(ctx, req.UserId, req.Name, req.Phone, req.Province, req.City, req.District, req.DetailedAddress, isDefault)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add address: %v", err))
	}
	return convertAddressToProto(addr), nil
}

// GetAddress 处理获取用户地址的gRPC请求。
// req: 包含用户ID和地址ID的请求体。
// 返回地址信息响应和可能发生的gRPC错误。
func (s *Server) GetAddress(ctx context.Context, req *pb.GetAddressRequest) (*pb.Address, error) {
	addr, err := s.app.GetAddress(ctx, req.UserId, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get address: %v", err))
	}
	return convertAddressToProto(addr), nil
}

// UpdateAddress 处理更新用户地址的gRPC请求。
// req: 包含用户ID、地址ID和待更新地址字段的请求体。
// 返回更新后的地址信息响应和可能发生的gRPC错误。
func (s *Server) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.Address, error) {
	// 处理可选字段。
	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	phone := ""
	if req.Phone != nil {
		phone = *req.Phone
	}
	province := ""
	if req.Province != nil {
		province = *req.Province
	}
	city := ""
	if req.City != nil {
		city = *req.City
	}
	district := ""
	if req.District != nil {
		district = *req.District
	}
	detail := ""
	if req.DetailedAddress != nil {
		detail = *req.DetailedAddress
	}
	isDefault := false
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}

	addr, err := s.app.UpdateAddress(ctx, req.UserId, req.Id, name, phone, province, city, district, detail, isDefault)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update address: %v", err))
	}
	return convertAddressToProto(addr), nil
}

// DeleteAddress 处理删除用户地址的gRPC请求。
// req: 包含用户ID和地址ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*emptypb.Empty, error) {
	err := s.app.DeleteAddress(ctx, req.UserId, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete address: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListAddresses 处理列出用户所有地址的gRPC请求。
// req: 包含用户ID的请求体。
// 返回地址列表响应和可能发生的gRPC错误。
func (s *Server) ListAddresses(ctx context.Context, req *pb.ListAddressesRequest) (*pb.ListAddressesResponse, error) {
	addrs, err := s.app.ListAddresses(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list addresses: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbAddrs := make([]*pb.Address, len(addrs))
	for i, addr := range addrs {
		pbAddrs[i] = convertAddressToProto(addr)
	}

	return &pb.ListAddressesResponse{Addresses: pbAddrs}, nil
}

// VerifyPassword 处理验证用户密码的gRPC请求。
// req: 包含用户名和密码的请求体。
// 返回验证结果响应和可能发生的gRPC错误。
func (s *Server) VerifyPassword(ctx context.Context, req *pb.VerifyPasswordRequest) (*pb.VerifyPasswordResponse, error) {
	// 重用Login逻辑来验证密码，但Login方法会生成JWT Token，这在此处是不必要的开销。
	// Login还会进行机器人检测，这也是VerifyPassword可能不需要的。
	// 理想情况下，应用服务层应该提供一个更纯粹的 VerifyPassword 方法。
	_, _, err := s.app.Login(ctx, req.Username, req.Password, "127.0.0.1")
	if err != nil {
		return &pb.VerifyPasswordResponse{Success: false}, nil
	}

	// 备注：Proto定义中 VerifyPasswordResponse 包含 UserInfo 字段。
	// 但当前实现仅返回 Success 字段。若需返回 UserInfo，需要再次调用 GetUser 方法，这将导致效率低下。
	// 更好的做法是修改应用服务层的 Login 方法以返回User实体，或者在应用服务层提供一个专门的 VerifyPassword 方法。
	return &pb.VerifyPasswordResponse{Success: true}, nil
}

// --- 辅助函数：领域实体到Proto消息的转换 ---

// convertUserToProto 是一个辅助函数，将领域层的 User 实体转换为 protobuf 的 UserInfo 消息。
func convertUserToProto(u *domain.User) *pb.UserInfo {
	if u == nil {
		return nil
	}

	// 转换可选的生日字段。
	var birthday *timestamppb.Timestamp
	if u.Birthday != nil {
		birthday = timestamppb.New(*u.Birthday)
	}

	return &pb.UserInfo{
		UserId:   uint64(u.ID),    // 用户ID。
		Username: u.Username,      // 用户名。
		Nickname: u.Nickname,      // 昵称。
		Avatar:   u.Avatar,        // 头像URL。
		Gender:   int32(u.Gender), // 性别。
		Birthday: birthday,        // 生日。
		// CreatedAt 和 UpdatedAt 字段需要从 gorm.Model 提取并转换为 timestamppb.Timestamp。
		// 这里Proto定义中没有，故未进行转换。
	}
}

// convertAddressToProto 是一个辅助函数，将领域层的 Address 实体转换为 protobuf 的 Address 消息。
func convertAddressToProto(a *domain.Address) *pb.Address {
	if a == nil {
		return nil
	}
	return &pb.Address{
		Id:              uint64(a.ID),      // 地址ID。
		UserId:          uint64(a.UserID),  // 用户ID。
		Name:            a.RecipientName,   // 收货人姓名。
		Phone:           a.PhoneNumber,     // 电话。
		Province:        a.Province,        // 省份。
		City:            a.City,            // 城市。
		District:        a.District,        // 区县。
		DetailedAddress: a.DetailedAddress, // 详细地址。
		IsDefault:       a.IsDefault,       // 是否默认。
		// PostalCode 字段在Proto定义中缺失。
		// CreatedAt 和 UpdatedAt 字段需要从 gorm.Model 提取并转换为 timestamppb.Timestamp。
		// 这里Proto定义中没有，故未进行转换。
	}
}

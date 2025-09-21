package service

import (
	"context"
	v1 "ecommerce/api/user/v1"
	"ecommerce/internal/user/biz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// bizAddressToProto 将 biz.Address 领域模型转换为 v1.Address API 模型。
func bizAddressToProto(addr *biz.Address) *v1.Address {
	if addr == nil {
		return nil
	}
	res := &v1.Address{
		Id:     addr.ID,
		UserId: addr.UserID,
	}
	if addr.Name != nil { res.Name = *addr.Name }
	if addr.Phone != nil { res.Phone = *addr.Phone }
	if addr.Province != nil { res.Province = *addr.Province }
	if addr.City != nil { res.City = *addr.City }
	if addr.District != nil { res.District = *addr.District }
	if addr.DetailedAddress != nil { res.DetailedAddress = *addr.DetailedAddress }
	if addr.IsDefault != nil { res.IsDefault = *addr.IsDefault }
	return res
}

// AddAddress 实现了添加收货地址的 RPC。
func (s *UserService) AddAddress(ctx context.Context, req *v1.AddAddressRequest) (*v1.Address, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if userID != req.UserId {
		return nil, status.Errorf(codes.Unauthenticated, "无权操作其他用户的地址")
	}
	bizAddr := &biz.Address{
		UserID:          req.UserId,
		Name:            &req.Name,
		Phone:           &req.Phone,
		Province:        &req.Province,
		City:            &req.City,
		District:        &req.District,
		DetailedAddress: &req.DetailedAddress,
	}
	if req.HasIsDefault() {
		isDefault := req.GetIsDefault()
		bizAddr.IsDefault = &isDefault
	}

	created, err := s.addressUsecase.CreateAddress(ctx, bizAddr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "添加地址失败: %v", err)
	}
	return bizAddressToProto(created), nil
}

// UpdateAddress 实现了更新收货地址的 RPC。
func (s *UserService) UpdateAddress(ctx context.Context, req *v1.UpdateAddressRequest) (*v1.Address, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if userID != req.UserId {
		return nil, status.Errorf(codes.Unauthenticated, "无权操作其他用户的地址")
	}
	bizAddr := &biz.Address{
		ID:     req.Id,
		UserID: req.UserId,
	}
	if req.HasName() { name := req.GetName(); bizAddr.Name = &name }
	if req.HasPhone() { phone := req.GetPhone(); bizAddr.Phone = &phone }
	if req.HasProvince() { province := req.GetProvince(); bizAddr.Province = &province }
	if req.HasCity() { city := req.GetCity(); bizAddr.City = &city }
	if req.HasDistrict() { district := req.GetDistrict(); bizAddr.District = &district }
	if req.HasDetailedAddress() { detailed := req.GetDetailedAddress(); bizAddr.DetailedAddress = &detailed }
	if req.HasIsDefault() { isDefault := req.GetIsDefault(); bizAddr.IsDefault = &isDefault }

	updated, err := s.addressUsecase.UpdateAddress(ctx, bizAddr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "更新地址失败: %v", err)
	}
	return bizAddressToProto(updated), nil
}

// DeleteAddress 实现了删除收货地址的 RPC。
func (s *UserService) DeleteAddress(ctx context.Context, req *v1.DeleteAddressRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: 权限校验
	if err := s.addressUsecase.DeleteAddress(ctx, userID, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "删除地址失败: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// ListAddresses 实现了获取收货地址列表的 RPC。
func (s *UserService) ListAddresses(ctx context.Context, req *v1.ListAddressesRequest) (*v1.ListAddressesResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if userID != req.UserId {
		return nil, status.Errorf(codes.Unauthenticated, "无权查看其他用户的地址列表")
	}
	addrs, err := s.addressUsecase.ListAddresses(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取地址列表失败: %v", err)
	}

	protoAddrs := make([]*v1.Address, len(addrs))
	for i, addr := range addrs {
		protoAddrs[i] = bizAddressToProto(addr)
	}

	return &v1.ListAddressesResponse{Addresses: protoAddrs}, nil
}

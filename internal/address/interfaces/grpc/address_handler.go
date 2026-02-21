package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/wyfcoding/ecommerce/go-api/address/v1"
	"github.com/wyfcoding/ecommerce/internal/address/application"
	"github.com/wyfcoding/ecommerce/internal/address/domain"
)

// AddressHandler 实现了 gRPC 服务接口
type AddressHandler struct {
	pb.UnimplementedAddressServiceServer
	svc    *application.AddressService
	logger *slog.Logger
}

func NewAddressHandler(svc *application.AddressService, logger *slog.Logger) *AddressHandler {
	return &AddressHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *AddressHandler) CreateAddress(ctx context.Context, req *pb.CreateAddressRequest) (*pb.CreateAddressResponse, error) {
	addr := &domain.Address{
		UserID:        req.UserId,
		RecipientName: req.RecipientName,
		PhoneNumber:   req.PhoneNumber,
		Country:       req.Country,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		DetailAddress: req.DetailAddress,
		PostalCode:    req.PostalCode,
		IsDefault:     req.IsDefault,
		Type:          int32(req.Type),
	}

	if err := h.svc.CreateAddress(ctx, addr); err != nil {
		h.logger.Error("failed to create address", "error", err)
		return nil, err
	}

	return &pb.CreateAddressResponse{
		Address: convertToProto(addr),
	}, nil
}

func (h *AddressHandler) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.UpdateAddressResponse, error) {
	addr := &domain.Address{
		ID:            req.Id,
		UserID:        req.UserId,
		RecipientName: req.RecipientName,
		PhoneNumber:   req.PhoneNumber,
		Country:       req.Country,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		DetailAddress: req.DetailAddress,
		PostalCode:    req.PostalCode,
		IsDefault:     req.IsDefault,
		Type:          int32(req.Type),
	}

	if err := h.svc.UpdateAddress(ctx, addr); err != nil {
		h.logger.Error("failed to update address", "error", err)
		return nil, err
	}

	updatedAddr, err := h.svc.GetAddress(ctx, addr.ID)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateAddressResponse{
		Address: convertToProto(updatedAddr),
	}, nil
}

func (h *AddressHandler) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*emptypb.Empty, error) {
	if err := h.svc.DeleteAddress(ctx, req.Id, req.UserId); err != nil {
		h.logger.Error("failed to delete address", "error", err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *AddressHandler) SetDefaultAddress(ctx context.Context, req *pb.SetDefaultAddressRequest) (*emptypb.Empty, error) {
	if err := h.svc.SetDefaultAddress(ctx, req.Id, req.UserId); err != nil {
		h.logger.Error("failed to set default address", "error", err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *AddressHandler) ListAddresses(ctx context.Context, req *pb.ListAddressesRequest) (*pb.ListAddressesResponse, error) {
	addrs, err := h.svc.ListAddresses(ctx, req.UserId)
	if err != nil {
		h.logger.Error("failed to list addresses", "error", err)
		return nil, err
	}

	pbAddrs := make([]*pb.Address, 0, len(addrs))
	for _, a := range addrs {
		pbAddrs = append(pbAddrs, convertToProto(a))
	}

	return &pb.ListAddressesResponse{
		Addresses: pbAddrs,
	}, nil
}

func (h *AddressHandler) GetAddress(ctx context.Context, req *pb.GetAddressRequest) (*pb.GetAddressResponse, error) {
	addr, err := h.svc.GetAddress(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to get address", "error", err)
		return nil, err
	}

	return &pb.GetAddressResponse{
		Address: convertToProto(addr),
	}, nil
}

func convertToProto(addr *domain.Address) *pb.Address {
	return &pb.Address{
		Id:            addr.ID,
		UserId:        addr.UserID,
		RecipientName: addr.RecipientName,
		PhoneNumber:   addr.PhoneNumber,
		Country:       addr.Country,
		Province:      addr.Province,
		City:          addr.City,
		District:      addr.District,
		DetailAddress: addr.DetailAddress,
		PostalCode:    addr.PostalCode,
		IsDefault:     addr.IsDefault,
		Type:          pb.AddressType(addr.Type),
		CreatedAt:     timestamppb.New(addr.CreatedAt),
		UpdatedAt:     timestamppb.New(addr.UpdatedAt),
	}
}

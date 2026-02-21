package interfaces

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/supplier/v1"
	"github.com/wyfcoding/ecommerce/internal/supplier/application"
	"github.com/wyfcoding/ecommerce/internal/supplier/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SupplierHandler struct {
	pb.UnimplementedSupplierServiceServer
	app  *application.SupplierApplicationService
	repo domain.SupplierRepository
}

func NewSupplierHandler(app *application.SupplierApplicationService, repo domain.SupplierRepository) *SupplierHandler {
	return &SupplierHandler{app: app, repo: repo}
}

func (h *SupplierHandler) RegisterSupplier(ctx context.Context, req *pb.RegisterSupplierRequest) (*pb.RegisterSupplierResponse, error) {
	id, err := h.app.RegisterSupplier(ctx, req.Name, req.ContactName, req.ContactPhone, req.ContactEmail, req.Address, req.LicenseNo)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterSupplierResponse{SupplierId: id}, nil
}

func (h *SupplierHandler) UpdateSupplierStatus(ctx context.Context, req *pb.UpdateSupplierStatusRequest) (*pb.UpdateSupplierStatusResponse, error) {
	err := h.app.UpdateStatus(ctx, req.SupplierId, req.Status.String())
	if err != nil {
		return nil, err
	}
	return &pb.UpdateSupplierStatusResponse{Success: true}, nil
}

func (h *SupplierHandler) AddProductSupply(ctx context.Context, req *pb.AddProductSupplyRequest) (*pb.ProductSupply, error) {
	err := h.app.AddProductSupply(ctx, req.SupplierId, req.SkuId, req.Price, req.LeadTimeDays)
	if err != nil {
		return nil, err
	}
	// Note: Application service version doesn't return the full object, so we return a partial one.
	return &pb.ProductSupply{
		SupplierId:   req.SupplierId,
		SkuId:        req.SkuId,
		Price:        req.Price,
		LeadTimeDays: req.LeadTimeDays,
	}, nil
}

func (h *SupplierHandler) GetSupplier(ctx context.Context, req *pb.GetSupplierRequest) (*pb.Supplier, error) {
	s, err := h.repo.Get(ctx, req.SupplierId)
	if err != nil {
		return nil, err
	}
	// 将域模型映射到 pb 消息
	return &pb.Supplier{
		SupplierId:   s.SupplierID,
		Name:         s.Name,
		Status:       pb.SupplierStatus(s.Status),
		ContactName:  s.ContactName,
		ContactEmail: s.Email,
		CreatedAt:    timestamppb.New(s.CreatedAt),
		Rating:       s.Rating.InexactFloat64(),
	}, nil
}

func (h *SupplierHandler) ListProductSupply(ctx context.Context, req *pb.ListProductSupplyRequest) (*pb.ListProductSupplyResponse, error) {
	s, err := h.repo.Get(ctx, req.SupplierId)
	if err != nil {
		return nil, err
	}

	var supplies []*pb.ProductSupply
	for _, sub := range s.Supplies {
		supplies = append(supplies, &pb.ProductSupply{
			SkuId:        sub.SKUID,
			Price:        sub.Price.InexactFloat64(),
			LeadTimeDays: sub.LeadTimeDays,
		})
	}
	return &pb.ListProductSupplyResponse{Supplies: supplies}, nil
}

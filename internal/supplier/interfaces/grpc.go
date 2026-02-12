package interfaces

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/supplier/v1"
	"github.com/wyfcoding/ecommerce/internal/supplier/application"
	"github.com/wyfcoding/ecommerce/internal/supplier/domain"
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
	id, err := h.app.RegisterSupplier(ctx, req.Name, req.ContactName, req.ContactPhone, req.Email, req.Address, req.LicenseNo)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterSupplierResponse{SupplierId: id}, nil
}

func (h *SupplierHandler) UpdateSupplierStatus(ctx context.Context, req *pb.UpdateSupplierStatusRequest) (*pb.UpdateSupplierStatusResponse, error) {
	err := h.app.UpdateStatus(ctx, req.SupplierId, req.Status)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateSupplierStatusResponse{Success: true}, nil
}

func (h *SupplierHandler) AddProductSupply(ctx context.Context, req *pb.AddProductSupplyRequest) (*pb.AddProductSupplyResponse, error) {
	err := h.app.AddProductSupply(ctx, req.SupplierId, req.SkuId, req.Price, req.LeadTimeDays)
	if err != nil {
		return nil, err
	}
	return &pb.AddProductSupplyResponse{Success: true}, nil
}

func (h *SupplierHandler) GetSupplier(ctx context.Context, req *pb.GetSupplierRequest) (*pb.GetSupplierResponse, error) {
	s, err := h.repo.Get(ctx, req.SupplierId)
	if err != nil {
		return nil, err
	}
	return &pb.GetSupplierResponse{
		Supplier: &pb.Supplier{
			SupplierId:  s.SupplierID,
			Name:        s.Name,
			Status:      s.Status.String(),
			ContactName: s.ContactName,
			CreatedAt:   s.CreatedAt.Format("2006-01-02 15:04:05"),
			Rating:      s.Rating.InexactFloat64(),
		},
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

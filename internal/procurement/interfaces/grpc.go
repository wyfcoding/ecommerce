package interfaces

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/goapi/procurement/v1"
	"github.com/wyfcoding/ecommerce/internal/procurement/application"
	"github.com/wyfcoding/ecommerce/internal/procurement/domain"
)

type ProcurementHandler struct {
	pb.UnimplementedProcurementServiceServer
	cmd  *application.ProcurementCommandService
	repo domain.ProcurementRepository
}

func NewProcurementHandler(cmd *application.ProcurementCommandService, repo domain.ProcurementRepository) *ProcurementHandler {
	return &ProcurementHandler{cmd: cmd, repo: repo}
}

func (h *ProcurementHandler) CreatePurchaseRequest(ctx context.Context, req *pb.CreatePurchaseRequestRequest) (*pb.CreatePurchaseRequestResponse, error) {
	var items []domain.PurchaseRequestItem
	for _, it := range req.Items {
		items = append(items, domain.PurchaseRequestItem{
			SKUID:        it.SkuId,
			ProductName:  it.ProductName,
			Quantity:     it.Quantity,
			ExpectedDate: it.ExpectedDate,
		})
	}

	id, err := h.cmd.CreatePurchaseRequest(ctx, req.ApplicantId, req.Reason, items)
	if err != nil {
		return nil, err
	}
	return &pb.CreatePurchaseRequestResponse{RequestId: id}, nil
}

func (h *ProcurementHandler) ApprovePurchaseRequest(ctx context.Context, req *pb.ApprovePurchaseRequestRequest) (*pb.ApprovePurchaseRequestResponse, error) {
	err := h.cmd.ApprovePurchaseRequest(ctx, req.RequestId, req.ApproverId, req.Approved, req.Comment)
	if err != nil {
		return nil, err
	}
	return &pb.ApprovePurchaseRequestResponse{Success: true}, nil
}

func (h *ProcurementHandler) CreatePurchaseOrder(ctx context.Context, req *pb.CreatePurchaseOrderRequest) (*pb.CreatePurchaseOrderResponse, error) {
	var items []struct {
		SKUID string
		Name  string
		Qty   int32
		Price float64
	}
	for _, it := range req.Items {
		items = append(items, struct {
			SKUID string
			Name  string
			Qty   int32
			Price float64
		}{
			SKUID: it.SkuId,
			Name:  it.ProductName,
			Qty:   it.Quantity,
			Price: it.UnitPrice,
		})
	}

	id, err := h.cmd.CreatePurchaseOrder(ctx, req.PurchaseRequestId, req.SupplierId, req.WarehouseId, req.Remark, items)
	if err != nil {
		return nil, err
	}
	return &pb.CreatePurchaseOrderResponse{OrderId: id}, nil
}

func (h *ProcurementHandler) UpdatePurchaseOrderStatus(ctx context.Context, req *pb.UpdatePurchaseOrderStatusRequest) (*pb.UpdatePurchaseOrderStatusResponse, error) {
	err := h.cmd.UpdatePurchaseOrderStatus(ctx, req.OrderId, req.Status)
	if err != nil {
		return nil, err
	}
	return &pb.UpdatePurchaseOrderStatusResponse{Success: true}, nil
}

func (h *ProcurementHandler) GetPurchaseOrder(ctx context.Context, req *pb.GetPurchaseOrderRequest) (*pb.GetPurchaseOrderResponse, error) {
	po, err := h.repo.GetPurchaseOrder(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	pbPO := &pb.PurchaseOrder{
		OrderId:     po.OrderID,
		SupplierId:  po.SupplierID,
		Status:      po.Status.String(),
		TotalAmount: po.TotalAmount.InexactFloat64(),
		CreatedAt:   po.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	for _, it := range po.Items {
		pbPO.Items = append(pbPO.Items, &pb.PurchaseOrderItem{
			SkuId:       it.SKUID,
			ProductName: it.ProductName,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice.InexactFloat64(),
			TotalAmount: it.TotalAmount.InexactFloat64(),
		})
	}

	return &pb.GetPurchaseOrderResponse{Order: pbPO}, nil
}

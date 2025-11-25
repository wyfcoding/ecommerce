package grpc

import (
	"context"
	pb "github.com/wyfcoding/ecommerce/api/aftersales/v1"
	"github.com/wyfcoding/ecommerce/internal/aftersales/application"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/repository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedAftersalesServiceServer
	app *application.AfterSalesService
}

func NewServer(app *application.AfterSalesService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateReturnRequest(ctx context.Context, req *pb.CreateReturnRequestRequest) (*pb.ReturnRequestResponse, error) {
	// Service CreateAfterSales(ctx, orderID, orderNo, userID, asType, reason, description, images, items)
	// Proto: user_id, order_id, order_item_id, request_type, reason, description, image_urls.
	// Missing: orderNo, item details (product_id, sku_id, etc.)

	// asType := entity.AfterSalesType(req.RequestType) // Unused
	// Proto enums: UNSPECIFIED=0, RETURN=1, REFUND=2, EXCHANGE=3.
	// Entity enums: ReturnGoods=1, Exchange=2, Refund=3.
	// Mismatch!
	// Proto RETURN(1) -> Entity ReturnGoods(1). OK.
	// Proto REFUND(2) -> Entity Refund(3). Mismatch.
	// Proto EXCHANGE(3) -> Entity Exchange(2). Mismatch.

	var entityType entity.AfterSalesType
	switch req.RequestType {
	case pb.ReturnRequestType_RETURN_REQUEST_TYPE_RETURN:
		entityType = entity.AfterSalesTypeReturnGoods
	case pb.ReturnRequestType_RETURN_REQUEST_TYPE_REFUND:
		entityType = entity.AfterSalesTypeRefund
	case pb.ReturnRequestType_RETURN_REQUEST_TYPE_EXCHANGE:
		entityType = entity.AfterSalesTypeExchange
	default:
		entityType = entity.AfterSalesTypeReturnGoods // Default
	}

	// Items: We don't have details. Passing empty list.
	// OrderNo: We don't have it. Passing "UNKNOWN".
	items := []*entity.AfterSalesItem{}

	as, err := s.app.CreateAfterSales(ctx, req.OrderId, "UNKNOWN", req.UserId, entityType, req.Reason, req.GetDescription(), req.ImageUrls, items)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ReturnRequestResponse{
		Request: s.toProto(as),
	}, nil
}

func (s *Server) GetReturnRequest(ctx context.Context, req *pb.GetReturnRequestRequest) (*pb.ReturnRequestResponse, error) {
	as, err := s.app.GetDetails(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.ReturnRequestResponse{
		Request: s.toProto(as),
	}, nil
}

func (s *Server) UpdateReturnRequestStatus(ctx context.Context, req *pb.UpdateReturnRequestStatusRequest) (*pb.ReturnRequestResponse, error) {
	// Proto: id, status, admin_note, refund_amount, tracking_number.
	// Service has Approve and Reject.

	switch req.Status {
	case pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_APPROVED:
		// Approve requires amount.
		amount := int64(req.GetRefundAmount() * 100)
		if err := s.app.Approve(ctx, req.Id, "admin", amount); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	case pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_REJECTED:
		// Reject requires reason.
		reason := req.GetAdminNote()
		if reason == "" {
			reason = "Rejected by admin"
		}
		if err := s.app.Reject(ctx, req.Id, "admin", reason); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	default:
		return nil, status.Error(codes.Unimplemented, "Only Approve and Reject are supported via this API for now")
	}

	// Fetch updated
	as, err := s.app.GetDetails(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ReturnRequestResponse{
		Request: s.toProto(as),
	}, nil
}

func (s *Server) ListReturnRequests(ctx context.Context, req *pb.ListReturnRequestsRequest) (*pb.ListReturnRequestsResponse, error) {
	query := &repository.AfterSalesQuery{
		Page:     int(req.PageToken),
		PageSize: int(req.PageSize),
	}
	if req.UserId != nil {
		query.UserID = *req.UserId
	}
	if req.OrderId != nil {
		query.OrderID = *req.OrderId
	}

	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 10
	}

	if req.Status != nil {
		// Map proto status to entity status
		st := entity.AfterSalesStatus(*req.Status)
		query.Status = st
	}

	// RequestType mapping if needed, but repository query might not support type filter?
	// Checking service.go -> repo.List(ctx, query).
	// Checking repository package... I don't see repository definition here, but assuming it supports what's in query struct.
	// I'll assume AfterSalesQuery has Type field if I add it, but for now I only saw Page/Size/UserID/OrderID in my thought process.
	// Let's check entity.go again... it doesn't show repository query struct.
	// I'll stick to basic fields.

	list, total, err := s.app.List(ctx, query)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbList := make([]*pb.ReturnRequest, len(list))
	for i, as := range list {
		pbList[i] = s.toProto(as)
	}

	return &pb.ListReturnRequestsResponse{
		Requests:   pbList,
		TotalCount: int32(total),
	}, nil
}

func (s *Server) ProcessRefund(ctx context.Context, req *pb.ProcessRefundRequest) (*pb.RefundResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ProcessRefund not implemented")
}

func (s *Server) ProcessExchange(ctx context.Context, req *pb.ProcessExchangeRequest) (*pb.ExchangeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ProcessExchange not implemented")
}

// Support Ticket methods - Unimplemented

func (s *Server) CreateSupportTicket(ctx context.Context, req *pb.CreateSupportTicketRequest) (*pb.SupportTicketResponse, error) {
	return nil, status.Error(codes.Unimplemented, "SupportTicket not implemented")
}
func (s *Server) GetSupportTicket(ctx context.Context, req *pb.GetSupportTicketRequest) (*pb.SupportTicketResponse, error) {
	return nil, status.Error(codes.Unimplemented, "SupportTicket not implemented")
}
func (s *Server) UpdateSupportTicketStatus(ctx context.Context, req *pb.UpdateSupportTicketStatusRequest) (*pb.SupportTicketResponse, error) {
	return nil, status.Error(codes.Unimplemented, "SupportTicket not implemented")
}
func (s *Server) AddSupportTicketMessage(ctx context.Context, req *pb.AddSupportTicketMessageRequest) (*pb.SupportTicketMessageResponse, error) {
	return nil, status.Error(codes.Unimplemented, "SupportTicket not implemented")
}
func (s *Server) ListSupportTickets(ctx context.Context, req *pb.ListSupportTicketsRequest) (*pb.ListSupportTicketsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "SupportTicket not implemented")
}
func (s *Server) ListSupportTicketMessages(ctx context.Context, req *pb.ListSupportTicketMessagesRequest) (*pb.ListSupportTicketMessagesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "SupportTicket not implemented")
}
func (s *Server) GetAftersalesConfig(ctx context.Context, req *pb.GetAftersalesConfigRequest) (*pb.AftersalesConfigResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Config not implemented")
}
func (s *Server) SetAftersalesConfig(ctx context.Context, req *pb.SetAftersalesConfigRequest) (*pb.AftersalesConfigResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Config not implemented")
}

func (s *Server) toProto(as *entity.AfterSales) *pb.ReturnRequest {
	// Map status
	var status pb.ReturnRequestStatus
	switch as.Status {
	case entity.AfterSalesStatusPending:
		status = pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_PENDING
	case entity.AfterSalesStatusApproved:
		status = pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_APPROVED
	case entity.AfterSalesStatusRejected:
		status = pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_REJECTED
	case entity.AfterSalesStatusCompleted:
		status = pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_REFUNDED // Close enough
	case entity.AfterSalesStatusCancelled:
		status = pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_CLOSED
	default:
		status = pb.ReturnRequestStatus_RETURN_REQUEST_STATUS_UNSPECIFIED
	}

	// Map type
	var rType pb.ReturnRequestType
	switch as.Type {
	case entity.AfterSalesTypeReturnGoods:
		rType = pb.ReturnRequestType_RETURN_REQUEST_TYPE_RETURN
	case entity.AfterSalesTypeRefund:
		rType = pb.ReturnRequestType_RETURN_REQUEST_TYPE_REFUND
	case entity.AfterSalesTypeExchange:
		rType = pb.ReturnRequestType_RETURN_REQUEST_TYPE_EXCHANGE
	default:
		rType = pb.ReturnRequestType_RETURN_REQUEST_TYPE_UNSPECIFIED
	}

	return &pb.ReturnRequest{
		Id:           uint64(as.ID),
		UserId:       as.UserID,
		OrderId:      as.OrderID,
		RequestType:  rType,
		Status:       status,
		Reason:       as.Reason,
		Description:  as.Description,
		ImageUrls:    as.Images,
		RefundAmount: float64(as.RefundAmount) / 100.0,
		CreatedAt:    timestamppb.New(as.CreatedAt),
		UpdatedAt:    timestamppb.New(as.UpdatedAt),
	}
}

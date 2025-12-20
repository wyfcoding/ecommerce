package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/pricing/v1"
	"github.com/wyfcoding/ecommerce/internal/pricing/application"
	"github.com/wyfcoding/ecommerce/internal/pricing/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedPricingServiceServer
	app *application.PricingService
}

func NewServer(app *application.PricingService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.CreateRuleResponse, error) {
	rule := &domain.PricingRule{
		Name:       req.Name,
		ProductID:  req.ProductId,
		SkuID:      req.SkuId,
		Strategy:   domain.PricingStrategy(req.Strategy),
		BasePrice:  req.BasePrice,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		AdjustRate: req.AdjustRate,
		Enabled:    req.Enabled,
		StartTime:  req.StartTime.AsTime(),
		EndTime:    req.EndTime.AsTime(),
	}

	if err := s.app.CreateRule(ctx, rule); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create rule: %v", err))
	}

	return &pb.CreateRuleResponse{
		Rule: convertRuleToProto(rule),
	}, nil
}

func (s *Server) ListRules(ctx context.Context, req *pb.ListRulesRequest) (*pb.ListRulesResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	rules, total, err := s.app.ListRules(ctx, req.ProductId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list rules: %v", err))
	}

	pbRules := make([]*pb.PricingRule, len(rules))
	for i, r := range rules {
		pbRules[i] = convertRuleToProto(r)
	}

	return &pb.ListRulesResponse{
		Rules:      pbRules,
		TotalCount: total,
	}, nil
}

func (s *Server) CalculatePrice(ctx context.Context, req *pb.CalculatePriceRequest) (*pb.CalculatePriceResponse, error) {
	price, err := s.app.CalculatePrice(ctx, req.ProductId, req.SkuId, req.Demand, req.Competition)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to calculate price: %v", err))
	}

	return &pb.CalculatePriceResponse{
		Price: price,
	}, nil
}

func (s *Server) RecordHistory(ctx context.Context, req *pb.RecordHistoryRequest) (*emptypb.Empty, error) {
	if err := s.app.RecordHistory(ctx, req.ProductId, req.SkuId, req.Price, req.OldPrice, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record price history: %v", err))
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListHistory(ctx context.Context, req *pb.ListHistoryRequest) (*pb.ListHistoryResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	history, total, err := s.app.ListHistory(ctx, req.ProductId, req.SkuId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list history: %v", err))
	}

	pbHistory := make([]*pb.PriceHistory, len(history))
	for i, h := range history {
		pbHistory[i] = convertHistoryToProto(h)
	}

	return &pb.ListHistoryResponse{
		History:    pbHistory,
		TotalCount: total,
	}, nil
}

func convertRuleToProto(r *domain.PricingRule) *pb.PricingRule {
	if r == nil {
		return nil
	}
	return &pb.PricingRule{
		Id:         uint64(r.ID),
		Name:       r.Name,
		ProductId:  r.ProductID,
		SkuId:      r.SkuID,
		Strategy:   string(r.Strategy),
		BasePrice:  r.BasePrice,
		MinPrice:   r.MinPrice,
		MaxPrice:   r.MaxPrice,
		AdjustRate: r.AdjustRate,
		Enabled:    r.Enabled,
		StartTime:  timestamppb.New(r.StartTime),
		EndTime:    timestamppb.New(r.EndTime),
		CreatedAt:  timestamppb.New(r.CreatedAt),
		UpdatedAt:  timestamppb.New(r.UpdatedAt),
	}
}

func convertHistoryToProto(h *domain.PriceHistory) *pb.PriceHistory {
	if h == nil {
		return nil
	}
	return &pb.PriceHistory{
		Id:         uint64(h.ID),
		ProductId:  h.ProductID,
		SkuId:      h.SkuID,
		Price:      h.Price,
		OldPrice:   h.OldPrice,
		ChangeRate: h.ChangeRate,
		Reason:     h.Reason,
		CreatedAt:  timestamppb.New(h.CreatedAt),
		UpdatedAt:  timestamppb.New(h.UpdatedAt),
	}
}

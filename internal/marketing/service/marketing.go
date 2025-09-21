package service

import (
	"context"
	"strconv" // 用于 getUserIDFromContext
	"time"    // 用于 bizUserCouponToProto

	v1 "ecommerce/api/marketing/v1" // 修正路径
	"ecommerce/internal/marketing/biz" // 修正路径

	"google.golang.org/grpc/codes" // 添加
	"google.golang.org/grpc/metadata" // 添加
	"google.golang.org/grpc/status" // 添加
)

type MarketingService struct {
	v1.UnimplementedMarketingServer
	uc *biz.CouponUsecase
	pc *biz.PromotionUsecase // 新增：促销用例
}

func NewMarketingService(uc *biz.CouponUsecase, pc *biz.PromotionUsecase) *MarketingService {
	return &MarketingService{uc: uc, pc: pc}
}

// getUserIDFromContext 从 gRPC 上下文的 metadata 中提取用户ID。
func getUserIDFromContext(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "无法获取元数据")
	}
	// 兼容 gRPC-Gateway 在 HTTP 请求时注入的用户ID
	values := md.Get("x-md-global-user-id")
	if len(values) == 0 {
		// 兼容直接 gRPC 调用时注入的用户ID
		values = md.Get("x-user-id")
		if len(values) == 0 {
			return 0, status.Errorf(codes.Unauthenticated, "请求头中缺少 x-user-id 信息")
		}
	}
	userID, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "x-user-id 格式无效")
	}
	return userID, nil
}

func (s *MarketingService) CreateCouponTemplate(ctx context.Context, req *v1.CreateCouponTemplateRequest) (*v1.CreateCouponTemplateResponse, error) {
	bizTemplate := &biz.CouponTemplate{
		Title:               req.Title,
		Type:                int8(req.Type),
		ScopeType:           int8(req.ScopeType),
		ScopeIDs:            req.ScopeIds,
		Rules:               biz.RuleSet{Threshold: req.Rules.Threshold, Discount: req.Rules.Discount, MaxDeduction: req.Rules.MaxDeduction},
		TotalQuantity:       uint(req.TotalQuantity),
		PerUserLimit:        uint8(req.PerUserLimit),
		ValidityType:        int8(req.ValidityType),
		ValidDaysAfterClaim: uint(req.ValidDaysAfterClaim),
		Status:              int8(req.Status),
	}

	if req.ValidFrom != "" {
		t, err := time.Parse(time.RFC3339, req.ValidFrom)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid ValidFrom format: %v", err)
		}
		bizTemplate.ValidFrom = &t
	}
	if req.ValidTo != "" {
		t, err := time.Parse(time.RFC3339, req.ValidTo)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid ValidTo format: %v", err)
		}
		bizTemplate.ValidTo = &t
	}

	createdTemplate, err := s.uc.CreateCouponTemplate(ctx, bizTemplate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建优惠券模板失败: %v", err)
	}
	return &v1.CreateCouponTemplateResponse{TemplateId: createdTemplate.ID}, nil
}

func (s *MarketingService) ClaimCoupon(ctx context.Context, req *v1.ClaimCouponRequest) (*v1.ClaimCouponResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	_, err = s.uc.ClaimCoupon(ctx, userID, req.TemplateId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "领取优惠券失败: %v", err)
	}
	return &v1.ClaimCouponResponse{}, nil
}

func (s *MarketingService) CalculateDiscount(ctx context.Context, req *v1.CalculateDiscountRequest) (*v1.CalculateDiscountResponse, error) {
	bizItems := make([]*biz.OrderItemInfo, 0, len(req.Items))
	for _, item := range req.Items {
		bizItems = append(bizItems, &biz.OrderItemInfo{
			SkuID:      item.SkuId,
			SpuID:      item.SpuId,
			CategoryID: item.CategoryId,
			Price:      item.Price,
			Quantity:   item.Quantity,
		})
	}

	discount, err := s.uc.CalculateDiscount(ctx, req.UserId, req.CouponCode, bizItems)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "计算优惠金额失败: %v", err)
	}

	var totalAmount uint64
	for _, item := range req.Items {
		totalAmount += item.Price * uint64(item.Quantity)
	}

	return &v1.CalculateDiscountResponse{
		DiscountAmount: discount,
		FinalAmount:    totalAmount - discount,
	}, nil
}

// ListMyCoupons 获取用户优惠券列表。
func (s *MarketingService) ListMyCoupons(ctx context.Context, req *v1.ListMyCouponsRequest) (*v1.ListMyCouponsResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userCoupons, err := s.uc.ListUserCoupons(ctx, userID, int8(req.Status))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取我的优惠券失败: %v", err)
	}

	protoCoupons := make([]*v1.UserCouponInfo, 0, len(userCoupons))
	for _, uc := range userCoupons {
		// 获取优惠券模板信息以填充 Title
		template, err := s.uc.GetTemplateByID(ctx, uc.TemplateID)
		if err != nil {
			// 记录错误，但继续处理其他优惠券
			s.uc.Log.Warnf("failed to get coupon template for user coupon %d: %v", uc.ID, err)
			continue
		}
		protoCoupons = append(protoCoupons, bizUserCouponToProto(uc, template.Title))
	}

	return &v1.ListMyCouponsResponse{Coupons: protoCoupons}, nil
}

// bizUserCouponToProto 将 biz.UserCoupon 领域模型转换为 v1.UserCouponInfo API 模型。
func bizUserCouponToProto(uc *biz.UserCoupon, templateTitle string) *v1.UserCouponInfo {
	if uc == nil {
		return nil
	}
	return &v1.UserCouponInfo{
		Id:         uc.ID,
		TemplateId: uc.TemplateID,
		UserId:     uc.UserID,
		CouponCode: uc.CouponCode,
		Status:     int32(uc.Status),
		ValidFrom:  uc.ValidFrom.Format(time.RFC3339),
		ValidTo:    uc.ValidTo.Format(time.RFC3339),
		Title:      templateTitle, // 填充 Title
	}
}

// CreatePromotion 实现创建促销接口
func (s *MarketingService) CreatePromotion(ctx context.Context, req *v1.CreatePromotionRequest) (*v1.PromotionInfo, error) {
	promotion := &biz.Promotion{
		Name:           req.Name,
		Type:           int8(req.Type),
		Description:    req.Description,
		ProductIDs:     req.ProductIds,
		DiscountValue:  req.GetDiscountValue(),
		MinOrderAmount: req.GetMinOrderAmount(),
	}
	if req.StartTime != "" {
		t, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid StartTime format: %v", err)
		}
		promotion.StartTime = &t
	}
	if req.EndTime != "" {
		t, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid EndTime format: %v", err)
		}
		promotion.EndTime = &t
	}

	createdPromotion, err := s.pc.CreatePromotion(ctx, promotion)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建促销失败: %v", err)
	}
	return bizPromotionToProto(createdPromotion), nil
}

// UpdatePromotion 实现更新促销接口
func (s *MarketingService) UpdatePromotion(ctx context.Context, req *v1.UpdatePromotionRequest) (*v1.PromotionInfo, error) {
	promotion := &biz.Promotion{
		ID: req.Id,
	}
	if req.HasName() {
		promotion.Name = req.GetName()
	}
	if req.HasType() {
		promotion.Type = int8(req.GetType())
	}
	if req.HasDescription() {
		promotion.Description = req.GetDescription()
	}
	if req.StartTime != "" {
		t, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid StartTime format: %v", err)
		}
		promotion.StartTime = &t
	}
	if req.EndTime != "" {
		t, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid EndTime format: %v", err)
		}
		promotion.EndTime = &t
	}
	if len(req.ProductIds) > 0 {
		promotion.ProductIDs = req.ProductIds
	}
	if req.HasDiscountValue() {
		promotion.DiscountValue = req.GetDiscountValue()
	}
	if req.HasMinOrderAmount() {
		promotion.MinOrderAmount = req.GetMinOrderAmount()
	}
	if req.HasStatus() {
		promotion.Status = int8(req.GetStatus())
	}

	updatedPromotion, err := s.pc.UpdatePromotion(ctx, promotion)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "更新促销失败: %v", err)
	}
	return bizPromotionToProto(updatedPromotion), nil
}

// DeletePromotion 实现删除促销接口
func (s *MarketingService) DeletePromotion(ctx context.Context, req *v1.DeletePromotionRequest) (*emptypb.Empty, error) {
	err := s.pc.DeletePromotion(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "删除促销失败: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// GetPromotion 实现获取促销详情接口
func (s *MarketingService) GetPromotion(ctx context.Context, req *v1.GetPromotionRequest) (*v1.PromotionInfo, error) {
	promotion, err := s.pc.GetPromotion(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取促销详情失败: %v", err)
	}
	return bizPromotionToProto(promotion), nil
}

// ListPromotions 实现获取促销列表接口
func (s *MarketingService) ListPromotions(ctx context.Context, req *v1.ListPromotionsRequest) (*v1.ListPromotionsResponse, error) {
	promotions, total, err := s.pc.ListPromotions(ctx, req.PageSize, req.PageNum, req.GetName(), req.GetType(), req.GetStatus())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取促销列表失败: %v", err)
	}
	var protoPromotions []*v1.PromotionInfo
	for _, p := range promotions {
		protoPromotions = append(protoPromotions, bizPromotionToProto(p))
	}
	return &v1.ListPromotionsResponse{Promotions: protoPromotions, TotalCount: total}, nil
}

// bizPromotionToProto 将 biz.Promotion 领域模型转换为 v1.PromotionInfo API 模型。
func bizPromotionToProto(p *biz.Promotion) *v1.PromotionInfo {
	if p == nil {
		return nil
	}
	res := &v1.PromotionInfo{
		Id:             p.ID,
		Name:           p.Name,
		Type:           uint32(p.Type),
		Description:    p.Description,
		ProductIds:     p.ProductIDs,
		DiscountValue:  p.DiscountValue,
		MinOrderAmount: p.MinOrderAmount,
	}
	if p.StartTime != nil {
		res.StartTime = p.StartTime.Format(time.RFC3339)
	}
	if p.EndTime != nil {
		res.EndTime = p.EndTime.Format(time.RFC3339)
	}
	if p.Status != nil {
		res.Status = uint32(*p.Status)
	}
	return res
}

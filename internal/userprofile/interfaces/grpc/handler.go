//go:build ignore

package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/userprofile/v1"
	"github.com/wyfcoding/ecommerce/internal/userprofile/application"
	"github.com/wyfcoding/ecommerce/internal/userprofile/domain"
)

type UserProfileHandler struct {
	pb.UnimplementedUserProfileServiceServer
	commandService *application.ProfileCommandService
	queryService   *application.ProfileQueryService
}

func NewUserProfileHandler(
	commandService *application.ProfileCommandService,
	queryService *application.ProfileQueryService,
) *UserProfileHandler {
	return &UserProfileHandler{
		commandService: commandService,
		queryService:   queryService,
	}
}

func (h *UserProfileHandler) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.CreateProfileResponse, error) {
	profile, err := h.commandService.CreateProfile(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.CreateProfileResponse{
		ProfileId: profile.ID,
		UserId:    profile.UserID,
		Status:    profile.Status.String(),
	}, nil
}

func (h *UserProfileHandler) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	profile, err := h.queryService.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetProfileResponse{
		Profile: h.toProfileProto(profile),
	}, nil
}

func (h *UserProfileHandler) GetProfileSummary(ctx context.Context, req *pb.GetProfileSummaryRequest) (*pb.GetProfileSummaryResponse, error) {
	summary, err := h.queryService.GetProfileSummary(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetProfileSummaryResponse{
		UserId:              summary.UserID,
		Status:              summary.Status,
		OverallScore:        int32(summary.OverallScore),
		ActivityScore:       int32(summary.ActivityScore),
		EngagementScore:     int32(summary.EngagementScore),
		ValueScore:          int32(summary.ValueScore),
		LoyaltyScore:        int32(summary.LoyaltyScore),
		ProfileCompleteness: int32(summary.ProfileCompleteness),
		SpendingLevel:       summary.SpendingLevel,
		ValueSegment:        summary.ValueSegment,
		TopCategories:       summary.TopCategories,
		TopBrands:           summary.TopBrands,
		ChurnRisk:           summary.ChurnRisk,
		LastActiveAt:        summary.LastActiveAt,
	}, nil
}

func (h *UserProfileHandler) AddTag(ctx context.Context, req *pb.AddTagRequest) (*pb.AddTagResponse, error) {
	category := domain.TagCategory(req.Category)
	source := domain.TagSource(req.Source)

	if err := h.commandService.AddTag(ctx, req.UserId, req.TagKey, req.TagValue, category, source, req.Confidence); err != nil {
		return nil, err
	}

	return &pb.AddTagResponse{
		Success: true,
	}, nil
}

func (h *UserProfileHandler) RemoveTag(ctx context.Context, req *pb.RemoveTagRequest) (*pb.RemoveTagResponse, error) {
	if err := h.commandService.RemoveTag(ctx, req.UserId, req.TagKey); err != nil {
		return nil, err
	}

	return &pb.RemoveTagResponse{
		Success: true,
	}, nil
}

func (h *UserProfileHandler) GetTags(ctx context.Context, req *pb.GetTagsRequest) (*pb.GetTagsResponse, error) {
	tags, err := h.queryService.GetTags(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	pbTags := make([]*pb.UserTag, len(tags))
	for i, tag := range tags {
		pbTags[i] = &pb.UserTag{
			TagKey:     tag.TagKey,
			TagValue:   tag.TagValue,
			Category:   int32(tag.Category),
			Source:     int32(tag.Source),
			Confidence: tag.Confidence,
		}
	}

	return &pb.GetTagsResponse{
		Tags: pbTags,
	}, nil
}

func (h *UserProfileHandler) RecordBehavior(ctx context.Context, req *pb.RecordBehaviorRequest) (*pb.RecordBehaviorResponse, error) {
	behaviorType := domain.BehaviorType(req.BehaviorType)

	if err := h.commandService.RecordBehavior(ctx, req.UserId, behaviorType, req.TargetType, req.TargetId, req.Value, req.Duration); err != nil {
		return nil, err
	}

	return &pb.RecordBehaviorResponse{
		Success: true,
	}, nil
}

func (h *UserProfileHandler) GetBehaviorFeatures(ctx context.Context, req *pb.GetBehaviorFeaturesRequest) (*pb.GetBehaviorFeaturesResponse, error) {
	features, err := h.queryService.GetBehaviorFeatures(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetBehaviorFeaturesResponse{
		BrowseCount:        features.BrowseCount,
		SearchCount:        features.SearchCount,
		PurchaseCount:      features.PurchaseCount,
		ActivityScore:      int32(features.ActivityScore),
		EngagementScore:    int32(features.EngagementScore),
		ConversionRate:     features.ConversionRate,
		ReturnRate:         features.ReturnRate,
		RepeatPurchaseRate: features.RepeatPurchaseRate,
	}, nil
}

func (h *UserProfileHandler) GetPreferences(ctx context.Context, req *pb.GetPreferencesRequest) (*pb.GetPreferencesResponse, error) {
	preferences, err := h.queryService.GetPreferences(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	topCategories := preferences.GetTopCategories(10)
	pbCategories := make([]*pb.CategoryPreference, len(topCategories))
	for i, cat := range topCategories {
		pbCategories[i] = &pb.CategoryPreference{
			CategoryId:    cat.CategoryID,
			CategoryName:  cat.CategoryName,
			Score:         cat.Score,
			ViewCount:     cat.ViewCount,
			PurchaseCount: cat.PurchaseCount,
		}
	}

	topBrands := preferences.GetTopBrands(10)
	pbBrands := make([]*pb.BrandPreference, len(topBrands))
	for i, brand := range topBrands {
		pbBrands[i] = &pb.BrandPreference{
			BrandId:       brand.BrandID,
			BrandName:     brand.BrandName,
			Score:         brand.Score,
			ViewCount:     brand.ViewCount,
			PurchaseCount: brand.PurchaseCount,
		}
	}

	return &pb.GetPreferencesResponse{
		Categories: pbCategories,
		Brands:     pbBrands,
	}, nil
}

func (h *UserProfileHandler) GetConsumptionProfile(ctx context.Context, req *pb.GetConsumptionProfileRequest) (*pb.GetConsumptionProfileResponse, error) {
	consumption, err := h.queryService.GetConsumptionProfile(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetConsumptionProfileResponse{
		TotalSpent:            consumption.TotalSpent,
		TotalOrders:           consumption.TotalOrders,
		AvgOrderValue:         consumption.AvgOrderValue,
		SpendingLevel:         consumption.SpendingLevel.String(),
		ConsumptionFrequency:  consumption.ConsumptionFrequency.String(),
		PredictedMonthlySpend: consumption.PredictedMonthlySpend,
		PredictedYearlySpend:  consumption.PredictedYearlySpend,
		Ltv:                   consumption.LTV,
		SpendingTrend:         consumption.SpendingTrend.String(),
		GrowthRate:            consumption.GrowthRate,
		ChurnRisk:             consumption.ChurnRisk,
		ValueSegment:          consumption.ValueSegment.String(),
	}, nil
}

func (h *UserProfileHandler) RecalculateProfile(ctx context.Context, req *pb.RecalculateProfileRequest) (*pb.RecalculateProfileResponse, error) {
	if err := h.commandService.RecalculateProfile(ctx, req.UserId); err != nil {
		return nil, err
	}

	return &pb.RecalculateProfileResponse{
		Success: true,
	}, nil
}

func (h *UserProfileHandler) GetUsersByTag(ctx context.Context, req *pb.GetUsersByTagRequest) (*pb.GetUsersByTagResponse, error) {
	profiles, err := h.queryService.GetUsersByTag(ctx, req.TagKey, req.TagValue, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}

	userIds := make([]uint64, len(profiles))
	for i, p := range profiles {
		userIds[i] = p.UserID
	}

	return &pb.GetUsersByTagResponse{
		UserIds: userIds,
		Total:   int32(len(profiles)),
	}, nil
}

func (h *UserProfileHandler) GetUsersBySegment(ctx context.Context, req *pb.GetUsersBySegmentRequest) (*pb.GetUsersBySegmentResponse, error) {
	profiles, err := h.queryService.GetSegmentUsers(ctx, req.SegmentNo, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}

	userIds := make([]uint64, len(profiles))
	for i, p := range profiles {
		userIds[i] = p.UserID
	}

	return &pb.GetUsersBySegmentResponse{
		UserIds: userIds,
		Total:   int32(len(profiles)),
	}, nil
}

func (h *UserProfileHandler) toProfileProto(profile *domain.UserProfile) *pb.UserProfile {
	pbProfile := &pb.UserProfile{
		Id:                  profile.ID,
		UserId:              profile.UserID,
		Status:              profile.Status.String(),
		ProfileVersion:      int32(profile.ProfileVersion),
		ActivityScore:       int32(profile.ActivityScore),
		EngagementScore:     int32(profile.EngagementScore),
		ValueScore:          int32(profile.ValueScore),
		LoyaltyScore:        int32(profile.LoyaltyScore),
		OverallScore:        int32(profile.OverallScore),
		ProfileCompleteness: int32(profile.ProfileCompleteness),
	}

	if profile.LastCalculatedAt != nil {
		pbProfile.LastCalculatedAt = profile.LastCalculatedAt.Unix()
	}

	if profile.LastActiveAt != nil {
		pbProfile.LastActiveAt = profile.LastActiveAt.Unix()
	}

	return pbProfile
}

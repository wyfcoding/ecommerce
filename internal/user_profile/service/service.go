package service

import (
	"context"
	"errors"

	v1 "ecommerce/api/user_profile/v1"
	"ecommerce/internal/user_profile/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserProfileService is the gRPC service implementation for user profiles.
type UserProfileService struct {
	v1.UnimplementedUserProfileServiceServer
	uc *biz.UserProfileUsecase
}

// NewUserProfileService creates a new UserProfileService.
func NewUserProfileService(uc *biz.UserProfileUsecase) *UserProfileService {
	return &UserProfileService{uc: uc}
}

// bizUserProfileToProto converts biz.UserProfile to v1.UserProfile.
func bizUserProfileToProto(profile *biz.UserProfile) *v1.UserProfile {
	if profile == nil {
		return nil
	}
	return &v1.UserProfile{
		UserId:           profile.UserID,
		Gender:           profile.Gender,
		AgeGroup:         profile.AgeGroup,
		City:             profile.City,
		Interests:        profile.Interests,
		RecentCategories: profile.RecentCategories,
		RecentBrands:     profile.RecentBrands,
		TotalSpent:       profile.TotalSpent,
		OrderCount:       profile.OrderCount,
		LastActiveTime:   timestamppb.New(profile.LastActiveTime),
		CustomTags:       profile.CustomTags,
	}
}

// GetUserProfile implements the GetUserProfile RPC.
func (s *UserProfileService) GetUserProfile(ctx context.Context, req *v1.GetUserProfileRequest) (*v1.UserProfile, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	profile, err := s.uc.GetUserProfile(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, biz.ErrUserProfileNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to get user profile: %v", err)
	}

	return bizUserProfileToProto(profile), nil
}

// UpdateUserProfile implements the UpdateUserProfile RPC.
func (s *UserProfileService) UpdateUserProfile(ctx context.Context, req *v1.UpdateUserProfileRequest) (*v1.UserProfile, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	profile := &biz.UserProfile{
		UserID: req.UserId,
	}
	if req.HasGender() {
		profile.Gender = req.GetGender()
	}
	if req.HasAgeGroup() {
		profile.AgeGroup = req.GetAgeGroup()
	}
	if req.HasCity() {
		profile.City = req.GetCity()
	}
	if len(req.Interests) > 0 {
		profile.Interests = req.Interests
	}
	if len(req.RecentCategories) > 0 {
		profile.RecentCategories = req.RecentCategories
	}
	if len(req.RecentBrands) > 0 {
		profile.RecentBrands = req.RecentBrands
	}
	if req.HasTotalSpent() {
		profile.TotalSpent = req.GetTotalSpent()
	}
	if req.HasOrderCount() {
		profile.OrderCount = req.GetOrderCount()
	}
	if req.HasLastActiveTime() {
		profile.LastActiveTime = req.GetLastActiveTime().AsTime()
	}
	if len(req.CustomTags) > 0 {
		profile.CustomTags = req.CustomTags
	}

	updatedProfile, err := s.uc.UpdateUserProfile(ctx, profile)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user profile: %v", err)
	}

	return bizUserProfileToProto(updatedProfile), nil
}

// RecordUserBehavior implements the RecordUserBehavior RPC.
func (s *UserProfileService) RecordUserBehavior(ctx context.Context, req *v1.RecordUserBehaviorRequest) (*v1.RecordUserBehaviorResponse, error) {
	if req.UserId == "" || req.BehaviorType == "" || req.ItemId == "" || req.EventTime == nil {
		return nil, status.Error(codes.InvalidArgument, "user_id, behavior_type, item_id, and event_time are required")
	}

	properties := make(map[string]string)
	for k, v := range req.Properties {
		properties[k] = v
	}

	err := s.uc.RecordUserBehavior(ctx, req.UserId, req.BehaviorType, req.ItemId, properties, req.EventTime.AsTime())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to record user behavior: %v", err)
	}

	return &v1.RecordUserBehaviorResponse{Success: true, Message: "Behavior recorded successfully"}, nil
}

package client

import (
	"context"
	"fmt"

	userprofilev1 "ecommerce/api/user_profile/v1"
	"ecommerce/internal/realtime_analytics/model"
	"google.golang.org/grpc"
)

// UserProfileClient defines the interface to interact with the User Profile Service.
type UserProfileClient interface {
	UpdateUserProfile(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error)
	// GetUserProfile(ctx context.Context, userID string) (*UserProfile, error) // If needed
}

type userProfileClient struct {
	client userprofilev1.UserProfileServiceClient
}

func NewUserProfileClient(conn *grpc.ClientConn) UserProfileClient {
	return &userProfileClient{
		client: userprofilev1.NewUserProfileServiceClient(conn),
	}
}

func (c *userProfileClient) UpdateUserProfile(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error) {
	req := &userprofilev1.UpdateUserProfileRequest{
		UserId:         profile.UserID,
		LastActiveTime: &userprofilev1.Timestamp{Seconds: profile.LastActiveTime.Unix()},
		// ... map other fields
	}
	res, err := c.client.UpdateUserProfile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}
	return &model.UserProfile{UserID: res.GetProfile().GetUserId()}, nil
}

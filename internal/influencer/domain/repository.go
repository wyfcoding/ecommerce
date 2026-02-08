package domain

import (
	"context"
)

type InfluencerRepository interface {
	SaveInfluencer(ctx context.Context, i *Influencer) error
	GetInfluencer(ctx context.Context, id string) (*Influencer, error)
	GetInfluencerByUserID(ctx context.Context, userID string) (*Influencer, error)

	SaveCampaign(ctx context.Context, c *Campaign) error
	ListCampaigns(ctx context.Context, influencerID string) ([]*Campaign, error)
}

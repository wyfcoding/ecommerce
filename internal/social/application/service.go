package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/social/domain"
)
type SocialService struct { repo domain.SocialRepository }
func (s *SocialService) GroupBuy(ctx context.Context, id string) error { return nil }

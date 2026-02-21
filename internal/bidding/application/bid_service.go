package application

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/bidding/domain"
	"github.com/wyfcoding/pkg/xerrors"
)

// bidLuaScript 确保竞价的原子性校验。
const bidLuaScript = `
local current_price = tonumber(redis.call('get', KEYS[1]) or '0')
local new_bid = tonumber(ARGV[1])
if new_bid > current_price then
    redis.call('set', KEYS[1], ARGV[1])
    return 1
else
    return 0
end
`

type BiddingService struct {
	repo  domain.AuctionRepository
	cache redis.UniversalClient
}

func NewBiddingService(repo domain.AuctionRepository, cache redis.UniversalClient) *BiddingService {
	return &BiddingService{repo: repo, cache: cache}
}

func (s *BiddingService) Bid(ctx context.Context, auctionID uint64, userID uint64, amount decimal.Decimal) error {
	if s.cache == nil {
		return xerrors.InvalidArg("redis cache client is required")
	}

	// 1. Redis 原子性预扣/校验 (解决高并发下的无效出价)
	key := fmt.Sprintf("auction:price:%d", auctionID)
	res, err := s.cache.Eval(ctx, bidLuaScript, []string{key}, amount.String()).Result()
	if err != nil || res.(int64) == 0 {
		return xerrors.InvalidArg("bid too low or system busy")
	}

	// 2. 加锁更新 DB 聚合
	auction, err := s.repo.FindByIDForUpdate(ctx, auctionID)
	if err != nil {
		return xerrors.NotFound("auction not found")
	}

	if _, err := auction.PlaceBid(time.Now(), userID, amount); err != nil {
		return xerrors.InvalidArg(err.Error())
	}

	return s.repo.Save(ctx, auction)
}

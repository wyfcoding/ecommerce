// 变更说明：
// 1. 【核心机制】补全英式拍卖(English Auction)的核心出价递增验证、加价阶梯(Bid Step)。
// 2. 【防狙击】实现防狙击(Anti-Snipe)自动延时拓展：拍卖行倒计时防秒杀机制。
// 3. 【代理出价】引入自动代理出价(Proxy Bidding)逻辑，只要不超 MaxBid，系统自动代为还击出价。
// 4. 【高并发】状态机防超售与版本乐观锁，保证最终只有一个竞拍者胜出。
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrAuctionNotStarted   = errors.New("auction has not started yet")
	ErrAuctionEnded        = errors.New("auction has already ended")
	ErrBidTooLow           = errors.New("bid amount is too low based on current price and bid step")
	ErrSelfOutbid          = errors.New("you are already the highest bidder")
	ErrMaxBidLowerThanCurrent = errors.New("your maximum bid must be higher than current proxy price")
)

type AuctionStatus string

const (
	AuctionStatusDraft    AuctionStatus = "DRAFT"
	AuctionStatusActive   AuctionStatus = "ACTIVE"
	AuctionStatusEnded    AuctionStatus = "ENDED"
	AuctionStatusCanceled AuctionStatus = "CANCELLED"
	AuctionStatusPaid     AuctionStatus = "PAID"
	AuctionStatusFailed   AuctionStatus = "FAILED" // 流流标或弃标违约
)

// Auction 竞拍活动聚合根
type Auction struct {
	ID                uint64          `json:"id"`
	ProductID         uint64          `json:"product_id"`
	StartingPrice     decimal.Decimal `json:"starting_price"`
	ReservePrice      decimal.Decimal `json:"reserve_price"` // 保留价，低于此价格流标
	CurrentPrice      decimal.Decimal `json:"current_price"`
	CurrentWinnerID   uint64          `json:"current_winner_id"`
	BidIncrement      decimal.Decimal `json:"bid_increment"` // 最小加价阶梯
	DepositAmount     decimal.Decimal `json:"deposit_amount"` // 保证金要求
	StartTime         time.Time       `json:"start_time"`
	EndTime           time.Time       `json:"end_time"`
	AntiSnipeDuration time.Duration   `json:"anti_snipe_duration"` // 触发防狙击时的延时时长 (如 5 分钟)
	Status            AuctionStatus   `json:"status"`
	Version           int64           `json:"version"` // 乐观锁
}

// BidRecord 单个出价记录
type BidRecord struct {
	ID        uint64          `json:"id"`
	AuctionID uint64          `json:"auction_id"`
	UserID    uint64          `json:"user_id"`
	Amount    decimal.Decimal `json:"amount"` // 当次出价额
	IsProxy   bool            `json:"is_proxy"` // 是否机器代理出价生成
	Timestamp time.Time       `json:"timestamp"`
}

// ProxyBid 用户的自动代理最高出价指示
type ProxyBid struct {
	AuctionID uint64
	UserID    uint64
	MaxAmount decimal.Decimal // 用户能接受的最高价
}

// PlaceBid 执行正常的公开加价
func (a *Auction) PlaceBid(now time.Time, userID uint64, amount decimal.Decimal) (*BidRecord, error) {
	if a.Status != AuctionStatusActive {
		return nil, errors.New("auction is not active")
	}
	if now.Before(a.StartTime) {
		return nil, ErrAuctionNotStarted
	}
	if now.After(a.EndTime) {
		return nil, ErrAuctionEnded
	}

	// 如果他已经是最高出价者，不允许抬价自己
	if a.CurrentWinnerID == userID {
		return nil, ErrSelfOutbid
	}

	// 判断最低接受报价 = 现价 + 加价梯度
	// (或者首单：>= 起拍价)
	minRequired := a.CurrentPrice.Add(a.BidIncrement)
	if a.CurrentPrice.IsZero() || a.CurrentPrice.LessThan(a.StartingPrice) {
		minRequired = a.StartingPrice
	}

	if amount.LessThan(minRequired) {
		return nil, ErrBidTooLow
	}

	// 更新现价与赢家
	a.CurrentPrice = amount
	a.CurrentWinnerID = userID

	record := &BidRecord{
		AuctionID: a.ID,
		UserID:    userID,
		Amount:    amount,
		IsProxy:   false,
		Timestamp: now,
	}

	// 执行防狙击 (Anti-Sniper) 机制: 
	// 如果剩余时间少于防狙击时长，自动无限制延长 EndTime
	timeRemaining := a.EndTime.Sub(now)
	if a.AntiSnipeDuration > 0 && timeRemaining < a.AntiSnipeDuration {
		a.EndTime = a.EndTime.Add(a.AntiSnipeDuration - timeRemaining)
	}

	a.Version++
	return record, nil
}

// HandleProxyBid 处理多个用户的秘密代理自动最高价竞争
// 这是一个经典的英式代理算法（类似 eBay 的自动出价）。
// 谁的 MaxBid 高，谁就维持领先，且价格停留在【次高 MaxBid + 1 阶梯】的位置。
func (a *Auction) HandleProxyBid(now time.Time, newProxy ProxyBid, currentWinnerProxy *ProxyBid) (*BidRecord, error) {
	// 校验新代理限额是否低于现价+阶梯
	minRequired := a.CurrentPrice.Add(a.BidIncrement)
	if newProxy.MaxAmount.LessThan(minRequired) {
		return nil, ErrMaxBidLowerThanCurrent
	}

	if currentWinnerProxy == nil || currentWinnerProxy.UserID == newProxy.UserID {
		// 没有对手（或者就是自己）：只需让新用户占领，以首单价/加阶价拿住，但记录下其 Max 守护价限制
		return a.PlaceBid(now, newProxy.UserID, minRequired)
	}

	// 两虎相争：对比两者 Max
	if newProxy.MaxAmount.GreaterThan(currentWinnerProxy.MaxAmount) {
		// NewProxy 胜出
		// 新领先价格 = currentWinnerMax + 1 阶梯，如果超出 newProxyMax，就停在 newProxyMax
		newLeadingPrice := currentWinnerProxy.MaxAmount.Add(a.BidIncrement)
		if newLeadingPrice.GreaterThan(newProxy.MaxAmount) {
			newLeadingPrice = newProxy.MaxAmount
		}
		// 记录为代理打出，产生一条新的 BidRecord
		a.CurrentPrice = newLeadingPrice
		a.CurrentWinnerID = newProxy.UserID
		
		a.AntiSnipeCheck(now)
		a.Version++

		return &BidRecord{
			AuctionID: a.ID,
			UserID:    newProxy.UserID,
			Amount:    newLeadingPrice,
			IsProxy:   true,
			Timestamp: now,
		}, nil

	} else {
		// CurrentWinner 保住胜利
		// 竞争价格被顶高到 newProxyMax + 1 阶梯
		newLeadingPrice := newProxy.MaxAmount.Add(a.BidIncrement)
		if newLeadingPrice.GreaterThan(currentWinnerProxy.MaxAmount) {
			newLeadingPrice = currentWinnerProxy.MaxAmount
		}

		a.CurrentPrice = newLeadingPrice
		// 胜利者不变，只更新价格
		
		a.AntiSnipeCheck(now)
		a.Version++

		return &BidRecord{
			AuctionID: a.ID,
			UserID:    currentWinnerProxy.UserID,
			Amount:    newLeadingPrice,
			IsProxy:   true, // 老赢家的代理系统反击
			Timestamp: now,
		}, nil
	}
}

// AntiSnipeCheck 防狙击延时抽取辅助
func (a *Auction) AntiSnipeCheck(now time.Time) {
	timeRemaining := a.EndTime.Sub(now)
	if a.AntiSnipeDuration > 0 && timeRemaining < a.AntiSnipeDuration {
		a.EndTime = a.EndTime.Add(a.AntiSnipeDuration - timeRemaining)
	}
}

// EndAuction 判定拍卖结束。
func (a *Auction) EndAuction(now time.Time) {
	if now.Before(a.EndTime) { return }

	if a.CurrentWinnerID == 0 || a.CurrentPrice.LessThan(a.ReservePrice) {
		// 没人出价，或低于保留价
		a.Status = AuctionStatusFailed
	} else {
		a.Status = AuctionStatusEnded
	}
	a.Version++
}

// AuctionRepository 竞拍仓储
type AuctionRepository interface {
	Save(ctx context.Context, auction *Auction) error
	FindByIDForUpdate(ctx context.Context, id uint64) (*Auction, error) // Select ... For Update 悲观锁保证并行出价的互斥
	FindActiveAuctionsEndBefore(ctx context.Context, t time.Time) ([]*Auction, error)
	// Proxy 相关的记录持久化
	SaveProxyBid(ctx context.Context, proxy *ProxyBid) error
	GetHighestProxyBid(ctx context.Context, auctionID uint64) (*ProxyBid, error)
}

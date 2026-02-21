package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/bidding/domain"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/server"
)

type bidRecordView struct {
	AuctionID uint64 `json:"auction_id"`
	UserID    uint64 `json:"user_id"`
	Amount    string `json:"amount"`
	IsProxy   bool   `json:"is_proxy"`
	Timestamp string `json:"timestamp"`
}

type auctionStore struct {
	mu       sync.RWMutex
	auctions map[uint64]*domain.Auction
	records  map[uint64][]*domain.BidRecord
}

func newAuctionStore() *auctionStore {
	return &auctionStore{
		auctions: make(map[uint64]*domain.Auction),
		records:  make(map[uint64][]*domain.BidRecord),
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	store := newAuctionStore()
	engine := server.NewDefaultGinEngine(gin.Recovery())
	v1 := engine.Group("/api/v1/bidding")
	{
		v1.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"status": "ok"})
		})

		v1.POST("/auctions", func(c *gin.Context) {
			var req struct {
				ProductID     uint64 `json:"product_id"`
				StartingPrice string `json:"starting_price"`
				ReservePrice  string `json:"reserve_price"`
				BidIncrement  string `json:"bid_increment"`
				DurationMin   int    `json:"duration_min"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", err.Error())
				return
			}
			if req.ProductID == 0 || req.StartingPrice == "" || req.BidIncrement == "" {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", "product_id/starting_price/bid_increment are required")
				return
			}
			startingPrice, err := decimal.NewFromString(req.StartingPrice)
			if err != nil || !startingPrice.IsPositive() {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid starting_price", "starting_price must be positive decimal")
				return
			}
			reservePrice := startingPrice
			if req.ReservePrice != "" {
				reservePrice, err = decimal.NewFromString(req.ReservePrice)
				if err != nil || reservePrice.IsNegative() {
					response.ErrorWithStatus(c, http.StatusBadRequest, "invalid reserve_price", "reserve_price must be non-negative decimal")
					return
				}
			}
			increment, err := decimal.NewFromString(req.BidIncrement)
			if err != nil || !increment.IsPositive() {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid bid_increment", "bid_increment must be positive decimal")
				return
			}
			durationMin := req.DurationMin
			if durationMin <= 0 {
				durationMin = 30
			}
			now := time.Now().UTC()
			auctionID := idgen.GenID()
			auction := &domain.Auction{
				ID:                auctionID,
				ProductID:         req.ProductID,
				StartingPrice:     startingPrice,
				ReservePrice:      reservePrice,
				CurrentPrice:      decimal.Zero,
				CurrentWinnerID:   0,
				BidIncrement:      increment,
				DepositAmount:     decimal.Zero,
				StartTime:         now,
				EndTime:           now.Add(time.Duration(durationMin) * time.Minute),
				AntiSnipeDuration: 2 * time.Minute,
				Status:            domain.AuctionStatusActive,
				Version:           1,
			}
			store.mu.Lock()
			store.auctions[auctionID] = auction
			store.records[auctionID] = []*domain.BidRecord{}
			store.mu.Unlock()
			response.Success(c, auction)
		})

		v1.POST("/auctions/:id/bids", func(c *gin.Context) {
			auctionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil || auctionID == 0 {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "auction id must be unsigned integer")
				return
			}
			var req struct {
				UserID uint64 `json:"user_id"`
				Amount string `json:"amount"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", err.Error())
				return
			}
			if req.UserID == 0 || req.Amount == "" {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", "user_id/amount are required")
				return
			}
			amount, err := decimal.NewFromString(req.Amount)
			if err != nil || !amount.IsPositive() {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid amount", "amount must be positive decimal")
				return
			}

			store.mu.Lock()
			defer store.mu.Unlock()
			auction, ok := store.auctions[auctionID]
			if !ok {
				response.ErrorWithStatus(c, http.StatusNotFound, "not found", "auction not found")
				return
			}
			record, err := auction.PlaceBid(time.Now().UTC(), req.UserID, amount)
			if err != nil {
				response.ErrorWithStatus(c, http.StatusConflict, "bid rejected", err.Error())
				return
			}
			store.records[auctionID] = append(store.records[auctionID], record)
			response.Success(c, record)
		})

		v1.GET("/auctions/:id", func(c *gin.Context) {
			auctionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil || auctionID == 0 {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "auction id must be unsigned integer")
				return
			}
			store.mu.RLock()
			auction, ok := store.auctions[auctionID]
			records := store.records[auctionID]
			store.mu.RUnlock()
			if !ok {
				response.ErrorWithStatus(c, http.StatusNotFound, "not found", "auction not found")
				return
			}
			views := make([]*bidRecordView, 0, len(records))
			for _, r := range records {
				views = append(views, &bidRecordView{
					AuctionID: r.AuctionID,
					UserID:    r.UserID,
					Amount:    r.Amount.String(),
					IsProxy:   r.IsProxy,
					Timestamp: r.Timestamp.Format(time.RFC3339),
				})
			}
			response.Success(c, gin.H{"auction": auction, "records": views})
		})
	}

	addr := envOrDefault("BIDDING_HTTP_ADDR", ":9202")
	srv := server.NewGinServer(engine, addr, logger)
	go func() {
		if err := srv.Start(context.Background()); err != nil {
			slog.Error("server exit", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = srv.Stop(context.Background())
	slog.Info("service bidding gracefully stopped")
}

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

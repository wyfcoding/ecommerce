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

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/supplychainfinance/domain"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/server"
)

type scfStore struct {
	mu           sync.RWMutex
	applications map[string]*domain.FinanceApplication
	creditLines  map[string]*domain.CreditLine
}

func newSCFStore() *scfStore {
	return &scfStore{applications: make(map[string]*domain.FinanceApplication), creditLines: make(map[string]*domain.CreditLine)}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	store := newSCFStore()
	engine := server.NewDefaultGinEngine(gin.Recovery())
	v1 := engine.Group("/api/v1/supplychainfinance")
	{
		v1.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"status": "ok"})
		})

		v1.POST("/applications/apply", func(c *gin.Context) {
			var req struct {
				ApplicantID     string `json:"applicant_id"`
				ApplicantName   string `json:"applicant_name"`
				FinanceType     string `json:"finance_type"`
				RequestedAmount string `json:"requested_amount"`
				Currency        string `json:"currency"`
				TermDays        int    `json:"term_days"`
				Purpose         string `json:"purpose"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", err.Error())
				return
			}
			if req.ApplicantID == "" || req.ApplicantName == "" || req.RequestedAmount == "" || req.FinanceType == "" {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", "applicant_id/applicant_name/finance_type/requested_amount are required")
				return
			}
			amount, err := decimal.NewFromString(req.RequestedAmount)
			if err != nil || !amount.IsPositive() {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid requested_amount", "requested_amount must be positive decimal")
				return
			}
			termDays := req.TermDays
			if termDays <= 0 {
				termDays = 30
			}
			app := domain.NewFinanceApplication(req.ApplicantID, req.ApplicantName, domain.FinanceType(req.FinanceType), amount, nonEmpty(req.Currency, "CNY"), termDays, req.Purpose)
			app.ID = "FA" + strconv.FormatUint(idgen.GenID(), 10)
			app.Submit()
			store.mu.Lock()
			store.applications[app.ID] = app
			store.mu.Unlock()
			response.Success(c, app)
		})

		v1.POST("/applications/:id/approve", func(c *gin.Context) {
			id := c.Param("id")
			var req struct {
				ApprovedAmount string `json:"approved_amount"`
				InterestRate   string `json:"interest_rate"`
				FeeAmount      string `json:"fee_amount"`
				Operator       string `json:"operator"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", err.Error())
				return
			}
			approvedAmount, err := decimal.NewFromString(req.ApprovedAmount)
			if err != nil || !approvedAmount.IsPositive() {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid approved_amount", "approved_amount must be positive decimal")
				return
			}
			interestRate, err := decimal.NewFromString(nonEmpty(req.InterestRate, "0.08"))
			if err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid interest_rate", err.Error())
				return
			}
			feeAmount, err := decimal.NewFromString(nonEmpty(req.FeeAmount, "0"))
			if err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid fee_amount", err.Error())
				return
			}
			store.mu.Lock()
			app, ok := store.applications[id]
			if ok {
				app.Approve(approvedAmount, interestRate, feeAmount, nonEmpty(req.Operator, "system"))
			}
			store.mu.Unlock()
			if !ok {
				response.ErrorWithStatus(c, http.StatusNotFound, "not found", "application not found")
				return
			}
			response.Success(c, app)
		})

		v1.GET("/applications/:id", func(c *gin.Context) {
			id := c.Param("id")
			store.mu.RLock()
			app, ok := store.applications[id]
			store.mu.RUnlock()
			if !ok {
				response.ErrorWithStatus(c, http.StatusNotFound, "not found", "application not found")
				return
			}
			response.Success(c, app)
		})

		v1.POST("/credit-lines", func(c *gin.Context) {
			var req struct {
				OwnerID    string `json:"owner_id"`
				OwnerName  string `json:"owner_name"`
				OwnerType  string `json:"owner_type"`
				TotalLimit string `json:"total_limit"`
				Currency   string `json:"currency"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", err.Error())
				return
			}
			limit, err := decimal.NewFromString(req.TotalLimit)
			if err != nil || !limit.IsPositive() {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid total_limit", "total_limit must be positive decimal")
				return
			}
			line := domain.NewCreditLine(req.OwnerID, req.OwnerName, nonEmpty(req.OwnerType, "SUPPLIER"), limit, nonEmpty(req.Currency, "CNY"))
			line.ID = "CL" + strconv.FormatUint(idgen.GenID(), 10)
			store.mu.Lock()
			store.creditLines[line.ID] = line
			store.mu.Unlock()
			response.Success(c, line)
		})
	}

	addr := envOrDefault("SUPPLYCHAINFINANCE_HTTP_ADDR", ":9209")
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
	slog.Info("service supplychainfinance gracefully stopped")
}

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func nonEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

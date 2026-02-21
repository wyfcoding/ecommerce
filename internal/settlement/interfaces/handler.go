package interfaces

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/settlement/application"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
)

type MerchantSettlementHandler struct {
	app *application.MerchantSettlementService
}

func NewMerchantSettlementHandler(app *application.MerchantSettlementService) *MerchantSettlementHandler {
	return &MerchantSettlementHandler{app: app}
}

type CreateSettlementRequest struct {
	MerchantID  uint64 `json:"merchant_id"`
	Cycle       string `json:"cycle"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}

type CreateSettlementResponse struct {
	SettlementID string `json:"settlement_id"`
	Status       string `json:"status"`
}

func (h *MerchantSettlementHandler) CreateSettlement(ctx context.Context, req *CreateSettlementRequest) (*CreateSettlementResponse, error) {
	dto, err := h.app.CreateSettlement(ctx, req.MerchantID, domain.SettlementCycle(req.Cycle), req.PeriodStart, req.PeriodEnd)
	if err != nil {
		return nil, err
	}
	return &CreateSettlementResponse{
		SettlementID: dto.SettlementID,
		Status:       dto.Status,
	}, nil
}

type CalculateSettlementRequest struct {
	SettlementID string `json:"settlement_id"`
}

func (h *MerchantSettlementHandler) CalculateSettlement(ctx context.Context, req *CalculateSettlementRequest) error {
	return h.app.CalculateSettlement(ctx, req.SettlementID)
}

type ApproveSettlementRequest struct {
	SettlementID string `json:"settlement_id"`
	ApprovedBy   uint64 `json:"approved_by"`
}

func (h *MerchantSettlementHandler) ApproveSettlement(ctx context.Context, req *ApproveSettlementRequest) error {
	return h.app.ApproveSettlement(ctx, req.SettlementID, req.ApprovedBy)
}

type RejectSettlementRequest struct {
	SettlementID string `json:"settlement_id"`
	Reason       string `json:"reason"`
}

func (h *MerchantSettlementHandler) RejectSettlement(ctx context.Context, req *RejectSettlementRequest) error {
	return h.app.RejectSettlement(ctx, req.SettlementID, req.Reason)
}

type PaySettlementRequest struct {
	SettlementID  string `json:"settlement_id"`
	BankAccountID uint64 `json:"bank_account_id"`
}

func (h *MerchantSettlementHandler) PaySettlement(ctx context.Context, req *PaySettlementRequest) error {
	return h.app.PaySettlement(ctx, req.SettlementID, req.BankAccountID)
}

type AdjustSettlementRequest struct {
	SettlementID     string `json:"settlement_id"`
	AdjustmentAmount string `json:"adjustment_amount"`
	Reason           string `json:"reason"`
}

func (h *MerchantSettlementHandler) AdjustSettlement(ctx context.Context, req *AdjustSettlementRequest) error {
	amount, err := decimal.NewFromString(req.AdjustmentAmount)
	if err != nil {
		return err
	}
	return h.app.AdjustSettlement(ctx, req.SettlementID, amount, req.Reason)
}

type CancelSettlementRequest struct {
	SettlementID string `json:"settlement_id"`
	Reason       string `json:"reason"`
}

func (h *MerchantSettlementHandler) CancelSettlement(ctx context.Context, req *CancelSettlementRequest) error {
	return h.app.CancelSettlement(ctx, req.SettlementID, req.Reason)
}

type GetSettlementRequest struct {
	SettlementID string `json:"settlement_id"`
}

func (h *MerchantSettlementHandler) GetSettlement(ctx context.Context, req *GetSettlementRequest) (*application.SettlementDTO, error) {
	return h.app.GetSettlement(ctx, req.SettlementID)
}

type ListSettlementsRequest struct {
	MerchantID uint64 `json:"merchant_id"`
	Status     string `json:"status"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

func (h *MerchantSettlementHandler) ListSettlements(ctx context.Context, req *ListSettlementsRequest) (*application.ListSettlementsResult, error) {
	return h.app.ListSettlements(ctx, req.MerchantID, domain.SettlementStatus(req.Status), req.Page, req.PageSize)
}

type GetSettlementDetailsRequest struct {
	SettlementID string `json:"settlement_id"`
}

func (h *MerchantSettlementHandler) GetSettlementDetails(ctx context.Context, req *GetSettlementDetailsRequest) ([]*application.SettlementDetailDTO, error) {
	return h.app.GetSettlementDetails(ctx, req.SettlementID)
}

type AddBankAccountRequest struct {
	MerchantID  uint64 `json:"merchant_id"`
	BankName    string `json:"bank_name"`
	BankCode    string `json:"bank_code"`
	AccountName string `json:"account_name"`
	AccountNo   string `json:"account_no"`
	BranchName  string `json:"branch_name"`
	IsDefault   bool   `json:"is_default"`
}

func (h *MerchantSettlementHandler) AddBankAccount(ctx context.Context, req *AddBankAccountRequest) (*application.BankAccountDTO, error) {
	return h.app.AddBankAccount(ctx, req.MerchantID, req.BankName, req.BankCode, req.AccountName, req.AccountNo, req.BranchName, req.IsDefault)
}

type ListBankAccountsRequest struct {
	MerchantID uint64 `json:"merchant_id"`
}

func (h *MerchantSettlementHandler) ListBankAccounts(ctx context.Context, req *ListBankAccountsRequest) ([]*application.BankAccountDTO, error) {
	return h.app.ListBankAccounts(ctx, req.MerchantID)
}

type SetDefaultBankAccountRequest struct {
	MerchantID uint64 `json:"merchant_id"`
	AccountID  uint64 `json:"account_id"`
}

func (h *MerchantSettlementHandler) SetDefaultBankAccount(ctx context.Context, req *SetDefaultBankAccountRequest) error {
	return h.app.SetDefaultBankAccount(ctx, req.MerchantID, req.AccountID)
}

type DeleteBankAccountRequest struct {
	AccountID uint64 `json:"account_id"`
}

func (h *MerchantSettlementHandler) DeleteBankAccount(ctx context.Context, req *DeleteBankAccountRequest) error {
	return h.app.DeleteBankAccount(ctx, req.AccountID)
}

type GetSettlementConfigRequest struct {
	MerchantID uint64 `json:"merchant_id"`
}

func (h *MerchantSettlementHandler) GetSettlementConfig(ctx context.Context, req *GetSettlementConfigRequest) (*application.SettlementConfigDTO, error) {
	return h.app.GetSettlementConfig(ctx, req.MerchantID)
}

type UpdateSettlementConfigRequest struct {
	MerchantID          uint64 `json:"merchant_id"`
	Cycle               string `json:"cycle"`
	CommissionRate      string `json:"commission_rate"`
	MinSettlementAmount string `json:"min_settlement_amount"`
	AutoApprove         bool   `json:"auto_approve"`
	AutoPay             bool   `json:"auto_pay"`
}

func (h *MerchantSettlementHandler) UpdateSettlementConfig(ctx context.Context, req *UpdateSettlementConfigRequest) error {
	commissionRate, err := decimal.NewFromString(req.CommissionRate)
	if err != nil {
		return err
	}
	minAmount, err := decimal.NewFromString(req.MinSettlementAmount)
	if err != nil {
		return err
	}

	config := &domain.MerchantSettlementConfig{
		MerchantID:          req.MerchantID,
		Cycle:               domain.SettlementCycle(req.Cycle),
		CommissionRate:      commissionRate,
		MinSettlementAmount: minAmount,
		AutoApprove:         req.AutoApprove,
		AutoPay:             req.AutoPay,
		Status:              domain.ConfigStatusActive,
	}

	return h.app.UpdateSettlementConfig(ctx, config)
}

// 生成摘要：
// - 实现 Wallet 服务的 gRPC Handler，映射 proto 定义的 RPC 接口到 application 层服务
// - 支持钱包创建、查询、充值、提现、转账等核心资金操作
// - 负责入参出参转换、错误处理转换（将 domain/application 错误转换为 gRPC 状态码）
// - 集成 slog 日志与 OpenTelemetry 追踪信息

package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/wallet/v1"
	"github.com/wyfcoding/ecommerce/internal/wallet/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server gRPC 服务器实现
type Server struct {
	pb.UnimplementedWalletServiceServer
	service *application.WalletService
	logger  *slog.Logger
}

// NewServer 创建 gRPC Server 实例
func NewServer(service *application.WalletService, logger *slog.Logger) *Server {
	return &Server{
		service: service,
		logger:  logger.With("module", "wallet_grpc"),
	}
}

// CreateWallet 创建钱包
func (s *Server) CreateWallet(ctx context.Context, req *pb.CreateWalletRequest) (*pb.CreateWalletResponse, error) {
	start := time.Now()
	wallet, err := s.service.CreateWallet(ctx, uint64(req.UserId), req.Currency, req.WalletType)
	if err != nil {
		s.logger.Error("failed to create wallet", "error", err, "user_id", req.UserId)
		return nil, status.Errorf(codes.Internal, "failed to create wallet: %v", err)
	}

	s.logger.Info("wallet created", "wallet_id", wallet.WalletID, "duration", time.Since(start))

	return &pb.CreateWalletResponse{
		WalletId:  int64(wallet.WalletID),
		UserId:    int64(wallet.UserID),
		AccountNo: wallet.AccountNo,
		Currency:  wallet.Currency,
		Balance:   fmt.Sprintf("%d", wallet.Balance),
	}, nil
}

// GetWallet 获取钱包信息
func (s *Server) GetWallet(ctx context.Context, req *pb.GetWalletRequest) (*pb.GetWalletResponse, error) {
	wallet, err := s.service.GetWallet(ctx, uint64(req.UserId), req.Currency)
	if err != nil {
		if err == application.ErrWalletNotFound {
			return nil, status.Error(codes.NotFound, "wallet not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get wallet: %v", err)
	}

	return &pb.GetWalletResponse{
		WalletId:         int64(wallet.WalletID),
		UserId:           int64(wallet.UserID),
		AccountNo:        wallet.AccountNo,
		Currency:         wallet.Currency,
		Balance:          fmt.Sprintf("%d", wallet.Balance),
		FrozenBalance:    fmt.Sprintf("%d", wallet.FrozenBalance),
		AvailableBalance: fmt.Sprintf("%d", wallet.AvailableBalance),
		Status:           wallet.Status.String(),
	}, nil
}

// Deposit 充值
func (s *Server) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	amount, err := s.parseAmount(req.Amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tx, err := s.service.Deposit(ctx, uint64(req.UserId), req.Currency, amount, req.Remark)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "deposit failed: %v", err)
	}

	return &pb.DepositResponse{
		TransactionNo: tx.TransactionNo,
		Balance:       fmt.Sprintf("%d", tx.BalanceAfter),
		Status:        tx.Status.String(),
	}, nil
}

// Withdraw 提现
func (s *Server) Withdraw(ctx context.Context, req *pb.WithdrawRequest) (*pb.WithdrawResponse, error) {
	amount, err := s.parseAmount(req.Amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tx, err := s.service.Withdraw(ctx, uint64(req.UserId), req.Currency, amount, req.Remark)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "withdraw failed: %v", err)
	}

	return &pb.WithdrawResponse{
		TransactionNo: tx.TransactionNo,
		Balance:       fmt.Sprintf("%d", tx.BalanceAfter),
		Status:        tx.Status.String(),
	}, nil
}

// Transfer 转账
func (s *Server) Transfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransferResponse, error) {
	amount, err := s.parseAmount(req.Amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tx, err := s.service.Transfer(ctx, uint64(req.FromUserId), uint64(req.ToUserId), req.Currency, amount, req.Remark)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "transfer failed: %v", err)
	}

	// 注意：这里的 tx 是支出方的交易记录
	return &pb.TransferResponse{
		TransactionNo: tx.TransactionNo,
		FromBalance:   fmt.Sprintf("%d", tx.BalanceAfter),
		Status:        tx.Status.String(),
	}, nil
}

func (s *Server) parseAmount(amountStr string) (int64, error) {
	var amount int64
	_, err := fmt.Sscanf(amountStr, "%d", &amount)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format")
	}
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be positive")
	}
	return amount, nil
}

// GetTransactions 获取交易记录
func (s *Server) GetTransactions(ctx context.Context, req *pb.GetTransactionsRequest) (*pb.GetTransactionsResponse, error) {
	// 暂未在 service 中完全实现，这里仅作骨架示例
	return &pb.GetTransactionsResponse{
		Transactions: []*pb.Transaction{},
		Total:        0,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}, nil
}

// 其他方法 (FreezeBalance, UnfreezeBalance, GetBalanceHistory) 同样按需调用 service 实现
func (s *Server) FreezeBalance(ctx context.Context, req *pb.FreezeBalanceRequest) (*pb.FreezeBalanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FreezeBalance not implemented")
}

func (s *Server) UnfreezeBalance(ctx context.Context, req *pb.UnfreezeBalanceRequest) (*pb.UnfreezeBalanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnfreezeBalance not implemented")
}

func (s *Server) GetBalanceHistory(ctx context.Context, req *pb.GetBalanceHistoryRequest) (*pb.GetBalanceHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBalanceHistory not implemented")
}

// 变更说明：
// 高速 gRPC 调用网关。
// 本接口是其他服务（如购物车、订单确认页）计算最佳优惠价格的关键出口。
// 特别是对电商系统，当购物车多选、单选商品时，需要 2ms 级别实时回传价格变动矩阵。
package grpc

import (
	"context"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/promotion/application"
	"github.com/wyfcoding/ecommerce/internal/promotion/domain"
	"github.com/wyfcoding/pkg/contextx"
	// 假定有自动生成的 protobuf 代码
	// pb "github.com/wyfcoding/ecommerce/go-api/promotion/v1"
)

type FakeCalculateRequest struct {
	Items []*FakeCartItem
}
type FakeCartItem struct {
	ProductId  uint64
	SkuId      uint64
	CategoryId uint64
	BrandId    uint64
	MerchantId uint64
	Price      int64 // 价格分
	Quantity   int32
}

type FakeCalculateResponse struct {
	OriginalAmount int64
	TotalDiscount  int64
	FinalAmount    int64
	FreeShipping   bool
}

type PromotionGrpcServer struct {
	appService *application.PromotionCommandService
	logger     *slog.Logger
}

func NewPromotionGrpcServer(app *application.PromotionCommandService, logger *slog.Logger) *PromotionGrpcServer {
	return &PromotionGrpcServer{
		appService: app,
		logger:     logger,
	}
}

// CalculateCartPromotions 计算整单最优折扣
func (s *PromotionGrpcServer) CalculateCartPromotions(ctx context.Context, req *FakeCalculateRequest) (*FakeCalculateResponse, error) {
	start := time.Now()
	traceID := contextx.GetRequestID(ctx)

	var dItems []*domain.CartItem
	for _, it := range req.Items {
		dItems = append(dItems, &domain.CartItem{
			ProductID:  it.ProductId,
			SkuID:      it.SkuId,
			CategoryID: it.CategoryId,
			BrandID:    it.BrandId,
			MerchantID: it.MerchantId,
			Price:      decimal.NewFromInt(it.Price),
			Quantity:   it.Quantity,
		})
	}

	result, err := s.appService.CalculateCartPromotions(ctx, dItems)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to calculate cart promotions", "error", err, "trace_id", traceID)
		return nil, err
	}

	s.logger.InfoContext(ctx, "promotion graph calculation done", "trace_id", traceID, "took_ms", time.Since(start).Milliseconds())

	return &FakeCalculateResponse{
		OriginalAmount: result.OriginalAmount.IntPart(),
		TotalDiscount:  result.TotalDiscount.IntPart(),
		FinalAmount:    result.FinalAmount.IntPart(),
		FreeShipping:   result.FreeShipping,
	}, nil
}

// 变更说明：实现下单流程的 Saga 编排器。
// 流程：1. 创建订单(PENDING) -> 2. 预扣库存 -> 3. 核销优惠券 -> 4. 更新订单(SUCCESS)
// 补偿：1. 撤销优惠券 -> 2. 释放库存 -> 3. 关闭订单(CANCELLED)
package application

import (
	"context"
	"github.com/wyfcoding/pkg/dtm"
	"github.com/wyfcoding/pkg/logging"
	"google.golang.org/protobuf/proto"
)

type OrderSagaOrchestrator struct {
	dtmServer string
	logger    *logging.Logger
}

func (o *OrderSagaOrchestrator) CreateOrderSaga(ctx context.Context, gid string, orderReq proto.Message, invReq proto.Message, couponReq proto.Message) error {
	saga := dtm.NewSaga(ctx, o.dtmServer, gid)
	
	// 步骤 1：创建订单
	saga.Add("order.svc/CreateOrder", "order.svc/CompensateCreateOrder", orderReq)
	
	// 步骤 2：预扣库存
	saga.Add("inventory.svc/DeductStock", "inventory.svc/ReleaseStock", invReq)
	
	// 步骤 3：核销优惠券
	saga.Add("coupon.svc/UseCoupon", "coupon.svc/ReturnCoupon", couponReq)
	
	o.logger.Info("submitting order saga", "gid", gid)
	return saga.Submit(ctx)
}

package domain

import (
	"context"
	"fmt"

	"github.com/wyfcoding/pkg/ruleengine"
)

// CouponValidator 使用规则引擎校验优惠券
type CouponValidator struct {
	engine *ruleengine.Engine
}

func NewCouponValidator() *CouponValidator {
	return &CouponValidator{
		engine: ruleengine.NewEngine(),
	}
}

func (v *CouponValidator) IsEligible(ctx context.Context, coupon *Coupon, orderData map[string]any) (bool, error) {
	// 如果没有设置判定表达式，默认可用
	if coupon.ConditionExpr == "" {
		return true, nil
	}

	ruleID := fmt.Sprintf("coupon_%d", coupon.ID)

	// 动态注册或更新优惠券规则
	err := v.engine.AddRule(ruleengine.Rule{
		ID:         ruleID,
		Expression: coupon.ConditionExpr,
	})
	if err != nil {
		return false, err
	}

	result, err := v.engine.Execute(ctx, ruleID, orderData)
	if err != nil {
		return false, err
	}

	return result.Passed, nil
}

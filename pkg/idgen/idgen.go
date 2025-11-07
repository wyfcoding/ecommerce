package idgen

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
	once sync.Once
)

// Init 初始化ID生成器
func Init(machineID int64) error {
	var err error
	once.Do(func() {
		// 设置起始时间为2024-01-01
		snowflake.Epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano() / 1000000
		node, err = snowflake.NewNode(machineID)
	})
	return err
}

// GenID 生成唯一ID
func GenID() uint64 {
	if node == nil {
		// 如果未初始化，使用默认机器ID 1
		_ = Init(1)
	}
	return uint64(node.Generate().Int64())
}

// GenOrderNo 生成订单号
// 格式：O + 时间戳 + 随机数
func GenOrderNo() string {
	return fmt.Sprintf("O%d", GenID())
}

// GenPaymentNo 生成支付单号
// 格式：P + 时间戳 + 随机数
func GenPaymentNo() string {
	return fmt.Sprintf("P%d", GenID())
}

// GenRefundNo 生成退款单号
// 格式：R + 时间戳 + 随机数
func GenRefundNo() string {
	return fmt.Sprintf("R%d", GenID())
}

// GenSPUNo 生成SPU编号
// 格式：SPU + ID
func GenSPUNo() string {
	return fmt.Sprintf("SPU%d", GenID())
}

// GenSKUNo 生成SKU编号
// 格式：SKU + ID
func GenSKUNo() string {
	return fmt.Sprintf("SKU%d", GenID())
}

// GenCouponCode 生成优惠券码
// 格式：C + ID
func GenCouponCode() string {
	return fmt.Sprintf("C%d", GenID())
}

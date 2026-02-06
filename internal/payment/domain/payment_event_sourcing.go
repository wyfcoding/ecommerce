// 生成摘要：支付事件回放工具，用于读侧重建聚合。
package domain

import "github.com/wyfcoding/pkg/eventsourcing"

// RebuildPaymentFromEvents 从事件历史重建支付聚合状态。
func RebuildPaymentFromEvents(events []eventsourcing.DomainEvent) (*Payment, error) {
	if len(events) == 0 {
		return nil, nil
	}

	payment := &Payment{}
	eventsourcing.LoadFromHistory(payment, events)
	if payment.PaymentNo != "" {
		payment.SetID(payment.PaymentNo)
	}
	payment.initFSM()
	return payment, nil
}

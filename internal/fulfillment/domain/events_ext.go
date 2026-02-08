package domain

import "time"

// FulfillmentSplitEvent 履约单拆分事件
type FulfillmentSplitEvent struct {
	OriginalFulfillmentID uint64    `json:"original_fulfillment_id"`
	NewFulfillmentID      string    `json:"new_fulfillment_id"`
	Timestamp             time.Time `json:"timestamp"`
}

func (e *FulfillmentSplitEvent) EventName() string     { return "fulfillment.split" }
func (e *FulfillmentSplitEvent) OccurredAt() time.Time { return e.Timestamp }

// FulfillmentMergedEvent 履约单合并事件
type FulfillmentMergedEvent struct {
	TargetFulfillmentID uint64    `json:"target_fulfillment_id"`
	SourceFulfillmentID uint64    `json:"source_fulfillment_id"`
	Timestamp           time.Time `json:"timestamp"`
}

func (e *FulfillmentMergedEvent) EventName() string     { return "fulfillment.merged" }
func (e *FulfillmentMergedEvent) OccurredAt() time.Time { return e.Timestamp }

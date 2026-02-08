package domain

import (
	"errors"
	"fmt"
	"time"
)

// Split 将履约单拆分为多个
// splitItems: map[OriginalItemIndex]QuantityToSplit
func (f *Fulfillment) Split(splitItems map[int]int32) (*Fulfillment, error) {
	if f.Status != FulfillmentStatusPending {
		return nil, errors.New("can only split pending fulfillment")
	}

	newFulfillment := NewFulfillment(f.OrderNo, f.MerchantID, f.StoreID, f.WarehouseID, f.Type)
	newFulfillment.SetShippingAddress(f.ReceiverName, f.ReceiverPhone, f.Province, f.City, f.District, f.Address, f.PostalCode)

	for idx, qty := range splitItems {
		if idx < 0 || idx >= len(f.Items) {
			return nil, errors.New("invalid item index")
		}
		item := &f.Items[idx]

		if qty <= 0 || qty > item.Quantity {
			return nil, errors.New("invalid split quantity")
		}

		// Add to new fulfillment
		newFulfillment.AddItem(item.SKUID, item.ProductName, item.SKUName, item.ImageURL, item.Location, item.BatchNo, qty)

		// Deduct from original
		item.Quantity -= qty
	}

	// Remove items with 0 quantity from original
	// logic to compact Items slice... simplified here
	newItems := make([]FulfillmentItem, 0)
	for _, item := range f.Items {
		if item.Quantity > 0 {
			newItems = append(newItems, item)
		}
	}
	f.Items = newItems

	if len(f.Items) == 0 {
		return nil, errors.New("original fulfillment becomes empty after split, use update instead")
	}

	f.addEvent(&FulfillmentSplitEvent{
		OriginalFulfillmentID: uint64(f.ID),
		NewFulfillmentID:      newFulfillment.FulfillmentNo, // Temporary ID reference
		Timestamp:             time.Now(),
	})

	return newFulfillment, nil
}

// Merge 合并另一个履约单到当前单
func (f *Fulfillment) Merge(other *Fulfillment) error {
	if f.Status != FulfillmentStatusPending || other.Status != FulfillmentStatusPending {
		return errors.New("can only merge pending fulfillments")
	}
	if f.MerchantID != other.MerchantID || f.WarehouseID != other.WarehouseID {
		return errors.New("cannot merge fulfillments from different merchant or warehouse")
	}
	// Check address consistency...

	for _, item := range other.Items {
		f.AddItem(item.SKUID, item.ProductName, item.SKUName, item.ImageURL, item.Location, item.BatchNo, item.Quantity)
	}

	other.Status = FulfillmentStatusCancelled
	other.CancelReason = fmt.Sprintf("Merged into %s", f.FulfillmentNo)

	f.addEvent(&FulfillmentMergedEvent{
		TargetFulfillmentID: uint64(f.ID),
		SourceFulfillmentID: uint64(other.ID),
		Timestamp:           time.Now(),
	})

	return nil
}

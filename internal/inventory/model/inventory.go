package model

// Stock represents the stock quantity of a SKU in the business logic layer.
type Stock struct {
	SKUID          uint64
	Quantity       uint32
	LockedQuantity uint32
}

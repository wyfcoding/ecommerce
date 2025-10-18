package model

// ShippingRule represents a shipping rule in the business logic layer.
type ShippingRule struct {
	ID          uint
	Name        string
	Origin      string
	Destination string
	MinWeight   float64
	MaxWeight   float64
	BaseCost    uint64
	PerKgCost   uint64
}

// ItemInfo represents item information for shipping cost calculation.
type ItemInfo struct {
	ProductID uint64
	Quantity  uint32
	WeightKg  float64 // Weight per item in kilograms
}

// AddressInfo represents address information for shipping cost calculation.
type AddressInfo struct {
	Province string
	City     string
	District string
}

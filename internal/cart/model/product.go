package model

// SkuInfo is the product SKU information model used internally by the cart service.
// It is a Data Transfer Object (DTO) used to decouple the cart service from the product service.
// Even if the internal model `product.Sku` of the product service changes, the cart service is not affected.
type SkuInfo struct {
	SkuID  uint64
	SpuID  uint64
	Title  string
	Price  uint64
	Image  string
	Specs  map[string]string
	Status int32
}

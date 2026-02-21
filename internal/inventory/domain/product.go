package domain

// Product 产品实体 (用于库存优化)
type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Cost  float64 `json:"cost"`
	Price float64 `json:"price"`
}

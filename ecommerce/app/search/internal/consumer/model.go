package consumer

// CanalMessage 是 Canal 產生的消息的結構
type CanalMessage struct {
	Type     string                   `json:"type"` // "INSERT", "UPDATE", "DELETE"
	Data     []map[string]interface{} `json:"data"`
	Old      []map[string]interface{} `json:"old"` // For UPDATE, the old values
	Database string                   `json:"database"`
	Table    string                   `json:"table"`
}

// ProductDocument 是我們要存入 Elasticsearch 的商品文檔結構
type ProductDocument struct {
	SpuID      uint64 `json:"spu_id"`
	CategoryID uint64 `json:"category_id"`
	BrandID    uint64 `json:"brand_id"`
	Title      string `json:"title"`
	SubTitle   string `json:"sub_title"`
	Price      uint64 `json:"price"` // 假設 SPU 表中有一個最低價字段
	Status     int8   `json:"status"`
}

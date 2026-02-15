// Package domain 商家读模型仓储接口（CQRS 查询侧）
// 生成摘要：
// 1) 定义商家服务的读模型仓储接口，用于 CQRS 查询侧
// 2) 读模型基于 Redis 缓存，提供高性能的商家信息查询、店铺列表查询
// 3) 与写模型（MySQL）解耦，通过事件投影保持最终一致性
package domain

import (
	"context"
	"time"
)

// MerchantReadModel 商家读模型（用于快速查询的扁平化视图）
type MerchantReadModel struct {
	ID             uint64    `json:"id"`
	MerchantID     uint64    `json:"merchant_id"`
	MerchantNo     string    `json:"merchant_no"`
	UserID         uint64    `json:"user_id"`
	Name           string    `json:"name"`
	LegalName      string    `json:"legal_name"`
	ContactPhone   string    `json:"contact_phone"`
	ContactEmail   string    `json:"contact_email"`
	Type           string    `json:"type"`
	Status         string    `json:"status"`
	Level          string    `json:"level"`
	LogoURL        string    `json:"logo_url"`
	Description    string    `json:"description"`
	CommissionRate float64   `json:"commission_rate"`
	CreditScore    float64   `json:"credit_score"`
	TotalSales     int64     `json:"total_sales"`
	TotalOrders    int32     `json:"total_orders"`
	Rating         float64   `json:"rating"`
	ApprovedAt     time.Time `json:"approved_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// StoreReadModel 店铺读模型
type StoreReadModel struct {
	ID             uint64    `json:"id"`
	StoreID        uint64    `json:"store_id"`
	StoreNo        string    `json:"store_no"`
	MerchantID     uint64    `json:"merchant_id"`
	Name           string    `json:"name"`
	LogoURL        string    `json:"logo_url"`
	BannerURL      string    `json:"banner_url"`
	Description    string    `json:"description"`
	Announcement   string    `json:"announcement"`
	Categories     []string  `json:"categories"`
	Rating         float64   `json:"rating"`
	ProductCount   int32     `json:"product_count"`
	MonthlySales   int64     `json:"monthly_sales"`
	IsOpen         bool      `json:"is_open"`
	BusinessHours  string    `json:"business_hours"`
	Address        string    `json:"address"`
	CreatedAt      time.Time `json:"created_at"`
}

// MerchantSearchQuery 商家搜索查询条件
type MerchantSearchQuery struct {
	Keyword     string     `json:"keyword,omitempty"`
	Type        string     `json:"type,omitempty"`
	Status      string     `json:"status,omitempty"`
	Level       string     `json:"level,omitempty"`
	MinRating   float64    `json:"min_rating,omitempty"`
	MinSales    int64      `json:"min_sales,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Page        int        `json:"page"`
	PageSize    int        `json:"page_size"`
	SortBy      string     `json:"sort_by,omitempty"`
	SortOrder   string     `json:"sort_order,omitempty"`
}

// MerchantSearchResult 商家搜索结果
type MerchantSearchResult struct {
	Items    []*MerchantReadModel `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// MerchantReadRepository 商家读模型仓储接口
type MerchantReadRepository interface {
	// GetByMerchantID 根据商家ID获取读模型。
	GetByMerchantID(ctx context.Context, merchantID uint64) (*MerchantReadModel, error)
	// GetByUserID 根据用户ID获取读模型。
	GetByUserID(ctx context.Context, userID uint64) (*MerchantReadModel, error)
	// GetByMerchantNo 根据商家编号获取读模型。
	GetByMerchantNo(ctx context.Context, merchantNo string) (*MerchantReadModel, error)
	// Save 保存或更新读模型。
	Save(ctx context.Context, model *MerchantReadModel) error
	// Delete 删除读模型。
	Delete(ctx context.Context, merchantID uint64) error
}

// StoreReadRepository 店铺读模型仓储接口
type StoreReadRepository interface {
	// GetByStoreID 根据店铺ID获取读模型。
	GetByStoreID(ctx context.Context, storeID uint64) (*StoreReadModel, error)
	// GetByMerchantID 根据商家ID获取所有店铺读模型。
	GetByMerchantID(ctx context.Context, merchantID uint64) ([]*StoreReadModel, error)
	// Save 保存或更新读模型。
	Save(ctx context.Context, model *StoreReadModel) error
	// Delete 删除读模型。
	Delete(ctx context.Context, storeID uint64) error
}

// MerchantSearchRepository 商家搜索仓储接口（Elasticsearch）
type MerchantSearchRepository interface {
	// IndexMerchant 索引商家信息到 ES。
	IndexMerchant(ctx context.Context, merchant *MerchantReadModel) error
	// SearchMerchants 多维度搜索商家。
	SearchMerchants(ctx context.Context, query *MerchantSearchQuery) (*MerchantSearchResult, error)
	// IndexStore 索引店铺信息到 ES。
	IndexStore(ctx context.Context, store *StoreReadModel) error
	// SearchStores 搜索店铺。
	SearchStores(ctx context.Context, keyword string, page, pageSize int) ([]*StoreReadModel, int64, error)
}
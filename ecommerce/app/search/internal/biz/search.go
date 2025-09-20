package biz

import "context"

// SearchRequest 封裝了所有搜索參數
type SearchRequest struct {
	Keyword    string
	CategoryID uint64
	BrandID    uint64
	SortBy     string
	Page       uint32
	PageSize   uint32
}

// ProductHit 代表一個搜索結果項
type ProductHit struct {
	SpuID     uint64
	Title     string
	SubTitle  string
	MainImage string
	Price     uint64
}

// SearchResult 封裝了搜索結果
type SearchResult struct {
	Hits  []*ProductHit
	Total uint64
}

// SearchRepo 定義了搜索數據倉庫的接口
type SearchRepo interface {
	Search(ctx context.Context, req *SearchRequest) (*SearchResult, error)
}

// SearchUsecase 是搜索業務的容器
type SearchUsecase struct {
	repo SearchRepo
}

// NewSearchUsecase 創建一個新的 SearchUsecase
func NewSearchUsecase(repo SearchRepo) *SearchUsecase {
	return &SearchUsecase{repo: repo}
}

// SearchProducts 執行商品搜索
func (uc *SearchUsecase) SearchProducts(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	// 在這裡可以擴展業務邏輯，例如：
	// 1. 關鍵詞改寫、同義詞擴展
	// 2. 搜索結果與營銷活動數據（來自 marketing-service）進行聚合
	// 3. 記錄用戶搜索行為，用於分析和推薦
	return uc.repo.Search(ctx, req)
}

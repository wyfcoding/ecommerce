package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ecommerce/ecommerce/app/search/internal/biz"

	"github.com/elastic/go-elasticsearch/v8"
)

type searchRepo struct {
	es *elasticsearch.Client
}

func NewSearchRepo(es *elasticsearch.Client) biz.SearchRepo {
	return &searchRepo{es: es}
}

// ESSearchResponse 用於解析 ES 返回的 JSON
type ESSearchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source biz.ProductHit `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (r *searchRepo) Search(ctx context.Context, req *biz.SearchRequest) (*biz.SearchResult, error) {
	var query strings.Builder
	query.WriteString("{\n")

	// 分頁
	from := (req.Page - 1) * req.PageSize
	query.WriteString(fmt.Sprintf(`"from": %d, "size": %d,`, from, req.PageSize))

	// 查詢主體
	query.WriteString(`"query": { "bool": {`)
	// 關鍵詞查詢
	query.WriteString(fmt.Sprintf(`"must": [{"multi_match": {"query": "%s", "fields": ["title", "sub_title"]}}],`, req.Keyword))

	// 過濾條件
	var filters []string
	if req.CategoryID > 0 {
		filters = append(filters, fmt.Sprintf(`{"term": {"category_id": %d}}`, req.CategoryID))
	}
	if req.BrandID > 0 {
		filters = append(filters, fmt.Sprintf(`{"term": {"brand_id": %d}}`, req.BrandID))
	}
	query.WriteString(fmt.Sprintf(`"filter": [%s]`, strings.Join(filters, ",")))
	query.WriteString("}},\n") // end of bool query

	// 排序
	var sort string
	switch req.SortBy {
	case "price_asc":
		sort = `{"price": "asc"}`
	case "price_desc":
		sort = `{"price": "desc"}`
	default: // 默認按相關性得分
		sort = `{"_score": "desc"}`
	}
	query.WriteString(fmt.Sprintf(`"sort": [%s]`, sort))

	query.WriteString("\n}")

	// 執行搜索
	res, err := r.es.Search(
		r.es.Search.WithContext(ctx),
		r.es.Search.WithIndex("products"),
		r.es.Search.WithBody(strings.NewReader(query.String())),
		r.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch search error: %s", res.String())
	}

	// 解析結果
	var esRes ESSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&esRes); err != nil {
		return nil, err
	}

	// 轉換為 biz.SearchResult
	searchResult := &biz.SearchResult{
		Total: uint64(esRes.Hits.Total.Value),
		Hits:  make([]*biz.ProductHit, len(esRes.Hits.Hits)),
	}
	for i, hit := range esRes.Hits.Hits {
		searchResult.Hits[i] = &hit.Source
	}

	return searchResult, nil
}

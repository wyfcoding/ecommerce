package domain

import (
	"context"
	"math/rand"
	"sort"
	"time"
)

// VectorRepresentation 商品向量表示 (Embedding)
type VectorRepresentation struct {
	ID       uint64    // 商品/文档ID
	Vector   []float32 // 向量数据 (e.g. 768维)
	Metadata map[string]any
}

// VectorSearchResult 向量搜索结果
type VectorSearchResult struct {
	ID    uint64  `json:"id"`
	Score float32 `json:"score"` // 相似度分数 (Cosine Similarity / Euclidean)
}

// VectorEngine 向量搜索引擎接口
// 负责处理语义搜索、以图搜图等高级场景
type VectorEngine interface {
	// Index 将商品转换为向量并建立索引
	Index(ctx context.Context, item *VectorRepresentation) error
	// Search 搜索相似商品
	Search(ctx context.Context, queryVector []float32, topK int) ([]*VectorSearchResult, error)
	// MockEmbedding 生成模拟向量 (仅用于演示)
	MockEmbedding(text string) []float32
}

// SimulatedVectorEngine 模拟的向量引擎 (用于无需真实 VectorDB 的场景)
type SimulatedVectorEngine struct {
	// In-memory index for demo
	index map[uint64][]float32
}

func NewSimulatedVectorEngine() *SimulatedVectorEngine {
	return &SimulatedVectorEngine{
		index: make(map[uint64][]float32),
	}
}

func (e *SimulatedVectorEngine) Index(ctx context.Context, item *VectorRepresentation) error {
	e.index[item.ID] = item.Vector
	return nil
}

func (e *SimulatedVectorEngine) Search(ctx context.Context, queryVector []float32, topK int) ([]*VectorSearchResult, error) {
	// Brute-force search (Scan all) - O(N)
	// 真实生产环境会使用 HNSW (Hierarchical Navigable Small World) 索引实现 O(logN)

	results := make([]*VectorSearchResult, 0, len(e.index))

	for id, vec := range e.index {
		score := cosineSimilarity(queryVector, vec)
		results = append(results, &VectorSearchResult{
			ID:    id,
			Score: score,
		})
	}

	// Sort by score desc
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

func (e *SimulatedVectorEngine) MockEmbedding(text string) []float32 {
	// 模拟生成 128 维向量
	// 真实场景会调用 BERT/CLIP 模型
	dim := 128
	vec := make([]float32, dim)
	r := rand.New(rand.NewSource(int64(len(text)) + time.Now().UnixNano()))

	var norm float32
	for i := 0; i < dim; i++ {
		v := r.Float32()
		vec[i] = v
		norm += v * v
	}
	// Normalize
	/*
		normSqrt := float32(math.Sqrt(float64(norm)))
		for i := 0; i < dim; i++ {
			vec[i] /= normSqrt
		}
	*/
	return vec
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	// 简化计算，假设已归一化，否则需要除以 sqrt(normA * normB)
	return dot // / (sqrt(normA) * sqrt(normB))
}

package application

// ProductRecommendationDTO 结构体定义。
type ProductRecommendationDTO struct {
	ProductID uint64
	Score     float64
	Reason    string
}

// FeedItemDTO 结构体定义。
type FeedItemDTO struct {
	ItemType  string
	ItemID    string
	Title     string
	ImageURL  string
	TargetURL string
	Score     float64
}

// ProductSearchResultDTO 结构体定义。
type ProductSearchResultDTO struct {
	ProductID       uint64
	SimilarityScore float64
}

// FraudScoreDTO 结构体定义。
type FraudScoreDTO struct {
	FraudScore   float64
	IsFraudulent bool
	Reasons      []string
}

package model

import "time"

// Review is a Review model.
type Review struct {
	ID        uint64
	SpuID     uint64
	UserID    uint64
	Rating    uint32 // 评分 (1-5星)
	Comment   string
	Images    []string // 评论图片URL
	CreatedAt time.Time
}

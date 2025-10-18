package model

import (
	"time"

	"gorm.io/gorm"
)

// ReviewStatus 定义了评论的状态
type ReviewStatus string

const (
	StatusPending  ReviewStatus = "PENDING"  // 待审核
	StatusApproved ReviewStatus = "APPROVED" // 已批准
	StatusRejected ReviewStatus = "REJECTED" // 已拒绝
)

// Review 评论主模型
type Review struct {
	ID        uint         `gorm:"primarykey" json:"id"`
	UserID    uint         `gorm:"not null;index" json:"user_id"`
	ProductID uint         `gorm:"not null;index" json:"product_id"`
	OrderID   uint         `gorm:"not null;index" json:"order_id"` // 关联的订单ID，用于验证购买行为

	Rating    int          `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"` // 评分 (1-5)
	Title     string       `gorm:"type:varchar(255)" json:"title"`
	Content   string       `gorm:"type:text;not null" json:"content"`
	Images    string       `gorm:"type:text" json:"images"` // 图片URL列表 (JSON array)
	Likes     int          `gorm:"not null;default:0" json:"likes"` // 点赞数

	Status    ReviewStatus `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	IsAnonymous bool       `gorm:"not null;default:false" json:"is_anonymous"` // 是否匿名

	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`

	Comments []Comment `gorm:"foreignKey:ReviewID" json:"comments,omitempty"` // 评论的回复
}

// Comment 评论的回复模型
type Comment struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ReviewID  uint      `gorm:"not null;index" json:"review_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"` // 回复者ID
	Content   string    `gorm:"type:text;not null" json:"content"`
	IsSeller  bool      `gorm:"not null;default:false" json:"is_seller"` // 是否为商家回复
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductReviewStats 产品的评论统计 (可以单独作为一个表，也可以由其他服务聚合)
type ProductReviewStats struct {
	ProductID     uint    `gorm:"primaryKey" json:"product_id"`
	AverageRating float64 `gorm:"type:decimal(3,2)" json:"average_rating"`
	TotalReviews  int     `json:"total_reviews"`
	RatingCounts  string  `gorm:"type:varchar(255)" json:"rating_counts"` // e.g., {"1":10, "2":5, "3":20, "4":50, "5":100}
}


// TableName 自定义表名
func (Review) TableName() string {
	return "reviews"
}

func (Comment) TableName() string {
	return "review_comments"
}

func (ProductReviewStats) TableName() string {
	return "product_review_stats"
}

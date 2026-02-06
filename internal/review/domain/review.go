package domain

import (
	"database/sql/driver" // 导入数据库驱动接口。
	"encoding/json"       // 导入JSON编码/解码库。
	"errors"              // 导入标准错误处理库。
	"time"
)

// ReviewStatus 定义了评论的审核状态。
type ReviewStatus int

const (
	ReviewStatusPending  ReviewStatus = 1 // 待审核：评论已提交，等待管理员审核。
	ReviewStatusApproved ReviewStatus = 2 // 已通过：评论已审核通过，对外可见。
	ReviewStatusRejected ReviewStatus = 3 // 已拒绝：评论未通过审核，对外不可见。
)

// StringArray 定义了一个字符串切片类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将Go的 []string 类型作为JSON字符串存储到数据库，并从数据库读取。
type StringArray []string

// Value 实现 driver.Valuer 接口，将 StringArray 转换为数据库可以存储的值（JSON字节数组）。
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a) // 将切片编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 StringArray。
func (a *StringArray) Scan(value any) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a) // 将JSON字节数组解码为切片。
}

// Review 实体是评论模块的聚合根。
// 它代表一条用户对商品的评论，包含了用户、商品、订单信息、评分、内容和图片等。
type Review struct {
	ID        uint64       `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	UserID    uint64       `json:"user_id"`    // 评论用户ID。
	ProductID uint64       `json:"product_id"` // 评论商品ID。
	OrderID   uint64       `json:"order_id"`   // 关联订单ID。
	SkuID     uint64       `json:"sku_id"`     // 关联SKU ID。
	Rating    int          `json:"rating"`     // 评分，范围1到5。
	Content   string       `json:"content"`    // 评论内容，不允许为空。
	Images    StringArray  `json:"images"`     // 评论图片URL列表，存储为JSON。
	Status    ReviewStatus `json:"status"`     // 评论状态，默认为待审核。
	LikeCount int          `json:"like_count"` // 评论点赞数。
}

// ProductRatingStats 值对象定义了商品的评分统计信息。
type ProductRatingStats struct {
	ProductID     uint64  `json:"product_id"`     // 商品ID。
	AverageRating float64 `json:"average_rating"` // 平均评分。
	TotalReviews  int     `json:"total_reviews"`  // 评论总数。
	Rating5Count  int     `json:"rating_5_count"` // 5星评论数量。
	Rating4Count  int     `json:"rating_4_count"` // 4星评论数量。
	Rating3Count  int     `json:"rating_3_count"` // 3星评论数量。
	Rating2Count  int     `json:"rating_2_count"` // 2星评论数量。
	Rating1Count  int     `json:"rating_1_count"` // 1星评论数量。
}

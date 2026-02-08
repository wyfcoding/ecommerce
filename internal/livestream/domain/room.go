package domain

import (
	"gorm.io/gorm"
)

type RoomStatus string

const (
	StatusCreated RoomStatus = "CREATED"
	StatusLiving  RoomStatus = "LIVING"
	StatusEnded   RoomStatus = "ENDED"
)

// Room 直播间聚合
type Room struct {
	gorm.Model
	RoomID      string     `gorm:"column:room_id;type:varchar(32);unique_index;not null"`
	OwnerID     string     `gorm:"column:owner_id;type:varchar(32);index;not null"`
	Title       string     `gorm:"column:title;type:varchar(255);not null"`
	CoverURL    string     `gorm:"column:cover_url;type:varchar(512)"`
	Status      RoomStatus `gorm:"column:status;type:varchar(20);not null;default:'CREATED'"`
	StreamURL   string     `gorm:"column:stream_url;type:varchar(512)"`
	ViewerCount int32      `gorm:"column:viewer_count;default:0"`
	Products    []Product  `gorm:"foreignKey:RoomID;references:RoomID"`
}

// Product 直播间内展示的商品
type Product struct {
	gorm.Model
	RoomID    string `gorm:"column:room_id;type:varchar(32);index;not null"`
	ProductID string `gorm:"column:product_id;type:varchar(32);not null"`
	Price     string `gorm:"column:price;type:varchar(32)"`
	Stock     int32  `gorm:"column:stock"`
}

func (Room) TableName() string    { return "livestream_rooms" }
func (Product) TableName() string { return "livestream_products" }

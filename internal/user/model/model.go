package model

import (
	"time"
)

// User 是用户的业务领域模型。
type User struct {
	ID        uint64
	Username  string
	Password  string
	Nickname  string
	Avatar    string
	Gender    int32
	Birthday  time.Time
	Phone     string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Address 是收货地址的业务领域模型。
type Address struct {
	ID              uint64
	UserID          uint64
	Name            *string
	Phone           *string
	Province        *string
	City            *string
	District        *string
	DetailedAddress *string
	IsDefault       *bool
}

package application

import "time"

// CreateUserCommand 创建用户命令
type CreateUserCommand struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone"`
}

// UpdateUserCommand 更新用户命令
type UpdateUserCommand struct {
	ID       uint       `json:"id"`
	Nickname string     `json:"nickname"`
	Avatar   string     `json:"avatar"`
	Gender   int8       `json:"gender"`
	Birthday *time.Time `json:"birthday"`
}

// UserDTO 用户信息传输对象
type UserDTO struct {
	ID        uint       `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Nickname  string     `json:"nickname"`
	Avatar    string     `json:"avatar"`
	Gender    int8       `json:"gender"`
	Birthday  *time.Time `json:"birthday"`
	Status    int8       `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// LoginCommand 登录命令
type LoginCommand struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string   `json:"token"`
	ExpiresIn int      `json:"expires_in"`
	User      *UserDTO `json:"user"`
}

// AddressDTO 地址信息传输对象
type AddressDTO struct {
	ID              uint      `json:"id"`
	UserID          uint      `json:"user_id"`
	RecipientName   string    `json:"recipient_name"`
	PhoneNumber     string    `json:"phone_number"`
	Province        string    `json:"province"`
	City            string    `json:"city"`
	District        string    `json:"district"`
	DetailedAddress string    `json:"detailed_address"`
	PostalCode      string    `json:"postal_code"`
	IsDefault       bool      `json:"is_default"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AddAddressCommand 添加地址命令
type AddAddressCommand struct {
	UserID          uint   `json:"user_id"`
	RecipientName   string `json:"recipient_name" binding:"required"`
	PhoneNumber     string `json:"phone_number" binding:"required"`
	Province        string `json:"province" binding:"required"`
	City            string `json:"city" binding:"required"`
	District        string `json:"district" binding:"required"`
	DetailedAddress string `json:"detailed_address" binding:"required"`
	PostalCode      string `json:"postal_code"`
	IsDefault       bool   `json:"is_default"`
}

// UpdateAddressCommand 更新地址命令
type UpdateAddressCommand struct {
	ID              uint   `json:"id"`
	UserID          uint   `json:"user_id"`
	RecipientName   string `json:"recipient_name"`
	PhoneNumber     string `json:"phone_number"`
	Province        string `json:"province"`
	City            string `json:"city"`
	District        string `json:"district"`
	DetailedAddress string `json:"detailed_address"`
	PostalCode      string `json:"postal_code"`
	IsDefault       bool   `json:"is_default"`
}

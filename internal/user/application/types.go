package application

import "time"

// CreateUserCommand 创建用户命令
type CreateUserCommand struct {
	Username string
	Email    string
	Password string
	Phone    string
}

// UpdateUserCommand 更新用户命令
type UpdateUserCommand struct {
	ID       uint
	Nickname string
	Avatar   string
	Gender   int8
	Birthday *time.Time
}

// LoginCommand 登录命令
type LoginCommand struct {
	Username string
	Password string
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string   `json:"token"`
	ExpiresIn int      `json:"expires_in"`
	User      *UserDTO `json:"user"`
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
	UserID          uint
	RecipientName   string
	PhoneNumber     string
	Province        string
	City            string
	District        string
	DetailedAddress string
	PostalCode      string
	IsDefault       bool
}

// UpdateAddressCommand 更新地址命令
type UpdateAddressCommand struct {
	ID              uint
	UserID          uint
	RecipientName   string
	PhoneNumber     string
	Province        string
	City            string
	District        string
	DetailedAddress string
	PostalCode      string
	IsDefault       bool
}

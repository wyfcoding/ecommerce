package application

// DTO 定义

// LoginRequest 定义了请求参数结构。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 定义了响应数据结构。
type LoginResponse struct {
	Token    string   `json:"token"` // TODO: 实现 JWT
	UserInfo UserInfo `json:"userInfo"`
}

// UserInfo 结构体定义。
type UserInfo struct {
	ID          uint     `json:"id"`
	Username    string   `json:"username"`
	FullName    string   `json:"fullName"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// CreateUserRequest 定义了请求参数结构。
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"fullName"`
	Roles    []uint `json:"roles"` // 角色 ID 列表
}

// ApprovalCreateRequest 定义了请求参数结构。
type ApprovalCreateRequest struct {
	ActionType  string `json:"actionType" binding:"required"`
	Description string `json:"description"`
	Payload     string `json:"payload" binding:"required"` // JSON 字符串
}

// ApprovalActionRequest 定义了请求参数结构。
type ApprovalActionRequest struct {
	Action  string `json:"action" binding:"required,oneof=approve reject"`
	Comment string `json:"comment"`
}

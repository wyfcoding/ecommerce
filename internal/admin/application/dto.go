package application

// DTO 定义

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string   `json:"token"` // TODO: Implement JWT
	UserInfo UserInfo `json:"userInfo"`
}

type UserInfo struct {
	ID          uint     `json:"id"`
	Username    string   `json:"username"`
	FullName    string   `json:"fullName"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"fullName"`
	Roles    []uint `json:"roles"` // Role IDs
}

type ApprovalCreateRequest struct {
	ActionType  string `json:"actionType" binding:"required"`
	Description string `json:"description"`
	Payload     string `json:"payload" binding:"required"` // JSON string
}

type ApprovalActionRequest struct {
	Action  string `json:"action" binding:"required,oneof=approve reject"`
	Comment string `json:"comment"`
}

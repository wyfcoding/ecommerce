// 变更说明：新增身份管理功能，支持多维度用户画像、第三方账号关联、安全会话管理。
// 假设：用户画像信息为可选，第三方账号关联默认支持OAuth2流程。
package domain

import (
	"context"
	"time"
)

// --- 用户画像 (User Persona) ---

// PersonaType 画像类型
type PersonaType string

const (
	PersonaConsumer     PersonaType = "CONSUMER"     // 普通消费者
	PersonaMerchant     PersonaType = "MERCHANT"     // 商家/企业
	PersonaInfluencer   PersonaType = "INFLUENCER"   // 达人/博主
	PersonaProfessional PersonaType = "PROFESSIONAL" // 专业人士（如设计师、技工）
)

// UserPersona 用户画像聚合
type UserPersona struct {
	ID            uint64      `json:"id"`
	UserID        uint64      `json:"user_id"`
	Type          PersonaType `json:"type"`
	Tags          []string    `json:"tags"`           // 业务标签
	Interests     []string    `json:"interests"`      // 兴趣爱好
	Professional  *ProfInfo   `json:"professional"`   // 专业信息
	ShoppingHabit *HabitInfo  `json:"shopping_habit"` // 购物习惯
	SocialMetrics *SocialInfo `json:"social_metrics"` // 社交指标
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// ProfInfo 专业信息
type ProfInfo struct {
	Industry     string   `json:"industry"`     // 行业
	Skills       []string `json:"skills"`       // 技能
	Certificates []string `json:"certificates"` // 证书
	Experience   int      `json:"experience"`   // 经验年数
}

// HabitInfo 购物习惯
type HabitInfo struct {
	AvgOrderValue int64    `json:"avg_order_value"` // 平均客单价
	FreqCategory  []string `json:"freq_category"`   // 高频分类
	PreferredTime string   `json:"preferred_time"`  // 偏好下单时间
}

// SocialInfo 社交指标
type SocialInfo struct {
	Followers  int32   `json:"followers"`  // 粉丝数
	Following  int32   `json:"following"`  // 关注数
	Engagement float64 `json:"engagement"` // 互动率
}

// --- 关联账号 (Linked Account) ---

// AccountProvider 账号提供商
type AccountProvider string

const (
	ProviderGoogle AccountProvider = "GOOGLE"
	ProviderGitHub AccountProvider = "GITHUB"
	ProviderWeChat AccountProvider = "WECHAT"
	ProviderApple  AccountProvider = "APPLE"
)

// LinkedAccount 关联账号实体
type LinkedAccount struct {
	ID            uint64          `json:"id"`
	UserID        uint64          `json:"user_id"`
	Provider      AccountProvider `json:"provider"`
	ExternalID    string          `json:"external_id"`    // 第三方系统唯一ID
	ExternalName  string          `json:"external_name"`  // 第三方系统显示名
	ExternalEmail string          `json:"external_email"` // 第三方系统邮箱
	AccessToken   string          `json:"-"`              // 访问令牌
	RefreshToken  string          `json:"-"`              // 刷新令牌
	Expiry        *time.Time      `json:"expiry"`         // 令牌过期时间
	AvatarURL     string          `json:"avatar_url"`
	BoundAt       time.Time       `json:"bound_at"`
}

// --- 安全会话 (Security Session) ---

// SessionStatus 会话状态
type SessionStatus int

const (
	SessionActive  SessionStatus = 1
	SessionExpired SessionStatus = 2
	SessionRevoked SessionStatus = 3 // 已吊销（如安全退出或管理员剔除）
)

// AuthSession 认证会话实体
type AuthSession struct {
	ID           string        `json:"id"`
	UserID       uint64        `json:"user_id"`
	ClientID     string        `json:"client_id"`   // 客户端ID（Web/iOS/Android）
	DeviceID     string        `json:"device_id"`   // 设备唯一标识
	DeviceName   string        `json:"device_name"` // 设备名称
	IPAddress    string        `json:"ip_address"`  // 登录IP
	Location     string        `json:"location"`    // 地理位置
	UserAgent    string        `json:"user_agent"`  // 浏览器标识
	Status       SessionStatus `json:"status"`
	LastActiveAt time.Time     `json:"last_active_at"`
	CreatedAt    time.Time     `json:"created_at"`
	ExpiresAt    time.Time     `json:"expires_at"`
}

// IsValid 检查会话是否有效
func (s *AuthSession) IsValid() bool {
	return s.Status == SessionActive && time.Now().Before(s.ExpiresAt)
}

// Revoke 吊销会话
func (s *AuthSession) Revoke() {
	s.Status = SessionRevoked
}

// --- 身份仓储接口 ---

// IdentityRepository 身份仓储接口
type IdentityRepository interface {
	FindPersonaByUserID(ctx context.Context, userID uint64) (*UserPersona, error)
	SavePersona(ctx context.Context, persona *UserPersona) error

	FindLinkedAccounts(ctx context.Context, userID uint64) ([]*LinkedAccount, error)
	BindAccount(ctx context.Context, account *LinkedAccount) error
	UnbindAccount(ctx context.Context, userID uint64, provider AccountProvider) error

	FindSession(ctx context.Context, sessionID string) (*AuthSession, error)
	FindSessionsByUserID(ctx context.Context, userID uint64) ([]*AuthSession, error)
	SaveSession(ctx context.Context, session *AuthSession) error
}

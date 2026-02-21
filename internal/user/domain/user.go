// 变更说明：
// 1. 【增强】User 聚合根大幅扩展，新增手机号、密码哈希、头像、昵称、性别、生日等基础字段。
// 2. 【合并】从 identity/kyc 服务合并 KYC 实名认证和 MFA 多因素认证能力。
// 3. 【增强】新增用户状态管理（正常/锁定/禁用/注销）。
// 4. 【增强】新增登录安全（失败次数、锁定时间）。
// 5. 【增强】新增 OAuth 第三方登录绑定。
// 6. 【增强】Address 值对象增加省市区街道详细地址和经纬度。
package domain

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// 用户模块业务错误定义。
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserLocked         = errors.New("user account is locked")
	ErrUserDisabled       = errors.New("user account is disabled")
	ErrUserDeleted        = errors.New("user account has been deleted")
	ErrPhoneAlreadyBound  = errors.New("phone number already bound to another account")
	ErrEmailAlreadyBound  = errors.New("email already bound to another account")
)

// UserStatus 用户状态。
type UserStatus int8

const (
	UserStatusActive   UserStatus = 1 // 正常。
	UserStatusLocked   UserStatus = 2 // 锁定（登录失败过多）。
	UserStatusDisabled UserStatus = 3 // 禁用（管理员操作）。
	UserStatusDeleted  UserStatus = 4 // 注销（用户主动注销）。
)

// Gender 性别。
type Gender int8

const (
	GenderUnknown Gender = 0 // 未知。
	GenderMale    Gender = 1 // 男。
	GenderFemale  Gender = 2 // 女。
)

// User 用户聚合根。
// 整合了身份认证(identity)、实名认证(kyc)、用户画像(userprofile)的核心能力。
// 并发控制策略：乐观锁 (Version 字段)。
type User struct {
	gorm.Model
	// --- 基础信息 ---
	Username     string     `gorm:"column:username;type:varchar(64);uniqueIndex;not null" json:"username"` // 用户名，唯一。
	Email        string     `gorm:"column:email;type:varchar(255);uniqueIndex;not null" json:"email"`      // 邮箱，唯一。
	Phone        string     `gorm:"column:phone;type:varchar(20);uniqueIndex" json:"phone"`                // 手机号，唯一。
	FullName     string     `gorm:"column:full_name;type:varchar(128)" json:"full_name"`                   // 兼容旧用户模型。
	PasswordHash string     `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`              // 密码哈希（不序列化）。
	Password     string     `gorm:"-" json:"-"`                                                            // 兼容旧逻辑别名，运行期与 PasswordHash 保持一致。
	Nickname     string     `gorm:"column:nickname;type:varchar(64)" json:"nickname"`                      // 昵称。
	Avatar       string     `gorm:"column:avatar;type:varchar(512)" json:"avatar"`                         // 头像 URL。
	Gender       Gender     `gorm:"column:gender;type:tinyint;default:0" json:"gender"`                    // 性别。
	Birthday     *time.Time `gorm:"column:birthday;type:date" json:"birthday"`                             // 生日。
	Bio          string     `gorm:"column:bio;type:varchar(500)" json:"bio"`                               // 个人简介。
	Status       UserStatus `gorm:"column:status;type:tinyint;default:1;index" json:"status"`              // 用户状态。
	// --- 安全信息 ---
	FailedLoginAttempts int32      `gorm:"column:failed_login_attempts;default:0" json:"-"` // 连续登录失败次数。
	LockedUntil         *time.Time `gorm:"column:locked_until" json:"-"`                    // 锁定截止时间。
	LastLoginAt         *time.Time `gorm:"column:last_login_at" json:"last_login_at"`       // 最后登录时间。
	LastLoginIP         string     `gorm:"column:last_login_ip;type:varchar(45)" json:"-"`  // 最后登录 IP。
	PasswordChangedAt   *time.Time `gorm:"column:password_changed_at" json:"-"`             // 密码最后修改时间。
	// --- 关联信息 ---
	Addresses  []*Address  `gorm:"foreignKey:UserID" json:"addresses"`   // 收货地址列表。
	OAuthBinds []OAuthBind `gorm:"foreignKey:UserID" json:"oauth_binds"` // 第三方登录绑定。
	// --- 扩展信息（非 GORM 持久化，由值对象管理）---
	KYC *KYCInfo   `gorm:"-" json:"kyc,omitempty"` // 实名认证信息。
	MFA *MFAConfig `gorm:"-" json:"mfa,omitempty"` // 多因素认证配置。
	// --- 版本控制 ---
	Version int64 `gorm:"column:version;default:0" json:"version"` // 乐观锁版本号。
}

// Address 收货地址值对象。
type Address struct {
	gorm.Model
	UserID    uint    `gorm:"column:user_id;index;not null" json:"user_id"`         // 用户 ID。
	Contact   string  `gorm:"column:contact;type:varchar(64)" json:"contact"`       // 联系人姓名。
	Phone     string  `gorm:"column:phone;type:varchar(20)" json:"phone"`           // 联系电话。
	Province  string  `gorm:"column:province;type:varchar(32)" json:"province"`     // 省份。
	City      string  `gorm:"column:city;type:varchar(32)" json:"city"`             // 城市。
	District  string  `gorm:"column:district;type:varchar(32)" json:"district"`     // 区/县。
	Street    string  `gorm:"column:street;type:varchar(128)" json:"street"`        // 街道。
	Detail    string  `gorm:"column:detail;type:varchar(256)" json:"detail"`        // 详细地址。
	ZipCode   string  `gorm:"column:zip_code;type:varchar(10)" json:"zip_code"`     // 邮编。
	Latitude  float64 `gorm:"column:latitude;type:decimal(10,7)" json:"latitude"`   // 纬度。
	Longitude float64 `gorm:"column:longitude;type:decimal(10,7)" json:"longitude"` // 经度。
	Label     string  `gorm:"column:label;type:varchar(20)" json:"label"`           // 标签（家/公司/学校）。
	IsDefault bool    `gorm:"column:is_default;default:false" json:"is_default"`    // 是否默认地址。

	// 兼容旧应用层字段（不额外落库，统一复用上面的领域字段）。
	RecipientName   string `gorm:"-" json:"recipient_name,omitempty"`
	PhoneNumber     string `gorm:"-" json:"phone_number,omitempty"`
	DetailedAddress string `gorm:"-" json:"detailed_address,omitempty"`
	PostalCode      string `gorm:"-" json:"postal_code,omitempty"`
}

// NewUser 创建用户实体（兼容旧命令服务构造参数）。
func NewUser(username, email, password, phone string) (*User, error) {
	if username == "" || email == "" || password == "" {
		return nil, errors.New("username/email/password are required")
	}
	now := time.Now()
	return &User{
		Model: gorm.Model{
			CreatedAt: now,
			UpdatedAt: now,
		},
		Username:     username,
		Email:        email,
		Phone:        phone,
		Password:     password,
		PasswordHash: password,
		Status:       UserStatusActive,
	}, nil
}

// UpdateProfile 更新用户基础资料。
func (u *User) UpdateProfile(nickname, avatar string, gender int8, birthday *time.Time) {
	u.Nickname = nickname
	u.Avatar = avatar
	u.Gender = Gender(gender)
	u.Birthday = birthday
	u.UpdatedAt = time.Now()
}

// NewAddress 创建收货地址（兼容旧命令服务构造参数）。
func NewAddress(userID uint, recipientName, phoneNumber, province, city, district, detailedAddress, postalCode string, isDefault bool) *Address {
	now := time.Now()
	return &Address{
		Model: gorm.Model{
			CreatedAt: now,
			UpdatedAt: now,
		},
		UserID:          userID,
		Contact:         recipientName,
		Phone:           phoneNumber,
		Province:        province,
		City:            city,
		District:        district,
		Detail:          detailedAddress,
		ZipCode:         postalCode,
		IsDefault:       isDefault,
		RecipientName:   recipientName,
		PhoneNumber:     phoneNumber,
		DetailedAddress: detailedAddress,
		PostalCode:      postalCode,
	}
}

// OAuthBind 第三方登录绑定。
type OAuthBind struct {
	gorm.Model
	UserID   uint      `gorm:"column:user_id;index;not null" json:"user_id"`                                    // 用户 ID。
	Provider string    `gorm:"column:provider;type:varchar(32);not null" json:"provider"`                       // 提供商（wechat/alipay/google/apple）。
	OpenID   string    `gorm:"column:open_id;type:varchar(128);uniqueIndex:idx_provider_openid" json:"open_id"` // 第三方用户 ID。
	UnionID  string    `gorm:"column:union_id;type:varchar(128)" json:"union_id"`                               // 第三方统一 ID。
	Nickname string    `gorm:"column:nickname;type:varchar(64)" json:"nickname"`                                // 第三方昵称。
	Avatar   string    `gorm:"column:avatar;type:varchar(512)" json:"avatar"`                                   // 第三方头像。
	BindAt   time.Time `gorm:"column:bind_at" json:"bind_at"`                                                   // 绑定时间。
}

// AddAddress 添加收货地址。
// 如果设为默认地址，则取消其他地址的默认标记。
func (u *User) AddAddress(addr Address) {
	address := addr
	if addr.IsDefault {
		for i := range u.Addresses {
			if u.Addresses[i] != nil {
				u.Addresses[i].IsDefault = false
			}
		}
	}
	u.Addresses = append(u.Addresses, &address)
}

// SetDefaultAddress 设置默认地址。
func (u *User) SetDefaultAddress(addressID uint) error {
	found := false
	for i := range u.Addresses {
		if u.Addresses[i] != nil && u.Addresses[i].ID == addressID {
			u.Addresses[i].IsDefault = true
			found = true
		} else if u.Addresses[i] != nil {
			u.Addresses[i].IsDefault = false
		}
	}
	if !found {
		return errors.New("address not found")
	}
	return nil
}

// RecordLoginSuccess 记录登录成功。
func (u *User) RecordLoginSuccess(ip string) {
	now := time.Now()
	u.LastLoginAt = &now
	u.LastLoginIP = ip
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
}

// RecordLoginFailure 记录登录失败。
// 超过最大失败次数后自动锁定账户。
func (u *User) RecordLoginFailure(maxAttempts int32, lockDuration time.Duration) {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= maxAttempts {
		lockUntil := time.Now().Add(lockDuration)
		u.LockedUntil = &lockUntil
		u.Status = UserStatusLocked
	}
}

// IsLocked 判断用户是否被锁定。
func (u *User) IsLocked() bool {
	if u.Status != UserStatusLocked {
		return false
	}
	if u.LockedUntil != nil && time.Now().After(*u.LockedUntil) {
		// 锁定已过期，自动解锁。
		u.Status = UserStatusActive
		u.FailedLoginAttempts = 0
		u.LockedUntil = nil
		return false
	}
	return true
}

// IsActive 判断用户是否处于正常状态。
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// Disable 禁用用户。
func (u *User) Disable() {
	u.Status = UserStatusDisabled
}

// Enable 启用用户。
func (u *User) Enable() {
	u.Status = UserStatusActive
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
}

// SoftDelete 软删除（注销）。
func (u *User) SoftDelete() {
	u.Status = UserStatusDeleted
}

// BindOAuth 绑定第三方登录。
func (u *User) BindOAuth(provider, openID, unionID, nickname, avatar string) {
	// 检查是否已绑定。
	for i := range u.OAuthBinds {
		if u.OAuthBinds[i].Provider == provider {
			u.OAuthBinds[i].OpenID = openID
			u.OAuthBinds[i].UnionID = unionID
			u.OAuthBinds[i].Nickname = nickname
			u.OAuthBinds[i].Avatar = avatar
			return
		}
	}
	u.OAuthBinds = append(u.OAuthBinds, OAuthBind{
		UserID:   u.ID,
		Provider: provider,
		OpenID:   openID,
		UnionID:  unionID,
		Nickname: nickname,
		Avatar:   avatar,
		BindAt:   time.Now(),
	})
}

// UnbindOAuth 解绑第三方登录。
func (u *User) UnbindOAuth(provider string) error {
	for i := range u.OAuthBinds {
		if u.OAuthBinds[i].Provider == provider {
			u.OAuthBinds = append(u.OAuthBinds[:i], u.OAuthBinds[i+1:]...)
			return nil
		}
	}
	return errors.New("oauth provider not bound")
}

// 仓储接口定义位于 user_repository.go，避免同包重复声明。

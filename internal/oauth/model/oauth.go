package model

import "time"

// OAuthProvider 第三方登录提供商
type OAuthProvider string

const (
	OAuthProviderWechat   OAuthProvider = "WECHAT"   // 微信
	OAuthProviderQQ       OAuthProvider = "QQ"       // QQ
	OAuthProviderAlipay   OAuthProvider = "ALIPAY"   // 支付宝
	OAuthProviderApple    OAuthProvider = "APPLE"    // Apple
	OAuthProviderWeibo    OAuthProvider = "WEIBO"    // 微博
	OAuthProviderGithub   OAuthProvider = "GITHUB"   // GitHub
	OAuthProviderGoogle   OAuthProvider = "GOOGLE"   // Google
	OAuthProviderFacebook OAuthProvider = "FACEBOOK" // Facebook
)

// UserOAuth 用户第三方账号绑定
type UserOAuth struct {
	ID           uint64        `gorm:"primarykey" json:"id"`
	UserID       uint64        `gorm:"index:idx_user_provider;not null;comment:用户ID" json:"userId"`
	Provider     OAuthProvider `gorm:"type:varchar(20);index:idx_user_provider;not null;comment:提供商" json:"provider"`
	OpenID       string        `gorm:"type:varchar(255);uniqueIndex:idx_provider_openid;not null;comment:第三方OpenID" json:"openId"`
	UnionID      string        `gorm:"type:varchar(255);index;comment:第三方UnionID" json:"unionId"`
	Nickname     string        `gorm:"type:varchar(255);comment:第三方昵称" json:"nickname"`
	Avatar       string        `gorm:"type:varchar(500);comment:第三方头像" json:"avatar"`
	AccessToken  string        `gorm:"type:varchar(500);comment:访问令牌" json:"accessToken"`
	RefreshToken string        `gorm:"type:varchar(500);comment:刷新令牌" json:"refreshToken"`
	ExpiresAt    *time.Time    `gorm:"comment:令牌过期时间" json:"expiresAt"`
	RawData      string        `gorm:"type:text;comment:原始数据JSON" json:"rawData"`
	CreatedAt    time.Time     `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time     `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt    *time.Time    `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (UserOAuth) TableName() string {
	return "user_oauth"
}

// OAuthConfig 第三方登录配置
type OAuthConfig struct {
	Provider     OAuthProvider
	AppID        string
	AppSecret    string
	RedirectURL  string
	Scope        string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

// OAuthUserInfo 第三方用户信息
type OAuthUserInfo struct {
	OpenID   string
	UnionID  string
	Nickname string
	Avatar   string
	Gender   string
	Province string
	City     string
	Country  string
}

// OAuthState OAuth状态（用于防止CSRF攻击）
type OAuthState struct {
	ID        uint64     `gorm:"primarykey" json:"id"`
	State     string     `gorm:"type:varchar(64);uniqueIndex;not null;comment:状态码" json:"state"`
	Provider  OAuthProvider `gorm:"type:varchar(20);not null;comment:提供商" json:"provider"`
	UserID    uint64     `gorm:"comment:用户ID(绑定时使用)" json:"userId"`
	ExpiresAt time.Time  `gorm:"not null;comment:过期时间" json:"expiresAt"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (OAuthState) TableName() string {
	return "oauth_states"
}

// IsExpired 判断是否过期
func (os *OAuthState) IsExpired() bool {
	return time.Now().After(os.ExpiresAt)
}

// WechatAccessTokenResponse 微信AccessToken响应
type WechatAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// WechatUserInfoResponse 微信用户信息响应
type WechatUserInfoResponse struct {
	OpenID     string   `json:"openid"`
	UnionID    string   `json:"unionid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	ErrCode    int      `json:"errcode"`
	ErrMsg     string   `json:"errmsg"`
}

// QQAccessTokenResponse QQ AccessToken响应
type QQAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// QQOpenIDResponse QQ OpenID响应
type QQOpenIDResponse struct {
	ClientID string `json:"client_id"`
	OpenID   string `json:"openid"`
}

// QQUserInfoResponse QQ用户信息响应
type QQUserInfoResponse struct {
	Ret              int    `json:"ret"`
	Msg              string `json:"msg"`
	Nickname         string `json:"nickname"`
	FigureURL        string `json:"figureurl"`
	FigureURL1       string `json:"figureurl_1"`
	FigureURL2       string `json:"figureurl_2"`
	FigureURLQQ1     string `json:"figureurl_qq_1"`
	FigureURLQQ2     string `json:"figureurl_qq_2"`
	Gender           string `json:"gender"`
	IsYellowVip      string `json:"is_yellow_vip"`
	Vip              string `json:"vip"`
	YellowVipLevel   string `json:"yellow_vip_level"`
	Level            string `json:"level"`
	IsYellowYearVip  string `json:"is_yellow_year_vip"`
}

// AlipayAccessTokenResponse 支付宝AccessToken响应
type AlipayAccessTokenResponse struct {
	AlipaySystemOauthTokenResponse struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		ReExpiresIn  int    `json:"re_expires_in"`
		UserID       string `json:"user_id"`
	} `json:"alipay_system_oauth_token_response"`
	Sign string `json:"sign"`
}

// AlipayUserInfoResponse 支付宝用户信息响应
type AlipayUserInfoResponse struct {
	AlipayUserInfoShareResponse struct {
		Code     string `json:"code"`
		Msg      string `json:"msg"`
		UserID   string `json:"user_id"`
		Avatar   string `json:"avatar"`
		Province string `json:"province"`
		City     string `json:"city"`
		NickName string `json:"nick_name"`
		Gender   string `json:"gender"`
	} `json:"alipay_user_info_share_response"`
	Sign string `json:"sign"`
}

// AppleIDTokenClaims Apple ID Token Claims
type AppleIDTokenClaims struct {
	Issuer         string `json:"iss"`
	Subject        string `json:"sub"`
	Audience       string `json:"aud"`
	IssuedAt       int64  `json:"iat"`
	ExpirationTime int64  `json:"exp"`
	Email          string `json:"email"`
	EmailVerified  string `json:"email_verified"`
}

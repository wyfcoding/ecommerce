package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

var (
	ErrOAuthAccountAlreadyLinked = errors.New("oauth account already linked to another user")
	ErrOAuthAccountNotFound      = errors.New("oauth account not found")
	ErrUnsupportedOAuthProvider  = errors.New("unsupported oauth provider")
)

type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
	OAuthProviderWechat OAuthProvider = "wechat"
	OAuthProviderApple  OAuthProvider = "apple"
	OAuthProviderGithub OAuthProvider = "github"
	OAuthProviderQQ     OAuthProvider = "qq"
	OAuthProviderWeibo  OAuthProvider = "weibo"
)

func (p OAuthProvider) IsValid() bool {
	switch p {
	case OAuthProviderGoogle, OAuthProviderWechat, OAuthProviderApple,
		OAuthProviderGithub, OAuthProviderQQ, OAuthProviderWeibo:
		return true
	default:
		return false
	}
}

type OAuthAccount struct {
	ID           uint          `json:"id"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	UserID       uint64        `json:"user_id"`
	Provider     OAuthProvider `json:"provider"`
	ProviderID   string        `json:"provider_id"`
	UnionID      string        `json:"union_id"`
	OpenID       string        `json:"open_id"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresAt    *time.Time    `json:"expires_at"`
	Nickname     string        `json:"nickname"`
	Avatar       string        `json:"avatar"`
	Email        string        `json:"email"`
	Phone        string        `json:"phone"`
	Gender       int8          `json:"gender"`
	Location     string        `json:"location"`
	RawData      string        `json:"raw_data"`
}

type OAuthState struct {
	ID          uint          `json:"id"`
	CreatedAt   time.Time     `json:"created_at"`
	State       string        `json:"state"`
	Provider    OAuthProvider `json:"provider"`
	RedirectURL string        `json:"redirect_url"`
	ExpiresAt   time.Time     `json:"expires_at"`
	Used        bool          `json:"used"`
}

type OAuthToken struct {
	AccessToken  string     `json:"access_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int64      `json:"expires_in"`
	RefreshToken string     `json:"refresh_token"`
	Scope        string     `json:"scope"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

type OAuthUserInfo struct {
	Provider   OAuthProvider `json:"provider"`
	ProviderID string        `json:"provider_id"`
	UnionID    string        `json:"union_id"`
	OpenID     string        `json:"open_id"`
	Nickname   string        `json:"nickname"`
	Avatar     string        `json:"avatar"`
	Email      string        `json:"email"`
	Phone      string        `json:"phone"`
	Gender     int8          `json:"gender"`
	Location   string        `json:"location"`
	RawData    string        `json:"raw_data"`
}

func NewOAuthAccount(userID uint64, provider OAuthProvider, providerID string) *OAuthAccount {
	now := time.Now()
	return &OAuthAccount{
		UserID:     userID,
		Provider:   provider,
		ProviderID: providerID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (a *OAuthAccount) UpdateTokens(accessToken, refreshToken string, expiresIn int64) {
	a.AccessToken = accessToken
	a.RefreshToken = refreshToken
	if expiresIn > 0 {
		exp := time.Now().Add(time.Duration(expiresIn) * time.Second)
		a.ExpiresAt = &exp
	}
	a.UpdatedAt = time.Now()
}

func (a *OAuthAccount) UpdateProfile(nickname, avatar, email, phone string, gender int8, location string) {
	if nickname != "" {
		a.Nickname = nickname
	}
	if avatar != "" {
		a.Avatar = avatar
	}
	if email != "" {
		a.Email = email
	}
	if phone != "" {
		a.Phone = phone
	}
	if gender > 0 {
		a.Gender = gender
	}
	if location != "" {
		a.Location = location
	}
	a.UpdatedAt = time.Now()
}

func (a *OAuthAccount) IsTokenExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

func (a *OAuthAccount) IsTokenExpiringSoon(threshold time.Duration) bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().Add(threshold).After(*a.ExpiresAt)
}

func NewOAuthState(provider OAuthProvider, redirectURL string) *OAuthState {
	state := generateOAuthState()
	return &OAuthState{
		State:       state,
		Provider:    provider,
		RedirectURL: redirectURL,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		Used:        false,
		CreatedAt:   time.Now(),
	}
}

func (s *OAuthState) IsValid() bool {
	return !s.Used && time.Now().Before(s.ExpiresAt)
}

func (s *OAuthState) MarkUsed() {
	s.Used = true
}

func generateOAuthState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type OAuthRepository interface {
	FindByProviderID(ctx any, provider OAuthProvider, providerID string) (*OAuthAccount, error)
	FindByUserID(ctx any, userID uint64) ([]*OAuthAccount, error)
	FindByUserIDAndProvider(ctx any, userID uint64, provider OAuthProvider) (*OAuthAccount, error)
	Save(ctx any, account *OAuthAccount) error
	Update(ctx any, account *OAuthAccount) error
	Delete(ctx any, id uint) error

	SaveState(ctx any, state *OAuthState) error
	FindState(ctx any, state string) (*OAuthState, error)
	DeleteExpiredStates(ctx any) error
}

type OAuthProviderAdapter interface {
	GetAuthURL(state string) string
	ExchangeCode(code string) (*OAuthToken, error)
	GetUserInfo(token *OAuthToken) (*OAuthUserInfo, error)
	RefreshToken(refreshToken string) (*OAuthToken, error)
	GetProvider() OAuthProvider
}

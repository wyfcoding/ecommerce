package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"ecommerce/internal/oauth/model"
	"ecommerce/internal/oauth/repository"
	userModel "ecommerce/internal/user/model"
	userRepo "ecommerce/internal/user/repository"
)

var (
	ErrInvalidProvider    = errors.New("无效的登录提供商")
	ErrInvalidState       = errors.New("无效的state参数")
	ErrStateExpired       = errors.New("state已过期")
	ErrOAuthFailed        = errors.New("第三方登录失败")
	ErrAccountNotBound    = errors.New("账号未绑定")
	ErrAccountAlreadyBound = errors.New("账号已绑定")
)

// OAuthService 第三方登录服务接口
type OAuthService interface {
	// 获取授权URL
	GetAuthURL(ctx context.Context, provider model.OAuthProvider, redirectURL string, userID uint64) (string, error)
	
	// 处理回调
	HandleCallback(ctx context.Context, provider model.OAuthProvider, code, state string) (*userModel.User, bool, error)
	
	// 绑定账号
	BindAccount(ctx context.Context, userID uint64, provider model.OAuthProvider, code, state string) error
	
	// 解绑账号
	UnbindAccount(ctx context.Context, userID uint64, provider model.OAuthProvider) error
	
	// 获取用户绑定列表
	GetUserBindings(ctx context.Context, userID uint64) ([]*model.UserOAuth, error)
}

type oauthService struct {
	repo       repository.OAuthRepo
	userRepo   userRepo.UserRepo
	configs    map[model.OAuthProvider]*model.OAuthConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// NewOAuthService 创建第三方登录服务实例
func NewOAuthService(
	repo repository.OAuthRepo,
	userRepo userRepo.UserRepo,
	configs map[model.OAuthProvider]*model.OAuthConfig,
	logger *zap.Logger,
) OAuthService {
	return &oauthService{
		repo:     repo,
		userRepo: userRepo,
		configs:  configs,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// GetAuthURL 获取授权URL
func (s *oauthService) GetAuthURL(ctx context.Context, provider model.OAuthProvider, redirectURL string, userID uint64) (string, error) {
	config, ok := s.configs[provider]
	if !ok {
		return "", ErrInvalidProvider
	}

	// 生成state
	state, err := s.generateState(ctx, provider, userID)
	if err != nil {
		return "", err
	}

	// 构建授权URL
	params := url.Values{}
	
	switch provider {
	case model.OAuthProviderWechat:
		params.Set("appid", config.AppID)
		params.Set("redirect_uri", redirectURL)
		params.Set("response_type", "code")
		params.Set("scope", config.Scope)
		params.Set("state", state)
		return fmt.Sprintf("%s?%s#wechat_redirect", config.AuthURL, params.Encode()), nil
		
	case model.OAuthProviderQQ:
		params.Set("client_id", config.AppID)
		params.Set("redirect_uri", redirectURL)
		params.Set("response_type", "code")
		params.Set("scope", config.Scope)
		params.Set("state", state)
		return fmt.Sprintf("%s?%s", config.AuthURL, params.Encode()), nil
		
	case model.OAuthProviderAlipay:
		params.Set("app_id", config.AppID)
		params.Set("redirect_uri", redirectURL)
		params.Set("scope", config.Scope)
		params.Set("state", state)
		return fmt.Sprintf("%s?%s", config.AuthURL, params.Encode()), nil
		
	default:
		return "", ErrInvalidProvider
	}
}

// HandleCallback 处理回调
func (s *oauthService) HandleCallback(ctx context.Context, provider model.OAuthProvider, code, state string) (*userModel.User, bool, error) {
	// 1. 验证state
	oauthState, err := s.verifyState(ctx, state)
	if err != nil {
		return nil, false, err
	}

	if oauthState.Provider != provider {
		return nil, false, ErrInvalidProvider
	}

	// 2. 获取AccessToken
	accessToken, openID, err := s.getAccessToken(ctx, provider, code)
	if err != nil {
		return nil, false, err
	}

	// 3. 获取用户信息
	userInfo, err := s.getUserInfo(ctx, provider, accessToken, openID)
	if err != nil {
		return nil, false, err
	}

	// 4. 查找是否已绑定
	userOAuth, err := s.repo.GetByProviderAndOpenID(ctx, provider, userInfo.OpenID)
	if err == nil {
		// 已绑定，直接登录
		user, err := s.userRepo.GetUserByID(ctx, userOAuth.UserID)
		if err != nil {
			return nil, false, err
		}
		
		// 更新OAuth信息
		s.updateUserOAuth(ctx, userOAuth, accessToken, userInfo)
		
		return user, false, nil
	}

	// 5. 未绑定，创建新用户
	user := &userModel.User{
		Username: fmt.Sprintf("%s_%s", strings.ToLower(string(provider)), userInfo.OpenID[:8]),
		Nickname: userInfo.Nickname,
		Avatar:   userInfo.Avatar,
		Status:   userModel.UserStatusActive,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, false, err
	}

	// 6. 创建绑定关系
	userOAuth = &model.UserOAuth{
		UserID:      user.ID,
		Provider:    provider,
		OpenID:      userInfo.OpenID,
		UnionID:     userInfo.UnionID,
		Nickname:    userInfo.Nickname,
		Avatar:      userInfo.Avatar,
		AccessToken: accessToken,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateUserOAuth(ctx, userOAuth); err != nil {
		return nil, false, err
	}

	s.logger.Info("第三方登录成功",
		zap.String("provider", string(provider)),
		zap.Uint64("userID", user.ID))

	return user, true, nil
}

// BindAccount 绑定账号
func (s *oauthService) BindAccount(ctx context.Context, userID uint64, provider model.OAuthProvider, code, state string) error {
	// 1. 验证state
	oauthState, err := s.verifyState(ctx, state)
	if err != nil {
		return err
	}

	if oauthState.Provider != provider || oauthState.UserID != userID {
		return ErrInvalidState
	}

	// 2. 检查是否已绑定
	existing, err := s.repo.GetByUserIDAndProvider(ctx, userID, provider)
	if err == nil && existing != nil {
		return ErrAccountAlreadyBound
	}

	// 3. 获取AccessToken
	accessToken, openID, err := s.getAccessToken(ctx, provider, code)
	if err != nil {
		return err
	}

	// 4. 获取用户信息
	userInfo, err := s.getUserInfo(ctx, provider, accessToken, openID)
	if err != nil {
		return err
	}

	// 5. 检查OpenID是否已被其他用户绑定
	existingOAuth, err := s.repo.GetByProviderAndOpenID(ctx, provider, userInfo.OpenID)
	if err == nil && existingOAuth != nil {
		return fmt.Errorf("该第三方账号已被其他用户绑定")
	}

	// 6. 创建绑定关系
	userOAuth := &model.UserOAuth{
		UserID:      userID,
		Provider:    provider,
		OpenID:      userInfo.OpenID,
		UnionID:     userInfo.UnionID,
		Nickname:    userInfo.Nickname,
		Avatar:      userInfo.Avatar,
		AccessToken: accessToken,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateUserOAuth(ctx, userOAuth); err != nil {
		return err
	}

	s.logger.Info("绑定账号成功",
		zap.Uint64("userID", userID),
		zap.String("provider", string(provider)))

	return nil
}

// UnbindAccount 解绑账号
func (s *oauthService) UnbindAccount(ctx context.Context, userID uint64, provider model.OAuthProvider) error {
	userOAuth, err := s.repo.GetByUserIDAndProvider(ctx, userID, provider)
	if err != nil {
		return ErrAccountNotBound
	}

	if err := s.repo.DeleteUserOAuth(ctx, userOAuth.ID); err != nil {
		s.logger.Error("解绑账号失败", zap.Error(err))
		return err
	}

	s.logger.Info("解绑账号成功",
		zap.Uint64("userID", userID),
		zap.String("provider", string(provider)))

	return nil
}

// GetUserBindings 获取用户绑定列表
func (s *oauthService) GetUserBindings(ctx context.Context, userID uint64) ([]*model.UserOAuth, error) {
	bindings, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return bindings, nil
}

// generateState 生成state
func (s *oauthService) generateState(ctx context.Context, provider model.OAuthProvider, userID uint64) (string, error) {
	// 生成随机state
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)

	// 保存state
	oauthState := &model.OAuthState{
		State:     state,
		Provider:  provider,
		UserID:    userID,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateOAuthState(ctx, oauthState); err != nil {
		return "", err
	}

	return state, nil
}

// verifyState 验证state
func (s *oauthService) verifyState(ctx context.Context, state string) (*model.OAuthState, error) {
	oauthState, err := s.repo.GetOAuthStateByState(ctx, state)
	if err != nil {
		return nil, ErrInvalidState
	}

	if oauthState.IsExpired() {
		return nil, ErrStateExpired
	}

	// 删除已使用的state
	s.repo.DeleteOAuthState(ctx, oauthState.ID)

	return oauthState, nil
}

// getAccessToken 获取AccessToken
func (s *oauthService) getAccessToken(ctx context.Context, provider model.OAuthProvider, code string) (string, string, error) {
	config := s.configs[provider]

	switch provider {
	case model.OAuthProviderWechat:
		return s.getWechatAccessToken(config, code)
	case model.OAuthProviderQQ:
		return s.getQQAccessToken(config, code)
	case model.OAuthProviderAlipay:
		return s.getAlipayAccessToken(config, code)
	default:
		return "", "", ErrInvalidProvider
	}
}

// getWechatAccessToken 获取微信AccessToken
func (s *oauthService) getWechatAccessToken(config *model.OAuthConfig, code string) (string, string, error) {
	params := url.Values{}
	params.Set("appid", config.AppID)
	params.Set("secret", config.AppSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")

	url := fmt.Sprintf("%s?%s", config.TokenURL, params.Encode())
	
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var result model.WechatAccessTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", err
	}

	if result.ErrCode != 0 {
		return "", "", fmt.Errorf("微信登录失败: %s", result.ErrMsg)
	}

	return result.AccessToken, result.OpenID, nil
}

// getQQAccessToken 获取QQ AccessToken
func (s *oauthService) getQQAccessToken(config *model.OAuthConfig, code string) (string, string, error) {
	// 1. 获取AccessToken
	params := url.Values{}
	params.Set("grant_type", "authorization_code")
	params.Set("client_id", config.AppID)
	params.Set("client_secret", config.AppSecret)
	params.Set("code", code)
	params.Set("redirect_uri", config.RedirectURL)

	tokenURL := fmt.Sprintf("%s?%s", config.TokenURL, params.Encode())
	
	resp, err := s.httpClient.Get(tokenURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	// 解析返回值（QQ返回的是URL参数格式）
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return "", "", err
	}

	accessToken := values.Get("access_token")
	if accessToken == "" {
		return "", "", fmt.Errorf("获取QQ AccessToken失败")
	}

	// 2. 获取OpenID
	openIDURL := fmt.Sprintf("https://graph.qq.com/oauth2.0/me?access_token=%s", accessToken)
	
	resp2, err := s.httpClient.Get(openIDURL)
	if err != nil {
		return "", "", err
	}
	defer resp2.Body.Close()

	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return "", "", err
	}

	// 解析OpenID（返回格式：callback( {"client_id":"YOUR_APPID","openid":"YOUR_OPENID"} );）
	bodyStr := string(body2)
	start := strings.Index(bodyStr, "{")
	end := strings.LastIndex(bodyStr, "}")
	if start == -1 || end == -1 {
		return "", "", fmt.Errorf("解析QQ OpenID失败")
	}

	var openIDResp model.QQOpenIDResponse
	if err := json.Unmarshal([]byte(bodyStr[start:end+1]), &openIDResp); err != nil {
		return "", "", err
	}

	return accessToken, openIDResp.OpenID, nil
}

// getAlipayAccessToken 获取支付宝AccessToken
func (s *oauthService) getAlipayAccessToken(config *model.OAuthConfig, code string) (string, string, error) {
	// TODO: 实现支付宝OAuth
	// 支付宝需要使用RSA签名，实现较复杂
	return "", "", fmt.Errorf("支付宝登录暂未实现")
}

// getUserInfo 获取用户信息
func (s *oauthService) getUserInfo(ctx context.Context, provider model.OAuthProvider, accessToken, openID string) (*model.OAuthUserInfo, error) {
	config := s.configs[provider]

	switch provider {
	case model.OAuthProviderWechat:
		return s.getWechatUserInfo(config, accessToken, openID)
	case model.OAuthProviderQQ:
		return s.getQQUserInfo(config, accessToken, openID)
	case model.OAuthProviderAlipay:
		return s.getAlipayUserInfo(config, accessToken)
	default:
		return nil, ErrInvalidProvider
	}
}

// getWechatUserInfo 获取微信用户信息
func (s *oauthService) getWechatUserInfo(config *model.OAuthConfig, accessToken, openID string) (*model.OAuthUserInfo, error) {
	url := fmt.Sprintf("%s?access_token=%s&openid=%s&lang=zh_CN", config.UserInfoURL, accessToken, openID)
	
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result model.WechatUserInfoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("获取微信用户信息失败: %s", result.ErrMsg)
	}

	return &model.OAuthUserInfo{
		OpenID:   result.OpenID,
		UnionID:  result.UnionID,
		Nickname: result.Nickname,
		Avatar:   result.HeadImgURL,
		Province: result.Province,
		City:     result.City,
		Country:  result.Country,
	}, nil
}

// getQQUserInfo 获取QQ用户信息
func (s *oauthService) getQQUserInfo(config *model.OAuthConfig, accessToken, openID string) (*model.OAuthUserInfo, error) {
	url := fmt.Sprintf("%s?access_token=%s&oauth_consumer_key=%s&openid=%s",
		config.UserInfoURL, accessToken, config.AppID, openID)
	
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result model.QQUserInfoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Ret != 0 {
		return nil, fmt.Errorf("获取QQ用户信息失败: %s", result.Msg)
	}

	avatar := result.FigureURLQQ2
	if avatar == "" {
		avatar = result.FigureURLQQ1
	}
	if avatar == "" {
		avatar = result.FigureURL2
	}

	return &model.OAuthUserInfo{
		OpenID:   openID,
		Nickname: result.Nickname,
		Avatar:   avatar,
		Gender:   result.Gender,
	}, nil
}

// getAlipayUserInfo 获取支付宝用户信息
func (s *oauthService) getAlipayUserInfo(config *model.OAuthConfig, accessToken string) (*model.OAuthUserInfo, error) {
	// TODO: 实现支付宝用户信息获取
	return nil, fmt.Errorf("支付宝登录暂未实现")
}

// updateUserOAuth 更新OAuth信息
func (s *oauthService) updateUserOAuth(ctx context.Context, userOAuth *model.UserOAuth, accessToken string, userInfo *model.OAuthUserInfo) {
	userOAuth.AccessToken = accessToken
	userOAuth.Nickname = userInfo.Nickname
	userOAuth.Avatar = userInfo.Avatar
	userOAuth.UpdatedAt = time.Now()

	s.repo.UpdateUserOAuth(ctx, userOAuth)
}

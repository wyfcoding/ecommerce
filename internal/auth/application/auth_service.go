package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/auth/domain"
	"github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/jwt"
	"github.com/wyfcoding/pkg/xerrors"
)

// 生成摘要：Auth 应用服务实现。
// 关键改动：封装令牌签发逻辑，集成 JWT 配置。

type AuthService struct {
	repo domain.Repository
	cfg  *config.Config
}

func NewAuthService(repo domain.Repository, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

type LoginResult struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	// 1. 获取凭证
	auth, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, xerrors.Unauthenticated("user not found")
	}

	// 2. 校验密码
	if err := auth.CheckPassword(password); err != nil {
		_ = s.repo.Save(ctx, auth)
		return nil, xerrors.Unauthenticated(err.Error())
	}

	// 3. 异步持久化更新状态
	_ = s.repo.Save(ctx, auth)

	// 4. 签发 JWT (使用 pkg/jwt)
	token, err := jwt.GenerateToken(
		auth.UserID,
		auth.Username,
		[]string{"USER"}, // 缺省角色
		s.cfg.JWT.Secret,
		s.cfg.JWT.Issuer,
		s.cfg.JWT.ExpireDuration,
	)
	if err != nil {
		return nil, xerrors.Internal("token generation failed", err)
	}

	return &LoginResult{
		AccessToken: token,
		ExpiresIn:   int64(s.cfg.JWT.ExpireDuration.Seconds()),
	}, nil
}

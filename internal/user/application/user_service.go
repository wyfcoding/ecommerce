package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/pkg/xerrors"
)

// 生成摘要：用户应用服务。
// 关键改动：集成多表级联保存逻辑。

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) AddAddress(ctx context.Context, userID uint, contact, phone, province, city, detail string, isDefault bool) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return xerrors.NotFound("user not found")
	}

	user.AddAddress(domain.Address{
		UserID:    userID,
		Contact:   contact,
		Phone:     phone,
		Province:  province,
		City:      city,
		Detail:    detail,
		IsDefault: isDefault,
	})

	return s.repo.Save(ctx, user)
}

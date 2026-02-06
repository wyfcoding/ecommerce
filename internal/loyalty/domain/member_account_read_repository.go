// 生成摘要：定义会员账户读模型仓储接口（Redis），用于高频查询。
package domain

import "context"

// MemberAccountReadRepository 定义会员账户读模型的高性能访问接口。
type MemberAccountReadRepository interface {
	Save(ctx context.Context, account *MemberAccount) error
	GetByUserID(ctx context.Context, userID uint64) (*MemberAccount, error)
	Delete(ctx context.Context, userID uint64) error
}

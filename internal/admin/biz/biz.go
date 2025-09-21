package biz

import "context"

// Transaction 定义了事务管理器接口。
type Transaction interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

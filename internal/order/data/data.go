package data

import (
	"context"
	"ecommerce/internal/order/biz"

	"gorm.io/gorm"
)

// Data 结构体持有数据库连接。
type Data struct {
	db *gorm.DB
}

// NewData 是 Data 结构体的构造函数。
func NewData(db *gorm.DB) *Data {
	return &Data{db: db}
}

// transaction 实现了 biz.Transaction 接口。
type transaction struct {
	db *gorm.DB
}

// NewTransaction 是 transaction 的构造函数。
func NewTransaction(data *Data) biz.Transaction {
	return &transaction{db: data.db}
}

type txKey struct{}

// GetDBFromContext 从 context 中获取 GORM DB 实例，如果存在事务，则返回事务 DB。
func GetDBFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return nil // 或者返回一个非事务的 DB 实例，取决于设计
}

// ExecTx 在一个数据库事务中执行函数。
func (t *transaction) ExecTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 将事务句柄 tx 放入 context 中，以便 repo 层可以获取并使用它。
		ctxWithTx := context.WithValue(ctx, txKey{}, tx)
		return fn(ctxWithTx)
	})
}

package data

import (
	"context"
	"ecommerce/internal/admin/biz"
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

// InTx 在一个数据库事务中执行函数。
func (t *transaction) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(fn)
}

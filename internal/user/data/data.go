package data

// package data 存放与数据持久化相关的代码。
import (
	"ecommerce/internal/user/data/model"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Data 结构体是数据层的核心，它作为一个容器，集中管理所有数据源的连接。
// 在当前的用户服务中，它只包含一个数据库连接 `db`。
// 在更复杂的服务中，它可能还会包含 Redis 客户端、MQ 生产者等。
type Data struct {
	db *gorm.DB
}

// NewData 是 Data 结构体的构造函数，负责初始化所有数据源的连接。
// 这种模式被称为“依赖注入容器”，在 main.go 中调用，然后将返回的 Data 实例注入到各个 Repo 中。
// 它返回三个值：
// 1. *Data: 成功初始化的 Data 实例。
// 2. func(): 一个清理函数，用于在服务优雅关闭时释放所有数据源连接。
// 3. error: 初始化过程中发生的任何错误。
func NewData(dsn string) (*Data, func(), error) {
	// 使用 GORM 和指定的 MySQL 驱动来打开一个数据库连接。
	// dsn (Data Source Name) 是数据库连接字符串，从配置文件中读取。
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		// 如果连接失败，直接返回错误。
		return nil, nil, err
	}

	// 定义一个闭包函数 `cleanup`，用于资源释放。
	// 当服务需要关闭时，这个函数会被调用，以确保数据库连接被正常关闭，防止连接泄露。
	cleanup := func() {
		// 获取底层的 *sql.DB 对象并调用 Close()。
		sqlDB, err := db.DB()
		if err != nil {
			zap.S().Errorf("failed to get database instance for cleanup: %v", err)
			return
		}
		if sqlDB != nil {
			zap.S().Info("closing database connection...")
			if err := sqlDB.Close(); err != nil {
				zap.S().Errorf("failed to close database connection: %v", err)
			}
		}
	}

	// GORM 的 AutoMigrate 功能可以自动根据定义的模型（如此处的 User 结构体）
	// 来创建或更新数据库表结构。这在开发阶段非常方便。
	// 注意：在生产环境中，数据库变更通常会由更严格的工具（如 Flyway, Liquibase）来管理。
	if err := db.AutoMigrate(&model.User{}, &model.Address{}); err != nil {
		// 如果自动迁移失败，也需要调用 cleanup 来关闭已建立的连接。
		cleanup()
		return nil, nil, err
	}

	// 所有初始化成功，返回 Data 实例和清理函数。
	return &Data{db: db}, cleanup, nil
}

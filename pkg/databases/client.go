package databases

import (
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/logging"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/opentelemetry/tracing"
)

// NewDB creates a new GORM DB instance.
func NewDB(cfg config.DatabaseConfig, logger *logging.Logger) (*gorm.DB, error) {
	// Use custom GormLogger
	gormLogger := logging.NewGormLogger(logger, cfg.SlowThreshold)

	gormConfig := &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // Use singular table names
		},
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	var dialector gorm.Dialector
	switch cfg.Driver {
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	// Add other drivers here as needed, e.g., postgres, sqlite
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, err
	}

	// Enable OpenTelemetry Tracing
	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("failed to use tracing plugin: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

package dbhelper

import (
	"context"
	"fmt"
	"time"

	logs "briefcash-jwt/internal/helper/loghelper"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormAdapter struct {
	DB *gorm.DB
}

type GormConfig struct {
	Address  string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewGormAdapter(gormCfg GormConfig) (*GormAdapter, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		gormCfg.Address, gormCfg.Port, gormCfg.User, gormCfg.Password, gormCfg.DBName, gormCfg.SSLMode,
	)

	// Init config
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: false,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		logs.Logger.WithError(err).Error("Failed to connect to database")
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Create generic function
	sqlDB, err := db.DB()
	if err != nil {
		logs.Logger.WithError(err).Error("Failed to get generic database")
		return nil, fmt.Errorf("failed to get generic database: %w", err)
	}

	// Ping database connection
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		logs.Logger.WithError(err).Error("Failed to ping database")
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &GormAdapter{DB: db}, nil
}

func (h *GormAdapter) AutoMigrate(models ...interface{}) error {
	return h.DB.AutoMigrate(models...)
}

func (h *GormAdapter) Close() error {
	sqlDB, err := h.DB.DB()
	if err == nil {
		return sqlDB.Close()
	}
	return fmt.Errorf("failed to close database: %w", err)
}

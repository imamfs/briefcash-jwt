package dbhelper

import (
	"fmt"
	"time"

	logs "briefcash-jwt/internal/helper/loghelper"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormHelper struct {
	DB *gorm.DB
}

type GormConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewGormHelper(gormCfg GormConfig) (*GormHelper, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		gormCfg.Host, gormCfg.Port, gormCfg.User, gormCfg.Password, gormCfg.DBName, gormCfg.SSLMode,
	)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: false,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		logs.Logger.Error("failed to connect database: %w", err)
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logs.Logger.Error("failed to get generic database: %w", err)
		return nil, fmt.Errorf("failed to get generic database: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &GormHelper{DB: db}, nil
}

func (h *GormHelper) AutoMigrate(models ...interface{}) error {
	return h.DB.AutoMigrate(models...)
}

func (h *GormHelper) Close() error {
	sqlDB, err := h.DB.DB()
	if err == nil {
		return sqlDB.Close()
	}
	return nil
}

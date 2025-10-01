package dbhelper

import (
	logs "briefcash-jwt/internal/helper/loghelper"
	"database/sql"
	"fmt"
	"time"
)

type SQLHelper struct {
	DB *sql.DB
}

type SQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewSQLHelper(config SQLConfig) (*SQLHelper, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		logs.Logger.Error("failed to connect database: %w", err)
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		logs.Logger.Error("failed to open connection with db: %w", err)
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &SQLHelper{DB: db}, nil
}

func (h *SQLHelper) Close() {
	h.DB.Close()
}

package dbhelper

import (
	logs "briefcash-jwt/internal/helper/loghelper"
	"database/sql"
	"fmt"
	"time"
)

type SQLAdapter struct {
	DB *sql.DB
}

type SQLConfig struct {
	Address  string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewSQLAdapter(config SQLConfig) (*SQLAdapter, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Address, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		logs.Logger.Error("Failed established database connection: %w", err)
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		logs.Logger.Error("Failed to ping database: %w", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &SQLAdapter{DB: db}, nil
}

func (h *SQLAdapter) Close() {
	h.DB.Close()
}

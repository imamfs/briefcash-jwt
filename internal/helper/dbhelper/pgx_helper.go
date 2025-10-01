package dbhelper

import (
	"context"
	"fmt"
	"time"

	logs "briefcash-jwt/internal/helper/loghelper"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxHelper struct {
	Pool *pgxpool.Pool
}

type PgxConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewPgxHelper(pgxCfg PgxConfig) (*PgxHelper, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		pgxCfg.User, pgxCfg.Password, pgxCfg.Host, pgxCfg.Port, pgxCfg.DBName, pgxCfg.SSLMode,
	)

	config, err := pgxpool.ParseConfig(dsn)

	if err != nil {
		logs.Logger.Error("failed to parse pgx config: %w", err)
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	config.MaxConns = 50
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		logs.Logger.Error("failed to connect pgxpool: %w", err)
		return nil, fmt.Errorf("failed to connect pgxpool: %w", err)
	}

	return &PgxHelper{Pool: pool}, nil
}

func (h *PgxHelper) Close() {
	h.Pool.Close()
}

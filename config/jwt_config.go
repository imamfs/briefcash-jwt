package config

import (
	"os"
	"time"

	env "github.com/joho/godotenv"
)

type Config struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	DBUrl           string
	RedisAddress    string
	RedisPassword   string
}

func LoadConfiguration() *Config {
	_ = env.Load()

	return &Config{
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		DBUrl:           os.Getenv("DB_URL"),
		RedisAddress:    os.Getenv("REDIS_ADDR"),
		RedisPassword:   os.Getenv("REDIS_PSSWRD"),
	}
}

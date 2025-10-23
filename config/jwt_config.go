package config

import (
	"fmt"
	"os"

	logs "briefcash-jwt/internal/helper/loghelper"

	env "github.com/joho/godotenv"
)

type Config struct {
	JWTSecret     string
	DBUrl         string
	DbHost        string
	DbUsername    string
	DbPassword    string
	DbPort        string
	DbName        string
	RedisAddress  string
	RedisPort     string
	RedisPassword string
	AppPort       string
}

func LoadConfiguration() (*Config, error) {
	if err := env.Load(); err != nil {
		logs.Logger.Error("No .env file found, using system environment variables")
	}

	cfg := &Config{
		JWTSecret:     os.Getenv("JWT_SECRET"),
		DBUrl:         os.Getenv("DB_URL"),
		DbHost:        os.Getenv("DB_HOST"),
		DbUsername:    os.Getenv("DB_USERNAME"),
		DbPassword:    os.Getenv("DB_PASSWORD"),
		DbPort:        os.Getenv("DB_PORT"),
		DbName:        os.Getenv("DB_NAME"),
		RedisAddress:  os.Getenv("REDIS_ADDRESS"),
		RedisPort:     os.Getenv("REDIS_PORT"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		AppPort: func() string {
			if value := os.Getenv("APP_PORT"); value != "" {
				return value
			}
			return ":8080"
		}(),
	}

	if cfg.JWTSecret == "" {
		logs.Logger.Error("JWT_SECRET is not set in environment")
		return nil, fmt.Errorf("JWT_SECRET is not set in environment")
	}

	if cfg.DbHost == "" {
		logs.Logger.Error("DB_HOST is not set in environment")
		return nil, fmt.Errorf("DB_HOST is not set in environment")
	}

	return cfg, nil
}

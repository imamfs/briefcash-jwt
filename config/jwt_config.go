package config

import (
	"fmt"
	"os"

	logs "briefcash-jwt/internal/helper/loghelper"

	env "github.com/joho/godotenv"
)

type Config struct {
	JWTSecret     string
	DbAddress     string
	DbUsername    string
	DbPassword    string
	DbPort        string
	DbName        string
	RedisAddress  string
	RedisPort     string
	RedisPassword string
	AppPort       string
}

func LoadConfig() (*Config, error) {
	// Load credentials form .env file, if file not found, load OS environment variables
	if err := env.Load(); err != nil {
		logs.Logger.Error("No .env file found, using system environment variables")
	}

	// Set credentials to struct
	cfg := &Config{
		JWTSecret:     os.Getenv("JWT_SECRET"),
		DbAddress:     os.Getenv("DB_ADDRESS"),
		DbPort:        os.Getenv("DB_PORT"),
		DbUsername:    os.Getenv("DB_USERNAME"),
		DbPassword:    os.Getenv("DB_PASSWORD"),
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

	// Validate jwt secret and db host
	if cfg.JWTSecret == "" {
		logs.Logger.Error("JWT_SECRET is not set in environment")
		return nil, fmt.Errorf("JWT_SECRET is not set in environment")
	}

	if cfg.DbAddress == "" {
		logs.Logger.Error("DB_HOST is not set in environment")
		return nil, fmt.Errorf("DB_HOST is not set in environment")
	}

	return cfg, nil
}

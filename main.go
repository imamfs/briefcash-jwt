package main

import (
	"briefcash-jwt/config"
	controller "briefcash-jwt/internal/controller"
	gormHelper "briefcash-jwt/internal/helper/dbhelper"
	logHelper "briefcash-jwt/internal/helper/loghelper"
	redisHelper "briefcash-jwt/internal/helper/redishelper"
	middleware "briefcash-jwt/internal/middleware"
	repo "briefcash-jwt/internal/repository"
	service "briefcash-jwt/internal/service"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func main() {
	logHelper.InitLogger("./briefcash-jwt/resource/app.log", logrus.InfoLevel)

	cfg, err := config.LoadConfiguration()

	if err != nil {
		logHelper.Logger.WithError(err).Fatal("Failed to load configuration")
	}

	grmConfig := gormHelper.GormConfig{
		Host:     cfg.DbHost,
		Port:     cfg.DbPort,
		User:     cfg.DbUsername,
		Password: cfg.DbPassword,
		DBName:   cfg.DbName,
		SSLMode:  "false",
	}

	redisClient, err := redisHelper.NewRedisHelper(cfg)

	if err != nil {
		logHelper.Logger.WithError(err).Fatal("failed to connect to redis")
	}

	defer redisClient.Close()

	dbHelper, err := gormHelper.NewGormHelper(grmConfig)

	if err != nil {
		logHelper.Logger.WithError(err).Fatal("failed to connect to database")
	}

	defer dbHelper.Close()

	jwtRepo := repo.NewJwtRepository(dbHelper.DB)
	redisRepo := repo.NewRedisRepository(redisClient.Client)

	jwtService := service.NewTokenService(jwtRepo, redisRepo, dbHelper.DB, cfg.JWTSecret)

	jwtController := controller.NewTokenController(jwtService)

	mux := http.NewServeMux()

	mux.HandleFunc("/token/generate", jwtController.GenerateToken)
	mux.HandleFunc("/token/refresh", jwtController.RefreshToken)

	mux.Handle("/token/validate", middleware.AuthMiddleware(http.HandlerFunc(jwtController.ValidateToken)))
	mux.Handle("/token/logout", middleware.AuthMiddleware(http.HandlerFunc(jwtController.Logout)))

	addr := fmt.Sprintf(":%s", cfg.AppPort)

	if err := http.ListenAndServe(addr, mux); err != nil {
		logHelper.Logger.WithError(err).Fatal("server failed to start")
	}

}

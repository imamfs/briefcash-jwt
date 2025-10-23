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
	"context"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func main() {
	logHelper.InitLogger("./resource/app.log", logrus.InfoLevel)

	cfg, err := config.LoadConfiguration()
	ctx := context.Background()

	if err != nil {
		logHelper.Logger.WithError(err).Fatal("Failed to load configuration")
	}

	grmConfig := gormHelper.GormConfig{
		Host:     cfg.DbHost,
		Port:     cfg.DbPort,
		User:     cfg.DbUsername,
		Password: cfg.DbPassword,
		DBName:   cfg.DbName,
		SSLMode:  "disable",
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
	merchantRepo := repo.NewMerchantRepository(dbHelper.DB)
	redisRepo := repo.NewRedisRepository(redisClient.Client)
	merchantRedisRepo := repo.NewMerchantRedisRepository(redisClient.Client)

	jwtService := service.NewTokenService(jwtRepo, redisRepo, dbHelper.DB, cfg.JWTSecret)
	merchantService := service.NewMerchantService(merchantRepo, merchantRedisRepo)

	if err := merchantService.LoadActiveMerchantCodeToRedis(ctx); err != nil {
		logHelper.Logger.WithError(err).Fatal("Failed to load merchant code to redis")
	}

	jwtController := controller.NewTokenController(jwtService)
	merchantController := controller.NewMerchantController(merchantService)

	mware := middleware.NewMiddleware(merchantService)

	mux := http.NewServeMux()

	mux.HandleFunc("/token/generate", jwtController.GenerateToken)
	mux.HandleFunc("/token/refresh", jwtController.RefreshToken)

	mux.Handle("/token/validate", mware.AuthMiddleware(http.HandlerFunc(jwtController.ValidateToken)))
	mux.Handle("/token/logout", mware.AuthMiddleware(http.HandlerFunc(jwtController.Logout)))

	mux.HandleFunc("/data-merchant/sync", merchantController.SyncMerchantCode)
	mux.HandleFunc("/data-merchant/add", merchantController.AddMerchantCode)
	mux.HandleFunc("/data-merchant/remove", merchantController.RemoveMerchantCode)

	addr := fmt.Sprintf(":%s", cfg.AppPort)

	if err := http.ListenAndServe(addr, mux); err != nil {
		logHelper.Logger.WithError(err).Fatal("server failed to start")
	}

}

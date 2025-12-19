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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logHelper.InitLogger("./resource/app.log", logrus.InfoLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create goroutine signal for gracefully shutdown
	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
		sign := <-sc
		logHelper.Logger.WithField("signal", sign.String()).Info("Received shutdown signal...")
		cancel()
	}()

	// Load credentials configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logHelper.Logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Build redis connection
	redisClient, err := redisHelper.NewRedisAdapter(cfg)
	if err != nil {
		logHelper.Logger.WithError(err).Fatal("failed to connect to redis")
	}
	defer redisClient.Close()

	// Build and init database connection
	gormConfig := gormHelper.GormConfig{
		Address:  cfg.DbAddress,
		Port:     cfg.DbPort,
		User:     cfg.DbUsername,
		Password: cfg.DbPassword,
		DBName:   cfg.DbName,
		SSLMode:  "disable",
	}

	dbHelper, err := gormHelper.NewGormAdapter(gormConfig)
	if err != nil {
		logHelper.Logger.WithError(err).Fatal("failed to connect to database")
	}
	defer dbHelper.Close()

	// Create repository instance
	jwtRepo := repo.NewJwtRepository(dbHelper.DB)
	merchantRepo := repo.NewMerchantRepository(dbHelper.DB)
	redisRepo := repo.NewRedisRepository(redisClient.Client)
	merchantRedisRepo := repo.NewMerchantRedisRepository(redisClient.Client)

	// Create service instance
	jwtService := service.NewTokenService(jwtRepo, redisRepo, dbHelper.DB, cfg.JWTSecret)
	merchantService := service.NewMerchantService(merchantRepo, merchantRedisRepo)

	// Load list of merchant code from database to redis
	if err := merchantService.CachingCode(ctx); err != nil {
		logHelper.Logger.WithError(err).Fatal("Failed to load merchant code to redis")
	}

	// Create controller instance
	jwtController := controller.NewTokenController(jwtService)
	merchantController := controller.NewMerchantController(merchantService)

	// Create middleware instance
	mw := middleware.NewMiddleware(merchantService)

	// Init http connection
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(RequestLoggerMiddleware())

	api := router.Group("/api/v1")
	{
		token := api.Group("/token")
		{
			token.POST("/generate", jwtController.GenerateToken)
			token.POST("/refresh", jwtController.RefreshToken)
			token.POST("/validate", mw.AuthMiddleware(), jwtController.ValidateToken)
			token.POST("/logout", mw.AuthMiddleware(), jwtController.Logout)
		}

		merchant := api.Group("/merchant")
		{
			merchant.POST("/sync", gin.WrapF(merchantController.SyncMerchantCode))
			merchant.POST("/add", gin.WrapF(merchantController.AddMerchantCode))
			merchant.POST("/remove", gin.WrapF(merchantController.RemoveMerchantCode))
		}
	}

	server := &http.Server{
		Addr:    cfg.AppPort,
		Handler: router,
	}

	// Create goroutine for openning http connection
	go func() {
		logHelper.Logger.WithField("port", cfg.AppPort).Info("JWT Service is running...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logHelper.Logger.WithError(err).Fatal("Failed to start JWT Service")
		}
	}()

	// If apps receives shutdown signal, then shutdown gracefully
	<-ctx.Done()

	logHelper.Logger.Info("Shutting Down JWT Service gracefully...")

	// Set timeout for shutdown
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logHelper.Logger.WithError(err).Error("Forced shutdown due to timeout")
	} else {
		logHelper.Logger.Info("JWT Service shutdown completed")
	}

}

// Middleware function for logging http activity
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		logHelper.Logger.WithFields(logrus.Fields{
			"method":   c.Request.Method,
			"path":     c.FullPath(),
			"status":   status,
			"duration": duration.String(),
			"clientIp": c.ClientIP(),
		}).Info("Handled request")
	}
}

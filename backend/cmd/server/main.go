package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/edberthansel/BtechDevCases/backend/internal/config"
	"github.com/edberthansel/BtechDevCases/backend/internal/database"
	"github.com/edberthansel/BtechDevCases/backend/internal/handler"
	"github.com/edberthansel/BtechDevCases/backend/internal/repository"
	"github.com/edberthansel/BtechDevCases/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/edberthansel/BtechDevCases/backend/internal/token"
	"github.com/edberthansel/BtechDevCases/backend/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool, "migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	userRepo := repository.NewUserRepository(pool)
	tokenManager := token.NewManager(cfg.JWTSecret, time.Duration(cfg.JWTExpiryMinutes)*time.Minute)
	authService := service.NewAuthService(userRepo, tokenManager, cfg.InitialBalance)
	authHandler := handler.NewAuthHandler(authService)
	walletRepo := repository.NewWalletRepository(pool)
	walletService := service.NewWalletService(walletRepo)
	walletHandler := handler.NewWalletHandler(walletService)

	r := gin.Default()
	r.Use(middleware.CORS(cfg.CORSOrigin))

	r.GET("/health", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	protected := r.Group("/")
	protected.Use(middleware.Auth(tokenManager))
	{
		protected.GET("/me", handler.Me)
		protected.GET("/wallet", walletHandler.GetBalance)
	}

	log.Printf("server listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
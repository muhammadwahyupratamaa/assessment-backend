package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/config"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/handler"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/lib"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/middleware"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/repo"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	appConfig, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	if appConfig.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := lib.RunMigrations(appConfig.DatabaseURL); err != nil {
		log.Fatal("failed to run migrations: ", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := lib.ConnectDB(ctx, appConfig.DatabaseURL)
	if err != nil {
		log.Fatal("database connection failed: ", err)
	}
	defer db.Close()

	if err := lib.SeedAdmin(
		ctx,
		db,
		appConfig.AdminEmail,
		appConfig.AdminPassword,
	); err != nil {
		log.Fatal("failed to seed admin: ", err)
	}

	// Repository
	userRepo := repo.NewUserRepo(db)

	// Service
	authService := services.NewAuthService(userRepo)

	// Handler
	authHandler := handler.NewAuthHandler(
		authService,
		appConfig.JWTSecret,
		appConfig.JWTExpiresHours,
	)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"message": "database unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"status": "ok",
			},
		})
	})

	api := router.Group("/api/v1")

	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(appConfig.JWTSecret))
	{
		protected.GET("/me", authHandler.Me)
	}

	server := &http.Server{
		Addr:              ":" + appConfig.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("server running", "port", appConfig.AppPort)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server failed: ", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("server shutdown failed: ", err)
	}
}

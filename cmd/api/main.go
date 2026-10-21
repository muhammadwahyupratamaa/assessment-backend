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
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/repo"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/routes"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/services"

	_ "github.com/muhammadwahyupratamaa/assessment-backend/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

// @title						Assessment Backend API
// @version					1.0
// @description				REST API for a mini e-commerce application.
// @BasePath					/api/v1
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and a JWT token.
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
	categoryRepo := repo.NewCategoryRepo(db)
	productRepo := repo.NewProductRepo(db)
	orderRepo := repo.NewOrderRepo(db)

	// Service
	authService := services.NewAuthService(userRepo)
	categoryService := services.NewCategoryService(categoryRepo)
	productService := services.NewProductService(productRepo, categoryRepo)
	orderService := services.NewOrderService(orderRepo)

	// Handler
	authHandler := handler.NewAuthHandler(
		authService,
		appConfig.JWTSecret,
		appConfig.JWTExpiresHours,
	)
	healthHandler := handler.NewHealthHandler(db)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	productHandler := handler.NewProductHandler(productService)
	orderHandler := handler.NewOrderHandler(orderService)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	routes.Register(router, routes.Dependencies{
		HealthHandler:   healthHandler,
		AuthHandler:     authHandler,
		CategoryHandler: categoryHandler,
		ProductHandler:  productHandler,
		OrderHandler:    orderHandler,
		JWTSecret:       appConfig.JWTSecret,
	})

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

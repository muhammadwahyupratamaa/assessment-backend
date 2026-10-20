package routes

import (
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/handler"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	HealthHandler   *handler.HealthHandler
	AuthHandler     *handler.AuthHandler
	CategoryHandler *handler.CategoryHandler
	ProductHandler  *handler.ProductHandler
	JWTSecret       string
}

func Register(router *gin.Engine, dependencies Dependencies) {
	router.GET("/health", dependencies.HealthHandler.Check)

	api := router.Group("/api/v1")

	registerAuthRoutes(api, dependencies)
	registerProductRoutes(api, dependencies)
}

func registerAuthRoutes(
	api *gin.RouterGroup,
	dependencies Dependencies,
) {
	auth := api.Group("/auth")
	{
		auth.POST("/register", dependencies.AuthHandler.Register)
		auth.POST("/login", dependencies.AuthHandler.Login)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(dependencies.JWTSecret))
	{
		protected.GET("/me", dependencies.AuthHandler.Me)
	}
}

package routes

import (
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerProductRoutes(
	api *gin.RouterGroup,
	dependencies Dependencies,
) {
	api.GET("/categories", dependencies.CategoryHandler.GetAll)

	api.GET("/products", dependencies.ProductHandler.List)
	api.GET("/products/:id", dependencies.ProductHandler.GetByID)

	admin := api.Group("")
	admin.Use(
		middleware.AuthMiddleware(dependencies.JWTSecret),
		middleware.RequireAdmin(),
	)
	{
		admin.POST("/categories", dependencies.CategoryHandler.Create)

		admin.POST("/products", dependencies.ProductHandler.Create)
		admin.PUT("/products/:id", dependencies.ProductHandler.Update)
		admin.DELETE("/products/:id", dependencies.ProductHandler.Delete)
	}
}

package routes

import (
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerOrderRoutes(
	api *gin.RouterGroup,
	dependencies Dependencies,
) {
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(dependencies.JWTSecret))
	{
		protected.POST("/orders", dependencies.OrderHandler.Checkout)
		protected.GET("/orders", dependencies.OrderHandler.ListMyOrders)
		protected.GET("/orders/:id", dependencies.OrderHandler.GetMyOrder)
	}

	admin := api.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(dependencies.JWTSecret),
		middleware.RequireAdmin(),
	)
	{
		admin.GET("/orders", dependencies.OrderHandler.ListAllOrders)

		admin.PATCH(
			"/orders/:id/status",
			dependencies.OrderHandler.UpdateStatus,
		)
	}
}

package middleware

import (
	"net/http"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"

	"github.com/gin-gonic/gin"
)

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := GetAuthUser(c)

		if !ok || user.Role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

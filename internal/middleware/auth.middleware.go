package middleware

import (
	"net/http"
	"strings"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/lib"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"

	"github.com/gin-gonic/gin"
)

const AuthUserContextKey = "auth_user"

type AuthUser struct {
	ID   string
	Role models.Role
}

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		parts := strings.Fields(authHeader)

		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "authorization header must use Bearer token",
			})
			c.Abort()
			return
		}

		claims, err := lib.ParseToken(parts[1], jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Set(AuthUserContextKey, AuthUser{
			ID:   claims.Subject,
			Role: claims.Role,
		})

		c.Next()
	}
}

func GetAuthUser(c *gin.Context) (AuthUser, bool) {
	value, exists := c.Get(AuthUserContextKey)
	if !exists {
		return AuthUser{}, false
	}

	user, ok := value.(AuthUser)

	return user, ok
}

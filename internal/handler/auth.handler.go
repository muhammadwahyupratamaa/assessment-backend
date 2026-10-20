package handler

import (
	"errors"
	"net/http"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/lib"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/middleware"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService     services.AuthService
	jwtSecret       string
	jwtExpiresHours int
}

func NewAuthHandler(
	authService services.AuthService,
	jwtSecret string,
	jwtExpiresHours int,
) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		jwtSecret:       jwtSecret,
		jwtExpiresHours: jwtExpiresHours,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var request models.RegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), request)
	if err != nil {
		if errors.Is(err, services.ErrWeakPassword) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "password must contain uppercase, lowercase, number, and special character",
			})
			return
		}

		if errors.Is(err, services.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "email already registered",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to register user",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    models.ToUserResponse(user),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request models.LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	user, err := h.authService.Login(c.Request.Context(), request)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid email or password",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to login",
		})
		return
	}

	token, err := lib.GenerateToken(
		user.ID,
		user.Role,
		h.jwtSecret,
		h.jwtExpiresHours,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token": token,
			"user":  models.ToUserResponse(user),
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	user, err := h.authService.GetProfile(
		c.Request.Context(),
		authUser.ID,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "user not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    models.ToUserResponse(user),
	})
}

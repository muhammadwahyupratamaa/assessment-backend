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

// Register godoc
//
//	@Summary		Register a customer account
//	@Description	Create a new customer account. Password must contain uppercase, lowercase, number, and special character.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.RegisterRequest	true	"Registration payload"
//	@Success		201		{object}	models.UserResponse		"User data is returned inside the data field"
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		409		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/auth/register [post]
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

// Login godoc
//
//	@Summary		Login
//	@Description	Authenticate with email and password, then receive a JWT access token.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.LoginRequest	true	"Login payload"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		401		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/auth/login [post]
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

// Me godoc
//
//	@Summary	Get current user profile
//	@Tags		Authentication
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	models.UserResponse	"User data is returned inside the data field"
//	@Failure	401	{object}	map[string]interface{}
//	@Router		/me [get]
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

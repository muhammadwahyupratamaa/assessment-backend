package handler

import (
	"errors"
	"net/http"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService services.CategoryService
}

func NewCategoryHandler(
	categoryService services.CategoryService,
) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) GetAll(c *gin.Context) {
	categories, err := h.categoryService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to get categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
	})
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var request models.CreateCategoryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	category, err := h.categoryService.Create(c.Request.Context(), request)
	if err != nil {
		if errors.Is(err, services.ErrCategoryAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "category already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to create category",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    category,
	})
}

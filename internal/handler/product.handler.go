package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService services.ProductService
}

func NewProductHandler(
	productService services.ProductService,
) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) List(c *gin.Context) {
	params, err := buildProductListParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	products, meta, err := h.productService.List(
		c.Request.Context(),
		params,
	)
	if err != nil {
		if errors.Is(err, services.ErrInvalidProductFilter) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "invalid product filter",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to get products",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    products,
		"meta":    meta,
	})
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	productID := c.Param("id")

	if !isUUID(productID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid product id",
		})
		return
	}

	product, err := h.productService.GetByID(
		c.Request.Context(),
		productID,
	)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to get product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    product,
	})
}

func (h *ProductHandler) Create(c *gin.Context) {
	var request models.CreateProductRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	product, err := h.productService.Create(c.Request.Context(), request)
	if err != nil {
		if errors.Is(err, services.ErrCategoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to create product",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    product,
	})
}

func (h *ProductHandler) Update(c *gin.Context) {
	productID := c.Param("id")

	if !isUUID(productID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid product id",
		})
		return
	}

	var request models.UpdateProductRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	product, err := h.productService.Update(
		c.Request.Context(),
		productID,
		request,
	)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "product not found",
			})
		case errors.Is(err, services.ErrCategoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "category not found",
			})
		case errors.Is(err, services.ErrNoProductChanges):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "no product changes provided",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to update product",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    product,
	})
}

func (h *ProductHandler) Delete(c *gin.Context) {
	productID := c.Param("id")

	if !isUUID(productID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid product id",
		})
		return
	}

	err := h.productService.Delete(c.Request.Context(), productID)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to delete product",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func buildProductListParams(
	c *gin.Context,
) (models.ProductListParams, error) {
	params := models.ProductListParams{
		Page:  1,
		Limit: 10,
		Sort:  c.DefaultQuery("sort", "created_at"),
		Order: c.DefaultQuery("order", "desc"),
	}

	if value := c.Query("page"); value != "" {
		page, err := strconv.Atoi(value)
		if err != nil || page < 1 {
			return params, errors.New("page must be a positive number")
		}

		params.Page = page
	}

	if value := c.Query("limit"); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit < 1 || limit > 100 {
			return params, errors.New("limit must be between 1 and 100")
		}

		params.Limit = limit
	}

	if value := c.Query("category_id"); value != "" {
		if !isUUID(value) {
			return params, errors.New("invalid category_id")
		}

		params.CategoryID = value
	}

	if value := c.Query("min_price"); value != "" {
		price, err := strconv.ParseInt(value, 10, 64)
		if err != nil || price < 0 {
			return params, errors.New("min_price must be a non-negative number")
		}

		params.MinPrice = &price
	}

	if value := c.Query("max_price"); value != "" {
		price, err := strconv.ParseInt(value, 10, 64)
		if err != nil || price < 0 {
			return params, errors.New("max_price must be a non-negative number")
		}

		params.MaxPrice = &price
	}

	if value := c.Query("in_stock"); value != "" {
		inStock, err := strconv.ParseBool(value)
		if err != nil {
			return params, errors.New("in_stock must be true or false")
		}

		params.InStock = &inStock
	}

	return params, nil
}

func isUUID(value string) bool {
	parts := strings.Split(value, "-")
	expectedLengths := []int{8, 4, 4, 4, 12}

	if len(parts) != len(expectedLengths) {
		return false
	}

	for index, part := range parts {
		if len(part) != expectedLengths[index] {
			return false
		}

		for _, character := range part {
			isNumber := character >= '0' && character <= '9'
			isLowerHex := character >= 'a' && character <= 'f'
			isUpperHex := character >= 'A' && character <= 'F'

			if !isNumber && !isLowerHex && !isUpperHex {
				return false
			}
		}
	}

	return true
}

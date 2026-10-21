package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/middleware"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService services.OrderService
}

func NewOrderHandler(orderService services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// Checkout godoc
//
//	@Summary		Checkout an order
//	@Description	Create an order transactionally. Stock is validated and deducted atomically.
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Idempotency-Key	header		string					false	"Optional idempotency key to safely retry a request"
//	@Param			request			body		models.CheckoutRequest	true	"Checkout payload"
//	@Success		201				{object}	models.Order			"Order data is returned inside the data field"
//	@Failure		400				{object}	map[string]interface{}
//	@Failure		401				{object}	map[string]interface{}
//	@Failure		404				{object}	map[string]interface{}
//	@Failure		409				{object}	map[string]interface{}
//	@Failure		500				{object}	map[string]interface{}
//	@Router			/orders [post]
func (h *OrderHandler) Checkout(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	var request models.CheckoutRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")
	if len(idempotencyKey) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "idempotency key must not exceed 100 characters",
		})
		return
	}

	order, err := h.orderService.Checkout(
		c.Request.Context(),
		authUser.ID,
		idempotencyKey,
		request,
	)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrOrderProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "product not found",
			})
		case errors.Is(err, services.ErrOrderInsufficientStock):
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "insufficient product stock",
			})
		case errors.Is(err, services.ErrInvalidCheckoutItem):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "invalid checkout item",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to create order",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    order,
	})
}

// ListMyOrders godoc
//
//	@Summary	Get current user's orders
//	@Tags		Orders
//	@Produce	json
//	@Security	BearerAuth
//	@Param		page	query		int				false	"Page number"					default(1)
//	@Param		limit	query		int				false	"Items per page (maximum 100)"	default(10)
//	@Success	200		{array}		models.Order	"Order data is returned inside the data field"
//	@Failure	400		{object}	map[string]interface{}
//	@Failure	401		{object}	map[string]interface{}
//	@Failure	500		{object}	map[string]interface{}
//	@Router		/orders [get]
func (h *OrderHandler) ListMyOrders(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	params, err := buildOrderListParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	orders, meta, err := h.orderService.ListMyOrders(
		c.Request.Context(),
		authUser.ID,
		params,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to get orders",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
		"meta":    meta,
	})
}

// GetMyOrder godoc
//
//	@Summary	Get current user's order detail
//	@Tags		Orders
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string			true	"Order UUID"
//	@Success	200	{object}	models.Order	"Order data is returned inside the data field"
//	@Failure	400	{object}	map[string]interface{}
//	@Failure	401	{object}	map[string]interface{}
//	@Failure	404	{object}	map[string]interface{}
//	@Failure	500	{object}	map[string]interface{}
//	@Router		/orders/{id} [get]
func (h *OrderHandler) GetMyOrder(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	orderID := c.Param("id")
	if !isUUID(orderID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid order id",
		})
		return
	}

	order, err := h.orderService.GetMyOrder(
		c.Request.Context(),
		authUser.ID,
		orderID,
	)
	if err != nil {
		if errors.Is(err, services.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "order not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to get order",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    order,
	})
}

// ListAllOrders godoc
//
//	@Summary		Get all orders
//	@Description	Admin-only endpoint to view all customer orders.
//	@Tags			Admin Orders
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int				false	"Page number"					default(1)
//	@Param			limit	query		int				false	"Items per page (maximum 100)"	default(10)
//	@Success		200		{array}		models.Order	"Order data is returned inside the data field"
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		401		{object}	map[string]interface{}
//	@Failure		403		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/admin/orders [get]
func (h *OrderHandler) ListAllOrders(c *gin.Context) {
	params, err := buildOrderListParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	orders, meta, err := h.orderService.ListAllOrders(
		c.Request.Context(),
		params,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to get orders",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
		"meta":    meta,
	})
}

func buildOrderListParams(
	c *gin.Context,
) (models.OrderListParams, error) {
	params := models.OrderListParams{
		Page:  1,
		Limit: 10,
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

	return params, nil
}

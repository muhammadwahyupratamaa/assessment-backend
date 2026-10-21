package models

import "time"

type OrderStatus string

const (
	OrderStatusPending OrderStatus = "PENDING"
)

type Order struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	Status      OrderStatus `json:"status"`
	TotalAmount int64       `json:"total_amount"`
	Items       []OrderItem `json:"items,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

type OrderItem struct {
	ID          string `json:"id"`
	OrderID     string `json:"order_id"`
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Price       int64  `json:"price"`
	Quantity    int    `json:"quantity"`
	Subtotal    int64  `json:"subtotal"`
}

type CheckoutRequest struct {
	Items []CheckoutItemRequest `json:"items" binding:"required,min=1,dive"`
}

type CheckoutItemRequest struct {
	ProductID string `json:"product_id" binding:"required,uuid"`
	Quantity  int    `json:"quantity" binding:"required,gte=1,lte=100"`
}

type OrderListParams struct {
	Page  int
	Limit int
}

type OrderPaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

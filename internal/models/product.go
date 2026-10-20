package models

import "time"

type Product struct {
	ID           string     `json:"id"`
	CategoryID   *string    `json:"category_id,omitempty"`
	CategoryName string     `json:"category_name,omitempty"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Price        int64      `json:"price"`
	Stock        int        `json:"stock"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

type CreateProductRequest struct {
	CategoryID  *string `json:"category_id" binding:"omitempty,uuid"`
	Name        string  `json:"name" binding:"required,min=2,max=200"`
	Description string  `json:"description" binding:"max=5000"`
	Price       *int64  `json:"price" binding:"required,gte=0"`
	Stock       *int    `json:"stock" binding:"required,gte=0"`
}

type UpdateProductRequest struct {
	CategoryID  *string `json:"category_id" binding:"omitempty,uuid"`
	Name        *string `json:"name" binding:"omitempty,min=2,max=200"`
	Description *string `json:"description" binding:"omitempty,max=5000"`
	Price       *int64  `json:"price" binding:"omitempty,gte=0"`
	Stock       *int    `json:"stock" binding:"omitempty,gte=0"`
}

type ProductListParams struct {
	Page       int
	Limit      int
	CategoryID string
	MinPrice   *int64
	MaxPrice   *int64
	InStock    *bool
	Sort       string
	Order      string
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

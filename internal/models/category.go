package models

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

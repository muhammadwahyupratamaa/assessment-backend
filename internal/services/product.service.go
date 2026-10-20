package services

import (
	"context"
	"errors"
	"strings"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/repo"

	"github.com/jackc/pgx/v5"
)

var (
	ErrProductNotFound      = errors.New("product not found")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrInvalidProductFilter = errors.New("invalid product filter")
	ErrNoProductChanges     = errors.New("no product changes provided")
)

type ProductService interface {
	Create(
		ctx context.Context,
		request models.CreateProductRequest,
	) (models.Product, error)
	GetByID(ctx context.Context, productID string) (models.Product, error)
	List(
		ctx context.Context,
		params models.ProductListParams,
	) ([]models.Product, models.PaginationMeta, error)
	Update(
		ctx context.Context,
		productID string,
		request models.UpdateProductRequest,
	) (models.Product, error)
	Delete(ctx context.Context, productID string) error
}

type productService struct {
	productRepo  repo.ProductRepo
	categoryRepo repo.CategoryRepo
}

func NewProductService(
	productRepo repo.ProductRepo,
	categoryRepo repo.CategoryRepo,
) ProductService {
	return &productService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *productService) Create(
	ctx context.Context,
	request models.CreateProductRequest,
) (models.Product, error) {
	if request.CategoryID != nil {
		if err := s.ensureCategoryExists(ctx, *request.CategoryID); err != nil {
			return models.Product{}, err
		}
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)

	return s.productRepo.Create(ctx, request)
}

func (s *productService) GetByID(
	ctx context.Context,
	productID string,
) (models.Product, error) {
	product, err := s.productRepo.GetByID(ctx, productID)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Product{}, ErrProductNotFound
	}

	return product, err
}

func (s *productService) List(
	ctx context.Context,
	params models.ProductListParams,
) ([]models.Product, models.PaginationMeta, error) {
	if params.Page < 1 {
		params.Page = 1
	}

	if params.Limit < 1 {
		params.Limit = 10
	}

	if params.Limit > 100 {
		params.Limit = 100
	}

	if params.Sort == "" {
		params.Sort = "created_at"
	}

	if params.Order == "" {
		params.Order = "desc"
	}

	params.Sort = strings.ToLower(params.Sort)
	params.Order = strings.ToLower(params.Order)

	if params.Sort != "price" && params.Sort != "created_at" {
		return nil, models.PaginationMeta{}, ErrInvalidProductFilter
	}

	if params.Order != "asc" && params.Order != "desc" {
		return nil, models.PaginationMeta{}, ErrInvalidProductFilter
	}

	if params.MinPrice != nil &&
		params.MaxPrice != nil &&
		*params.MinPrice > *params.MaxPrice {
		return nil, models.PaginationMeta{}, ErrInvalidProductFilter
	}

	products, total, err := s.productRepo.List(ctx, params)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	totalPages := int((total + int64(params.Limit) - 1) / int64(params.Limit))

	return products, models.PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *productService) Update(
	ctx context.Context,
	productID string,
	request models.UpdateProductRequest,
) (models.Product, error) {
	if request.CategoryID == nil &&
		request.Name == nil &&
		request.Description == nil &&
		request.Price == nil &&
		request.Stock == nil {
		return models.Product{}, ErrNoProductChanges
	}

	if request.CategoryID != nil {
		if err := s.ensureCategoryExists(ctx, *request.CategoryID); err != nil {
			return models.Product{}, err
		}
	}

	if request.Name != nil {
		trimmedName := strings.TrimSpace(*request.Name)
		request.Name = &trimmedName
	}

	if request.Description != nil {
		trimmedDescription := strings.TrimSpace(*request.Description)
		request.Description = &trimmedDescription
	}

	product, err := s.productRepo.Update(ctx, productID, request)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Product{}, ErrProductNotFound
	}

	return product, err
}

func (s *productService) Delete(
	ctx context.Context,
	productID string,
) error {
	deleted, err := s.productRepo.SoftDelete(ctx, productID)
	if err != nil {
		return err
	}

	if !deleted {
		return ErrProductNotFound
	}

	return nil
}

func (s *productService) ensureCategoryExists(
	ctx context.Context,
	categoryID string,
) error {
	exists, err := s.categoryRepo.ExistsByID(ctx, categoryID)
	if err != nil {
		return err
	}

	if !exists {
		return ErrCategoryNotFound
	}

	return nil
}

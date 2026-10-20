package services

import (
	"context"
	"errors"
	"strings"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"
	"github.com/muhammadwahyupratamaa/assessment-backend/internal/repo"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrCategoryAlreadyExists = errors.New("category already exists")

type CategoryService interface {
	Create(
		ctx context.Context,
		request models.CreateCategoryRequest,
	) (models.Category, error)

	GetAll(ctx context.Context) ([]models.Category, error)
}

type categoryService struct {
	categoryRepo repo.CategoryRepo
}

func NewCategoryService(
	categoryRepo repo.CategoryRepo,
) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *categoryService) Create(
	ctx context.Context,
	request models.CreateCategoryRequest,
) (models.Category, error) {
	category, err := s.categoryRepo.Create(
		ctx,
		strings.TrimSpace(request.Name),
	)
	if err != nil {
		var pgError *pgconn.PgError

		if errors.As(err, &pgError) && pgError.Code == "23505" {
			return models.Category{}, ErrCategoryAlreadyExists
		}

		return models.Category{}, err
	}

	return category, nil
}

func (s *categoryService) GetAll(
	ctx context.Context,
) ([]models.Category, error) {
	return s.categoryRepo.GetAll(ctx)
}

package repo

import (
	"context"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepo interface {
	Create(
		ctx context.Context,
		name string,
	) (models.Category, error)

	GetAll(ctx context.Context) ([]models.Category, error)

	ExistsByID(
		ctx context.Context,
		id string,
	) (bool, error)
}

type categoryRepo struct {
	db *pgxpool.Pool
}

func NewCategoryRepo(db *pgxpool.Pool) CategoryRepo {
	return &categoryRepo{
		db: db,
	}
}

func (r *categoryRepo) Create(
	ctx context.Context,
	name string,
) (models.Category, error) {
	query := `
		INSERT INTO categories (name)
		VALUES ($1)
		RETURNING id, name
	`

	return scanCategory(r.db.QueryRow(ctx, query, name))
}

func (r *categoryRepo) GetAll(
	ctx context.Context,
) ([]models.Category, error) {
	rows, err := r.db.Query(
		ctx,
		`
			SELECT id, name
			FROM categories
			ORDER BY name ASC
		`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]models.Category, 0)

	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepo) ExistsByID(
	ctx context.Context,
	id string,
) (bool, error) {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)`,
		id,
	).Scan(&exists)

	return exists, err
}

func scanCategory(row pgx.Row) (models.Category, error) {
	var category models.Category

	err := row.Scan(
		&category.ID,
		&category.Name,
	)
	if err != nil {
		return models.Category{}, err
	}

	return category, nil
}

package repo

import (
	"context"
	"strings"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepo interface {
	Create(
		ctx context.Context,
		request models.CreateProductRequest,
	) (models.Product, error)

	GetByID(
		ctx context.Context,
		id string,
	) (models.Product, error)

	List(
		ctx context.Context,
		params models.ProductListParams,
	) ([]models.Product, int64, error)

	Update(
		ctx context.Context,
		id string,
		request models.UpdateProductRequest,
	) (models.Product, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (bool, error)
}

type productRepo struct {
	db *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) ProductRepo {
	return &productRepo{
		db: db,
	}
}

func (r *productRepo) Create(
	ctx context.Context,
	request models.CreateProductRequest,
) (models.Product, error) {
	var productID string

	err := r.db.QueryRow(
		ctx,
		`
			INSERT INTO products (
				category_id,
				name,
				description,
				price,
				stock
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`,
		request.CategoryID,
		request.Name,
		request.Description,
		*request.Price,
		*request.Stock,
	).Scan(&productID)
	if err != nil {
		return models.Product{}, err
	}

	return r.GetByID(ctx, productID)
}

func (r *productRepo) GetByID(
	ctx context.Context,
	id string,
) (models.Product, error) {
	query := `
		SELECT
			p.id,
			COALESCE(p.category_id::text, ''),
			COALESCE(c.name, ''),
			p.name,
			COALESCE(p.description, ''),
			p.price,
			p.stock,
			p.created_at,
			p.updated_at,
			p.deleted_at
		FROM products p
		LEFT JOIN categories c
			ON c.id = p.category_id
		WHERE p.id = $1
			AND p.deleted_at IS NULL
	`

	return scanProduct(r.db.QueryRow(ctx, query, id))
}

func (r *productRepo) List(
	ctx context.Context,
	params models.ProductListParams,
) ([]models.Product, int64, error) {
	whereClause, args := buildProductFilters(params)

	var total int64

	err := r.db.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM products p `+whereClause,
		args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	sortColumn := "p.created_at"
	if params.Sort == "price" {
		sortColumn = "p.price"
	}

	sortOrder := "DESC"
	if params.Order == "asc" {
		sortOrder = "ASC"
	}

	args = append(args, params.Limit, (params.Page-1)*params.Limit)

	query := `
		SELECT
			p.id,
			COALESCE(p.category_id::text, ''),
			COALESCE(c.name, ''),
			p.name,
			COALESCE(p.description, ''),
			p.price,
			p.stock,
			p.created_at,
			p.updated_at,
			p.deleted_at
		FROM products p
		LEFT JOIN categories c
			ON c.id = p.category_id
	` + whereClause + `
		ORDER BY ` + sortColumn + ` ` + sortOrder + `
		LIMIT $` + intToString(len(args)-1) + `
		OFFSET $` + intToString(len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := make([]models.Product, 0)

	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepo) Update(
	ctx context.Context,
	id string,
	request models.UpdateProductRequest,
) (models.Product, error) {
	args := make([]any, 0)
	updates := []string{
		"updated_at = now()",
	}

	if request.CategoryID != nil {
		args = append(args, *request.CategoryID)
		updates = append(
			updates,
			"category_id = $"+intToString(len(args)),
		)
	}

	if request.Name != nil {
		args = append(args, *request.Name)
		updates = append(
			updates,
			"name = $"+intToString(len(args)),
		)
	}

	if request.Description != nil {
		args = append(args, *request.Description)
		updates = append(
			updates,
			"description = $"+intToString(len(args)),
		)
	}

	if request.Price != nil {
		args = append(args, *request.Price)
		updates = append(
			updates,
			"price = $"+intToString(len(args)),
		)
	}

	if request.Stock != nil {
		args = append(args, *request.Stock)
		updates = append(
			updates,
			"stock = $"+intToString(len(args)),
		)
	}

	args = append(args, id)

	var productID string

	err := r.db.QueryRow(
		ctx,
		`
			UPDATE products
			SET `+strings.Join(updates, ", ")+`
			WHERE id = $`+intToString(len(args))+`
				AND deleted_at IS NULL
			RETURNING id
		`,
		args...,
	).Scan(&productID)
	if err != nil {
		return models.Product{}, err
	}

	return r.GetByID(ctx, productID)
}

func (r *productRepo) SoftDelete(
	ctx context.Context,
	id string,
) (bool, error) {
	commandTag, err := r.db.Exec(
		ctx,
		`
			UPDATE products
			SET
				deleted_at = now(),
				updated_at = now()
			WHERE id = $1
				AND deleted_at IS NULL
		`,
		id,
	)
	if err != nil {
		return false, err
	}

	return commandTag.RowsAffected() > 0, nil
}

func buildProductFilters(
	params models.ProductListParams,
) (string, []any) {
	conditions := []string{
		"p.deleted_at IS NULL",
	}
	args := make([]any, 0)

	if params.CategoryID != "" {
		args = append(args, params.CategoryID)
		conditions = append(
			conditions,
			"p.category_id = $"+intToString(len(args)),
		)
	}

	if params.MinPrice != nil {
		args = append(args, *params.MinPrice)
		conditions = append(
			conditions,
			"p.price >= $"+intToString(len(args)),
		)
	}

	if params.MaxPrice != nil {
		args = append(args, *params.MaxPrice)
		conditions = append(
			conditions,
			"p.price <= $"+intToString(len(args)),
		)
	}

	if params.InStock != nil && *params.InStock {
		conditions = append(conditions, "p.stock > 0")
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func scanProduct(row pgx.Row) (models.Product, error) {
	var product models.Product
	var categoryID string

	err := row.Scan(
		&product.ID,
		&categoryID,
		&product.CategoryName,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.DeletedAt,
	)
	if err != nil {
		return models.Product{}, err
	}

	if categoryID != "" {
		product.CategoryID = &categoryID
	}

	return product, nil
}

func intToString(value int) string {
	if value == 0 {
		return "0"
	}

	digits := make([]byte, 0)

	for value > 0 {
		digits = append([]byte{byte(value%10) + '0'}, digits...)
		value /= 10
	}

	return string(digits)
}

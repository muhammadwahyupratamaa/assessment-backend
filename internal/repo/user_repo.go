package repo

import (
	"context"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo interface {
	Create(ctx context.Context, user models.User) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
	GetByID(ctx context.Context, id string) (models.User, error)
}

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) UserRepo {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) Create(
	ctx context.Context,
	user models.User,
) (models.User, error) {
	query := `
		INSERT INTO users (
			name,
			email,
			password_hash
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			name,
			email,
			password_hash,
			role,
			created_at
	`

	return scanUser(
		r.db.QueryRow(
			ctx,
			query,
			user.Name,
			user.Email,
			user.PasswordHash,
		),
	)
}

func (r *userRepo) GetByEmail(
	ctx context.Context,
	email string,
) (models.User, error) {
	query := `
		SELECT
			id,
			name,
			email,
			password_hash,
			role,
			created_at
		FROM users
		WHERE email = $1
	`

	return scanUser(r.db.QueryRow(ctx, query, email))
}

func (r *userRepo) GetByID(
	ctx context.Context,
	id string,
) (models.User, error) {
	query := `
		SELECT
			id,
			name,
			email,
			password_hash,
			role,
			created_at
		FROM users
		WHERE id = $1
	`

	return scanUser(r.db.QueryRow(ctx, query, id))
}

func scanUser(row pgx.Row) (models.User, error) {
	var user models.User

	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

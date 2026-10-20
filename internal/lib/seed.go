package lib

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SeedAdmin(
	ctx context.Context,
	db *pgxpool.Pool,
	email string,
	password string,
) error {
	passwordHash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	commandTag, err := db.Exec(
		ctx,
		`
			INSERT INTO users (
				name,
				email,
				password_hash,
				role
			)
			VALUES ($1, $2, $3, 'admin')
			ON CONFLICT (email) DO NOTHING
		`,
		"Administrator",
		strings.ToLower(strings.TrimSpace(email)),
		passwordHash,
	)
	if err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}

	if commandTag.RowsAffected() > 0 {
		slog.Info("admin account created", "email", email)
	}

	return nil
}

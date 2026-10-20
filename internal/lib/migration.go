package lib

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	assessmentbackend "github.com/muhammadwahyupratamaa/assessment-backend"
)

func RunMigrations(databaseURL string) error {
	sourceDriver, err := iofs.New(
		assessmentbackend.Migrations,
		"migrations",
	)
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	migration, err := migrate.NewWithSourceInstance(
		"iofs",
		sourceDriver,
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("create migration instance: %w", err)
	}

	defer migration.Close()

	if err := migration.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		return fmt.Errorf("run migration: %w", err)
	}

	return nil
}

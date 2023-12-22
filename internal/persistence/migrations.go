package persistence

import (
	"database/sql"
	"fmt"

	"github.com/w-k-s/simple-budget-tracker/log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/w-k-s/simple-budget-tracker/internal/config"
)

func RunMigrations(db *sql.DB, dbConfig config.DBConfig) error {
	driverName := dbConfig.DriverName()
	migrationsDirectory := dbConfig.MigrationDirectory()

	if len(migrationsDirectory) == 0 {
		return fmt.Errorf("invalid migrations directory: '%s'. Must be an absolute path", migrationsDirectory)
	}

	db.SetMaxIdleConns(0) // Required, otherwise pinging will result in EOF
	err := PingWithBackOff(db); 
	if err != nil {
		return fmt.Errorf("failed to ping database. Reason: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{
		DatabaseName: dbConfig.Name(),
		SchemaName:   dbConfig.Schema(),
	}); 
	if err != nil {
		return fmt.Errorf("failed to create instance of psql driver. Reason: %w", err)
	}

	migrations, err := migrate.NewWithDatabaseInstance(migrationsDirectory, driverName, driver); 
	if err != nil {
		return fmt.Errorf("failed to load migrations from %s. Reason: %w", migrationsDirectory, err)
	}

	if err = migrations.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations from %s. Reason: %w", migrationsDirectory, err)
	}

	return nil
}

func MustRunMigrations(db *sql.DB, dbConfig config.DBConfig) {
	if err := RunMigrations(db, dbConfig); err != nil {
		log.Fatalf("Failed to run migrations. Reason: %s", err)
	}
}

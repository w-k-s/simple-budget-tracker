package application

import (
	"database/sql"
	"fmt"
	"log"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var db *sql.DB
var ping backoff.Operation = func() error {
	err := db.Ping()
	if err != nil {
		log.Printf("DB is not ready...backing off...: %s", err)
		return err
	}
	log.Println("DB is ready!")
	return nil
}

func RunMigrations(driverName string, dataSourceName string, migrationsDirectory string) error {
	if len(migrationsDirectory) == 0 {
		return fmt.Errorf("invalid migrations directory: '%s'. Must be an absolute path", migrationsDirectory)
	}

	var (
		driver     database.Driver
		migrations *migrate.Migrate
		err        error
	)

	if db, err = sql.Open(driverName, dataSourceName); err != nil {
		return fmt.Errorf("failed to open connection. Reason: %w", err)
	}

	db.SetMaxIdleConns(0) // Required, otherwise pinging will result in EOF

	_ = backoff.Retry(ping, backoff.NewExponentialBackOff())
	if driver, err = postgres.WithInstance(db, &postgres.Config{
		DatabaseName: "simple_budget_tracker",
		SchemaName:   "budget",
	}); err != nil {
		return fmt.Errorf("failed to create instance of psql driver. Reason: %w", err)
	}

	if migrations, err = migrate.NewWithDatabaseInstance(migrationsDirectory, driverName, driver); err != nil {
		return fmt.Errorf("failed to load migrations from %s. Reason: %w", migrationsDirectory, err)
	}

	if err = migrations.Up(); err != nil {
		return fmt.Errorf("failed to apply migrations from %s. Reason: %w", migrationsDirectory, err)
	}

	return nil
}

func MustRunMigrations(driverName string, dataSourceName string, migrationsDirectory string) {
	if err := RunMigrations(driverName, dataSourceName, migrationsDirectory); err != nil {
		log.Fatalf("Failed to run migrations. Reason: %s", err)
	}
}

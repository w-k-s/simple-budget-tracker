package migrations

import (
	"database/sql"
	"fmt"
	"log"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	u "github.com/w-k-s/simple-budget-tracker/util"
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
		return fmt.Errorf("Invalid migrations directory: '%s'. Must be an absolute path", migrationsDirectory)
	}

	var err error
	db, err = sql.Open(driverName, dataSourceName)
	u.CheckError(err, "Migrations: Failed to connect to database")
	db.SetMaxIdleConns(0) // Required, otherwise pinging will result in EOF

	_ = backoff.Retry(ping, backoff.NewExponentialBackOff())
	driver, err := postgres.WithInstance(db, &postgres.Config{
		DatabaseName: "simple_budget_tracker",
		SchemaName:   "budget",
	})
	u.CheckError(err, "Migrations: Failed to create instance of psql driver")

	migrationsUrl := fmt.Sprintf("file://%s", migrationsDirectory)
	m, err := migrate.NewWithDatabaseInstance(migrationsUrl, "postgres", driver)
	u.CheckError(err, fmt.Sprintf("Failed to load migrations from %s", migrationsUrl))

	err = m.Up()
	u.CheckError(err, fmt.Sprintf("Failed to apply migrations from %s", migrationsUrl))

	return err
}

func MustRunMigrations(driverName string, dataSourceName string, migrationsDirectory string) {
	u.CheckError(RunMigrations(driverName, dataSourceName, migrationsDirectory), "Failed to run migrations")
}

package application

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/w-k-s/simple-budget-tracker/core"
	"github.com/w-k-s/simple-budget-tracker/migrations"
)

const (
	testContainerPostgresUser     = "test"
	testContainerPostgresPassword = "test"
	testContainerPostgresDB       = "simple_budget_tracker"
	testContainerDriverName       = "postgres"
)

var testContainerContext context.Context
var testPostgresContainer tc.Container
var testContainerDataSourceName string

var TestDB *sql.DB
var UserDao core.UserDao
var AccountDao core.AccountDao
var CategoryDao core.CategoryDao
var RecordDao core.RecordDao

func init() {
	testContainerContext = context.Background()
	req := tc.ContainerRequest{
		Image:        "postgres:11.6-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     testContainerPostgresUser,
			"POSTGRES_PASSWORD": testContainerPostgresPassword,
			"POSTGRES_DB":       testContainerPostgresDB,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}
	var err error
	testPostgresContainer, err = tc.GenericContainer(testContainerContext, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to request postgres test container: %s", err)
	}

	containerHost, _ := testPostgresContainer.Host(testContainerContext)
	containerPort, _ := testPostgresContainer.MappedPort(testContainerContext, "5432")
	testContainerDataSourceName = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", containerHost, containerPort.Int(), testContainerPostgresUser, testContainerPostgresPassword, testContainerPostgresDB)

	migrations.MustRunMigrations(testContainerDriverName, testContainerDataSourceName, os.Getenv("TEST_MIGRATIONS_DIRECTORY"))

	if TestDB, err = sql.Open(testContainerDriverName, testContainerDataSourceName); err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", testContainerDriverName, testContainerDataSourceName, err)
	}

	UserDao = MustOpenUserDao(testContainerDriverName, testContainerDataSourceName)
	AccountDao = MustOpenAccountDao(testContainerDriverName, testContainerDataSourceName)
	CategoryDao = MustOpenCategoryDao(testContainerDriverName, testContainerDataSourceName)
	RecordDao = MustOpenRecordDao(testContainerDriverName, testContainerDataSourceName)
}

func TestMain(m *testing.M) {
	defer func(exitCode int) {

		log.Println("Cleaning up after tests")

		if err := UserDao.Close(); err != nil {
			log.Printf("Error closing UserDao: %s", err)
		}

		if err := AccountDao.Close(); err != nil {
			log.Printf("Error closing AccountDao: %s", err)
		}

		if err := CategoryDao.Close(); err != nil {
			log.Printf("Error closing CategoryDao: %s", err)
		}

		if err := RecordDao.Close(); err != nil {
			log.Printf("Error closing RecordDao: %s", err)
		}

		if err := testPostgresContainer.Terminate(testContainerContext); err != nil {
			log.Printf("Error closing Test Postgres Container: %s", err)
		}

		log.Print("Cleanup complete\n\n\n")

		os.Exit(exitCode)
	}(m.Run())
}

func ClearTables() error {
	var db *sql.DB
	var err error

	if db, err = sql.Open(testContainerDriverName, testContainerDataSourceName); err != nil {
		return fmt.Errorf("Failed to connect to %q: %w", testContainerDataSourceName, err)
	}

	if _, err = db.Exec("DELETE FROM budget.user"); err != nil {
		return fmt.Errorf("Failed to delete user table: %w", err)
	}
	if _, err = db.Exec("DELETE FROM budget.account"); err != nil {
		return fmt.Errorf("Failed to delete account table: %w", err)
	}
	if _, err = db.Exec("DELETE FROM budget.category"); err != nil {
		return fmt.Errorf("Failed to delete category table: %w", err)
	}
	if _, err = db.Exec("DELETE FROM budget.record"); err != nil {
		return fmt.Errorf("Failed to delete record table: %w", err)
	}
	return nil
}

func quickMoney(currency string, amountMinorUnits int64) core.Money {
	amount, err := core.NewMoney(currency, amountMinorUnits)
	if err != nil {
		log.Fatalf("Failed to create money: %s", err)
	}
	return amount
}

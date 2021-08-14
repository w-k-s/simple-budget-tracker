package test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	cfg "github.com/w-k-s/simple-budget-tracker/internal/config"
	app "github.com/w-k-s/simple-budget-tracker/internal/server"
	db "github.com/w-k-s/simple-budget-tracker/internal/server/persistence"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
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
var UserDao dao.UserDao
var AccountDao dao.AccountDao
var CategoryDao dao.CategoryDao
var RecordDao dao.RecordDao
var TestConfig *cfg.Config
var TestApp *app.App

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

	if TestConfig, _ = cfg.NewConfig(
		cfg.NewServerConfigBuilder().
			SetPort(9898).
			Build(),
		cfg.NewDBConfigBuilder().
			SetUsername(testContainerPostgresUser).
			SetPassword(testContainerPostgresPassword).
			SetHost(containerHost).
			SetPort(containerPort.Int()).
			SetName(testContainerDataSourceName).
			Build(),
	); err != nil {
		log.Fatalf("Failed to configure application for tests. Reason: %s", err)
	}

	db.MustRunMigrations(TestConfig.Database())

	if TestDB, err = sql.Open(testContainerDriverName, testContainerDataSourceName); err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", testContainerDriverName, testContainerDataSourceName, err)
	}

	UserDao = db.MustOpenUserDao(testContainerDriverName, testContainerDataSourceName)
	AccountDao = db.MustOpenAccountDao(testContainerDriverName, testContainerDataSourceName)
	CategoryDao = db.MustOpenCategoryDao(testContainerDriverName, testContainerDataSourceName)
	RecordDao = db.MustOpenRecordDao(testContainerDriverName, testContainerDataSourceName)

	if TestApp, err = app.Init(TestConfig); err != nil {
		log.Fatalf("Failed to initialize application for tests. Reason: %s", err)
	}
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

func quickMoney(currency string, amountMinorUnits int64) ledger.Money {
	amount, err := ledger.NewMoney(currency, amountMinorUnits)
	if err != nil {
		log.Fatalf("Failed to create money: %s", err)
	}
	return amount
}

func AddAuthorizationHeader(r *http.Request, userId ledger.UserId){
	r.Header.Add("Authorization", fmt.Sprintf("%d", userId))
}

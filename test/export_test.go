package test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	cfg "github.com/w-k-s/simple-budget-tracker/internal/config"
	db "github.com/w-k-s/simple-budget-tracker/internal/persistence"
	app "github.com/w-k-s/simple-budget-tracker/internal/server"
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
		*cfg.NewGptConfig(""),
	); err != nil {
		log.Fatalf("Failed to configure application for tests. Reason: %s", err)
	}

	if TestDB, err = sql.Open(testContainerDriverName, testContainerDataSourceName); err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", testContainerDriverName, testContainerDataSourceName, err)
	}

	UserDao = db.MustOpenUserDao(TestDB)
	AccountDao = db.MustOpenAccountDao(TestDB)
	CategoryDao = db.MustOpenCategoryDao(TestDB)
	RecordDao = db.MustOpenRecordDao(TestDB)

	if TestApp, err = app.Init(TestConfig); err != nil {
		log.Fatalf("Failed to initialize application for tests. Reason: %s", err)
	}
}

func TestMain(m *testing.M) {
	defer func(exitCode int) {

		log.Println("Cleaning up after tests")

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
	if _, err = db.Exec("DELETE FROM budget.record"); err != nil {
		return fmt.Errorf("Failed to delete record table: %w", err)
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

	if _, err = db.Exec("ALTER SEQUENCE budget.user_id RESTART"); err != nil {
		return fmt.Errorf("Failed to delete record table: %w", err)
	}
	if _, err = db.Exec("ALTER SEQUENCE budget.account_id RESTART"); err != nil {
		return fmt.Errorf("Failed to delete record table: %w", err)
	}
	if _, err = db.Exec("ALTER SEQUENCE budget.category_id RESTART"); err != nil {
		return fmt.Errorf("Failed to delete record table: %w", err)
	}
	if _, err = db.Exec("ALTER SEQUENCE budget.record_id RESTART"); err != nil {
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

func AddAuthorizationHeader(r *http.Request, userId ledger.UserId) {
	r.Header.Add("Authorization", fmt.Sprintf("%d", userId))
}

func errorCode(err error, defaultValue uint64) uint64 {
	if errWithCode, ok := err.(interface {
		Code() uint64
	}); ok {
		return errWithCode.Code()
	}
	return defaultValue
}

type UserAndAccounts map[ledger.UserId][]ledger.AccountId

func (u UserAndAccounts) First() ledger.UserId {
	for userId := range u {
		return userId
	}
	return ledger.UserId(0)
}

func simulateRecords(db *sql.DB, numberOfUsers int, startMonth ledger.CalendarMonth, endMonth ledger.CalendarMonth) (UserAndAccounts, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	result := map[ledger.UserId][]ledger.AccountId{}

	for i := 0; i < numberOfUsers; i++ {
		// create user
		var user ledger.User
		userId := ledger.UserId(time.Now().UnixNano())
		if user, err = ledger.NewUserWithEmailString(userId, fmt.Sprintf("testUser+%d@gmail.com", userId)); err != nil {
			return nil, err
		}
		if err = UserDao.SaveTx(user, tx); err != nil {
			return nil, err
		}

		// create account (2 accounts per user)
		var (
			currentAccount ledger.Account
			savingsAccount ledger.Account
		)

		if currentAccount, err = ledger.NewAccount(ledger.AccountId(time.Now().UnixNano()), "Current", ledger.AccountTypeCurrent, "AED", ledger.MustMakeUpdatedByUserId(userId)); err != nil {
			return nil, err
		}
		if savingsAccount, err = ledger.NewAccount(ledger.AccountId(time.Now().UnixNano()), "Savings", ledger.AccountTypeSaving, "AED", ledger.MustMakeUpdatedByUserId(userId)); err != nil {
			return nil, err
		}

		if err = AccountDao.SaveTx(context.Background(), userId, ledger.Accounts{currentAccount, savingsAccount}, tx); err != nil {
			return nil, err
		}

		// create categories
		var (
			salaryCategory       ledger.Category
			savingsCaegory       ledger.Category
			billsCategory        ledger.Category
			foodAndDrinkCategory ledger.Category
		)
		if salaryCategory, err = ledger.NewCategory(ledger.CategoryId(time.Now().UnixNano()), "Salary", ledger.MustMakeUpdatedByUserId(userId)); err != nil {
			return nil, err
		}
		if savingsCaegory, err = ledger.NewCategory(ledger.CategoryId(time.Now().UnixNano()), "Savings", ledger.MustMakeUpdatedByUserId(userId)); err != nil {
			return nil, err
		}
		if billsCategory, err = ledger.NewCategory(ledger.CategoryId(time.Now().UnixNano()), "Bills", ledger.MustMakeUpdatedByUserId(userId)); err != nil {
			return nil, err
		}
		if foodAndDrinkCategory, err = ledger.NewCategory(ledger.CategoryId(time.Now().UnixNano()), "Food & Drink", ledger.MustMakeUpdatedByUserId(userId)); err != nil {
			return nil, err
		}

		categories := ledger.Categories{salaryCategory, savingsCaegory, billsCategory, foodAndDrinkCategory}
		if err = CategoryDao.SaveTx(context.Background(), userId, categories, tx); err != nil {
			return nil, err
		}

		// create records
		fromDate := startMonth.FirstDay()
		toDate := endMonth.LastDay()

		for date := fromDate; date.Before(toDate); date = date.AddDate(0, 0, 1) {
			if date.Day() == 1 {
				var (
					salaryRecord           ledger.Record
					billsRecord            ledger.Record
					sendToSavingsRecord    ledger.Record
					receiveInSavingsRecord ledger.Record
				)
				// Add salary
				if salaryRecord, err = ledger.NewRecord(
					ledger.RecordId(time.Now().UnixNano()),
					"Salary",
					salaryCategory,
					quickMoney("AED", 50_000_00),
					date,
					ledger.Income,
					ledger.NoSourceAccount,
					ledger.NoBeneficiaryAccount,
					ledger.NoBeneficiaryType,
					ledger.NoTransferReference,
					ledger.MustMakeUpdatedByUserId(userId),
				); err != nil {
					return nil, err
				}
				// Pay bills
				if billsRecord, err = ledger.NewRecord(
					ledger.RecordId(time.Now().UnixNano()),
					"Bills",
					billsCategory,
					quickMoney("AED", 1_000_00),
					date,
					ledger.Expense,
					ledger.NoSourceAccount,
					ledger.NoBeneficiaryAccount,
					ledger.NoBeneficiaryType,
					ledger.NoTransferReference,
					ledger.MustMakeUpdatedByUserId(userId),
				); err != nil {
					return nil, err
				}
				// Save a little money for a rainy day
				ref := ledger.MakeTransferReference()
				if sendToSavingsRecord, err = ledger.NewRecord(
					ledger.RecordId(
						time.Now().UnixNano()),
					"Savings",
					savingsCaegory,
					quickMoney("AED", -10_000_00),
					date,
					ledger.Transfer,
					currentAccount.Id(),
					savingsAccount.Id(),
					savingsAccount.Type(),
					ref,
					ledger.MustMakeUpdatedByUserId(userId),
				); err != nil {
					return nil, err
				}
				if receiveInSavingsRecord, err = ledger.NewRecord(
					ledger.RecordId(time.Now().UnixNano()),
					"Savings",
					savingsCaegory,
					quickMoney("AED", 10_000_00),
					date,
					ledger.Transfer,
					currentAccount.Id(),
					savingsAccount.Id(),
					savingsAccount.Type(),
					ref,
					ledger.MustMakeUpdatedByUserId(userId),
				); err != nil {
					return nil, err
				}

				if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), salaryRecord, tx); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), billsRecord, tx); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), sendToSavingsRecord, tx); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), receiveInSavingsRecord, tx); err != nil {
					return nil, err
				}
			}

			// buy lunch and dinner every day
			var (
				lunchRecord  ledger.Record
				dinnerRecord ledger.Record
			)
			if lunchRecord, err = ledger.NewRecord(
				ledger.RecordId(time.Now().UnixNano()),
				"Lunch",
				foodAndDrinkCategory,
				quickMoney("AED", 10_00),
				date,
				ledger.Expense,
				ledger.NoSourceAccount,
				ledger.NoBeneficiaryAccount,
				ledger.NoBeneficiaryType,
				ledger.NoTransferReference,
				ledger.MustMakeUpdatedByUserId(userId),
			); err != nil {
				return nil, err
			}
			if dinnerRecord, err = ledger.NewRecord(
				ledger.RecordId(time.Now().UnixNano()),
				"Dinner",
				foodAndDrinkCategory,
				quickMoney("AED", 20_00),
				date,
				ledger.Expense,
				ledger.NoSourceAccount,
				ledger.NoBeneficiaryAccount,
				ledger.NoBeneficiaryType,
				ledger.NoTransferReference,
				ledger.MustMakeUpdatedByUserId(userId),
			); err != nil {
				return nil, err
			}
			if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), lunchRecord, tx); err != nil {
				return nil, err
			}
			if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), dinnerRecord, tx); err != nil {
				return nil, err
			}

			// eat big lunch on birthday (e.g. June 2)
			if date.Day() == 2 && date.Month() == time.June {
				var specialOccasionRecord ledger.Record
				if specialOccasionRecord, err = ledger.NewRecord(
					ledger.RecordId(time.Now().UnixNano()),
					"Birthday",
					foodAndDrinkCategory,
					quickMoney("AED", 100_00),
					date,
					ledger.Expense,
					ledger.NoSourceAccount,
					ledger.NoBeneficiaryAccount,
					ledger.NoBeneficiaryType,
					ledger.NoTransferReference,
					ledger.MustMakeUpdatedByUserId(userId),
				); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), specialOccasionRecord, tx); err != nil {
					return nil, err
				}
			}
		}

		if err = tx.Commit(); err != nil {
			return nil, err
		}

		var count int
		_ = db.QueryRow("SELECT COUNT(*) FROM budget.record").Scan(&count)
		log.Printf("%d records inserted for simulation", count)

		result[userId] = []ledger.AccountId{currentAccount.Id(), savingsAccount.Id()}
	}
	return result, nil
}

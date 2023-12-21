package persistence

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
)

type Dao interface {
	BeginTx() (*sql.Tx, error)
	MustBeginTx() *sql.Tx
}

type UserDao interface {
	BeginTx() (*sql.Tx, error)
	MustBeginTx() *sql.Tx

	NewUserId() (ledger.UserId, error)

	Save(u ledger.User) error
	SaveTx(u ledger.User, tx *sql.Tx) error

	GetUserById(id ledger.UserId) (ledger.User, error)
}

type AccountDao interface {
	BeginTx() (*sql.Tx, error)
	MustBeginTx() *sql.Tx

	NewAccountId(tx *sql.Tx) (ledger.AccountId, error)

	SaveTx(ctx context.Context, id ledger.UserId, as ledger.Accounts, tx *sql.Tx) error

	GetAccountsByUserId(ctx context.Context, id ledger.UserId, tx *sql.Tx) (ledger.Accounts, error)
	GetAccountById(ctx context.Context, id ledger.AccountId, userId ledger.UserId, tx *sql.Tx) (ledger.Account, error)
}

type CategoryDao interface {
	BeginTx() (*sql.Tx, error)
	MustBeginTx() *sql.Tx

	NewCategoryId(tx *sql.Tx) (ledger.CategoryId, error)

	SaveTx(ctx context.Context, id ledger.UserId, c ledger.Categories, tx *sql.Tx) error

	GetCategoryById(ctx context.Context, id ledger.CategoryId, userId ledger.UserId, tx *sql.Tx) (ledger.Category, error)
	GetCategoriesForUser(ctx context.Context, id ledger.UserId, tx *sql.Tx) (ledger.Categories, error)

	UpdateCategoryLastUsed(ctx context.Context, id ledger.CategoryId, lastUsed time.Time, tx *sql.Tx) error
}

type RecordDao interface {
	BeginTx() (*sql.Tx, error)
	MustBeginTx() *sql.Tx

	NewRecordId(*sql.Tx) (ledger.RecordId, error)

	SaveTx(ctx context.Context, id ledger.AccountId, r ledger.Record, tx *sql.Tx) error

	Search(id ledger.AccountId, search RecordSearch) (ledger.Records, error)
	GetRecordsForMonth(id ledger.AccountId, month ledger.CalendarMonth) (ledger.Records, error)
	GetRecordsForLastPeriod(ctx context.Context, id ledger.AccountId, tx *sql.Tx) (ledger.Records, error)
}

type RecordSearch struct {
	SearchTerm              string
	FromDate                *time.Time
	ToDate                  *time.Time
	CategoryNames           []string
	RecordTypes             []ledger.RecordType
	BeneficiaryAccountNames []string
}

func DeferRollback(tx *sql.Tx, reference string) {
	if tx == nil {
		return
	}
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		log.Printf("failed to rollback transaction with reference %q. Reason: %s", reference, err)
	}
}

func Commit(tx *sql.Tx) error {
	if tx == nil {
		log.Fatal("Commit should not be passed a nil transaction")
	}
	if err := tx.Commit(); err != nil && err != sql.ErrTxDone {
		return ledger.NewError(ledger.ErrDatabaseConnectivity, "Failed to save changes", err)
	}
	return nil
}

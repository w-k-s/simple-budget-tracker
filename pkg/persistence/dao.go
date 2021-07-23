package persistence

import (
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

	Close() error
	NewUserId() (ledger.UserId, error)

	Save(u ledger.User) error
	SaveTx(u ledger.User, tx *sql.Tx) error

	GetUserById(id ledger.UserId) (ledger.User, error)
}

type AccountDao interface {
	Close() error
	NewAccountId() (ledger.AccountId, error)

	Save(id ledger.UserId, a *ledger.Account) error
	SaveTx(id ledger.UserId, a *ledger.Account, tx *sql.Tx) error

	GetAccountById(id ledger.AccountId) (*ledger.Account, error)
	GetAccountsByUserId(id ledger.UserId) ([]*ledger.Account, error)
}

type CategoryDao interface {
	Close() error
	NewCategoryId() (ledger.CategoryId, error)

	Save(id ledger.UserId, c ledger.Categories) error
	SaveTx(id ledger.UserId, c ledger.Categories, tx *sql.Tx) error

	GetCategoriesForUser(id ledger.UserId) (ledger.Categories, error)
}

type RecordDao interface {
	Close() error
	NewRecordId() (ledger.RecordId, error)

	Save(id ledger.AccountId, r *ledger.Record) error
	SaveTx(id ledger.AccountId, r *ledger.Record, tx *sql.Tx) error

	Search(id ledger.AccountId, search RecordSearch) (ledger.Records, error)
	GetRecordsForMonth(id ledger.AccountId, month ledger.CalendarMonth) (ledger.Records, error)
	GetRecordsForLastPeriod(id ledger.AccountId) (ledger.Records, error)
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
	if err := tx.Rollback(); err != nil {
		log.Printf("failed to rollback transaction with reference %q. Reason: %s", reference, err)
	}
}

func Commit(tx *sql.Tx) error {
	if tx == nil {
		log.Fatal("Commit should not be passed a nil transaction")
	}
	if err := tx.Commit(); err != nil {
		return ledger.NewError(ledger.ErrDatabaseConnectivity, "Failed to save changes", err)
	}
	return nil
}

package core

import (
	"database/sql"
	"log"
	"time"
)

type Dao interface {
	BeginTx() (*sql.Tx, error)
	MustBeginTx() *sql.Tx
}

type UserDao interface {
	BeginTx() (*sql.Tx, error)
	MustBeginTx() *sql.Tx

	Close() error
	NewUserId() (UserId, error)

	Save(u *User) error
	SaveTx(u *User, tx *sql.Tx) error

	GetUserById(id UserId) (*User, error)
}

type AccountDao interface {
	Close() error
	NewAccountId() (AccountId, error)

	Save(id UserId, a *Account) error
	SaveTx(id UserId, a *Account, tx *sql.Tx) error

	GetAccountById(id AccountId) (*Account, error)
	GetAccountsByUserId(id UserId) ([]*Account, error)
}

type CategoryDao interface {
	Close() error
	NewCategoryId() (CategoryId, error)

	Save(id UserId, c Categories) error
	SaveTx(id UserId, c Categories, tx *sql.Tx) error

	GetCategoriesForUser(id UserId) (Categories, error)
}

type RecordDao interface {
	Close() error
	NewRecordId() (RecordId, error)

	Save(id AccountId, r *Record) error
	SaveTx(id AccountId, r *Record, tx *sql.Tx) error

	Search(id AccountId, search RecordSearch) (Records, error)
	GetRecordsForMonth(id AccountId, month int, year int) (Records, error)
	GetRecordsForLastPeriod(id AccountId) (Records, error)
}

type RecordSearch struct {
	SearchTerm              string
	FromDate                *time.Time
	ToDate                  *time.Time
	CategoryNames           []string
	RecordTypes             []RecordType
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
		return NewError(ErrDatabaseConnectivity, "Failed to save changes", err)
	}
	return nil
}

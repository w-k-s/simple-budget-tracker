package persistence

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type DefaultUserDao struct {
	*RootDao
}

func MustOpenUserDao(driverName, dataSourceName string) dao.UserDao {
	var db *sql.DB
	var err error
	if db, err = sql.Open(driverName, dataSourceName); err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", dataSourceName, driverName, err)
	}
	return &DefaultUserDao{
		&RootDao{db},
	}
}

func (d DefaultUserDao) Close() error {
	return d.db.Close()
}

func (d *DefaultUserDao) NewUserId() (ledger.UserId, error) {
	var userId ledger.UserId
	err := d.db.QueryRow("SELECT nextval('budget.user_id')").Scan(&userId)
	if err != nil {
		log.Printf("Failed to assign user id. Reason; %s", err)
		return 0, ledger.NewError(ledger.ErrDatabaseState, "Failed to assign user id", err)
	}
	return userId, err
}

func (d *DefaultUserDao) SaveTx(u ledger.User, tx *sql.Tx) error {
	_, err := tx.Exec("INSERT INTO budget.user (id, email) VALUES ($1, $2)", u.Id(), u.Email().Address)
	if err != nil {
		log.Printf("Failed to save user %v. Reason: %s", u, err)
		if message, ok := d.isDuplicateKeyError(err); ok {
			return ledger.NewError(ledger.ErrUserEmailDuplicated, message, err)
		}
		return ledger.NewError(ledger.ErrDatabaseState, "Failed to save user", err)
	}
	return nil
}

func (d *DefaultUserDao) GetUserById(queryId ledger.UserId) (ledger.User, error) {
	var userId ledger.UserId
	var email string

	err := d.db.QueryRow("SELECT id, email FROM budget.user WHERE id = $1", queryId).Scan(&userId, &email)
	if err != nil {
		if err == sql.ErrNoRows {
			return ledger.User{}, ledger.NewError(ledger.ErrUserNotFound, fmt.Sprintf("User with id %d not found", queryId), err)
		}
		return ledger.User{}, ledger.NewError(ledger.ErrDatabaseState, fmt.Sprintf("User with id %d not found", queryId), err)
	}

	return ledger.NewUserWithEmailString(userId, email)
}

func (d *DefaultUserDao) Save(u ledger.User) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	err = d.SaveTx(u, tx)
	if err == nil {
		err = tx.Commit()
	}
	return err
}

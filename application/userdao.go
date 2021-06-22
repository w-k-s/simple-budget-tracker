package application

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/w-k-s/simple-budget-tracker/core"
)

type DefaultUserDao struct {
	db *sql.DB
}

func MustOpenUserDao(driverName, dataSourceName string) *DefaultUserDao {
	var db *sql.DB
	var err error
	if db, err = sql.Open(driverName, dataSourceName); err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", dataSourceName, driverName, err)
	}
	return &DefaultUserDao{db}
}

func (d DefaultUserDao) Close() error {
	return d.db.Close()
}

func (d *DefaultUserDao) NewUserId() (core.UserId, error) {
	var userId core.UserId
	err := d.db.QueryRow("SELECT nextval('budget.user_id')").Scan(&userId)
	if err != nil {
		log.Printf("Failed to assign user id. Reason; %s", err)
		return 0, core.NewError(core.ErrDatabaseState, "Failed to assign user id", err)
	}
	return userId, err
}

func (d *DefaultUserDao) SaveTx(u *core.User, tx *sql.Tx) error {
	_, err := tx.Exec("INSERT INTO budget.user (id, email) VALUES ($1, $2)", u.Id(), u.Email().Address)
	if err != nil {
		log.Printf("Failed to save user %v. Reason: %s", u, err)
		if message, ok := isDuplicateKeyError(err); ok {
			return core.NewError(core.ErrUserEmailDuplicated, message, err)
		}
		return core.NewError(core.ErrDatabaseState, "Failed to save user", err)
	}
	return nil
}

func (d *DefaultUserDao) GetUserById(queryId core.UserId) (*core.User, error) {
	var userId core.UserId
	var email string

	err := d.db.QueryRow("SELECT id, email FROM budget.user WHERE id = $1", queryId).Scan(&userId, &email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.NewError(core.ErrUserNotFound, fmt.Sprintf("User with id %d not found", queryId), err)
		}
		return nil, core.NewError(core.ErrDatabaseState, fmt.Sprintf("User with id %d not found", queryId), err)
	}

	return core.NewUserWithEmailString(userId, email)
}

func (d *DefaultUserDao) Save(u *core.User) error {
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

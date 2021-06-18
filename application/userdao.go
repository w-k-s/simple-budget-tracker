package application

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/w-k-s/simple-budget-tracker/core"
)

type DefaultUserDao struct {
	db *sqlx.DB
}

func OpenUserDao(driverName, dataSourceName string) (*DefaultUserDao, error) {
	var err error
	if db, err := sqlx.Connect(driverName, dataSourceName); err == nil {
		return &DefaultUserDao{db}, nil
	}
	return nil, err
}

func MustOpenUserDao(driverName, dataSourceName string) *DefaultUserDao {
	db, err := OpenUserDao(driverName, dataSourceName)
	if err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %q", dataSourceName, driverName, err)
	}
	return db
}

func (d DefaultUserDao) Close() error{
	return d.db.Close()
}

func (d *DefaultUserDao) NewUserId() (core.UserId, error){
	var userId core.UserId
	err := d.db.QueryRow("SELECT nextval('budget.user_id')").Scan(&userId)
	return userId, err
}

func (d *DefaultUserDao) Save(u *core.User) {
	tx := d.db.MustBegin()
	tx.MustExec("INSERT INTO budget.user (id, email) VALUES ($1, $2)", u.Id(), u.Email().Address)
	tx.Commit()
}

func (d *DefaultUserDao) GetUserById(queryId core.UserId) (*core.User, error){
    var userId core.UserId
	var email string

	err := d.db.QueryRow("SELECT id, email FROM budget.user WHERE id = $1", queryId).Scan(&userId, &email)
	if err != nil {
		return nil, fmt.Errorf("user with id %d could not be found: %w", queryId, err)
	}

	return core.NewUserWithEmailString(userId, email)
}
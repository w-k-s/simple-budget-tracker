package application

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/w-k-s/simple-budget-tracker/core"
)

type DefaultUserDao struct {
	db *sql.DB
}

func OpenUserDao(driverName, dataSourceName string) (*DefaultUserDao, error) {
	var err error
	if db, err := sql.Open(driverName, dataSourceName); err == nil {
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

func (d *DefaultUserDao) SaveTx(u *core.User, tx *sql.Tx) error {
	_,err := tx.Exec("INSERT INTO budget.user (id, email) VALUES ($1, $2)", u.Id(), u.Email().Address)
	if err != nil{
		return fmt.Errorf("failed to save user: %w", err) 
	}
	return nil
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

func (d *DefaultUserDao) Save(u *core.User) error {
	tx,err := d.db.Begin()
	if err != nil{
		return err
	}
	defer tx.Rollback()
	
	err = d.SaveTx(u, tx)
	if err == nil{
		err = tx.Commit()
	}
	return err
}

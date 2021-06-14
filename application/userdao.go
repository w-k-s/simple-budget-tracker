package application

import (
	"github.com/jmoiron/sqlx"
	"github.com/w-k-s/simple-budget-tracker/core"
	_ "github.com/lib/pq"
	"log"
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

func (d *DefaultUserDao) Save(u *core.User) {
	tx := d.db.MustBegin()
	tx.MustExec("INSERT INTO budget.user (id, email) VALUES ($1, $2)", u.Id(), u.Email().Address)
	tx.Commit()
}
